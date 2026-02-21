package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-nmos/backend/internal/alerting"
	"go-nmos/backend/internal/config"
	"go-nmos/backend/internal/db"
	"go-nmos/backend/internal/http/handlers"
	"go-nmos/backend/internal/mqtt"
	"go-nmos/backend/internal/repository"
	"go-nmos/backend/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	if err := ensureAdmin(ctx, pool, cfg); err != nil {
		log.Fatalf("admin bootstrap failed: %v", err)
	}

	repo := repository.NewPostgresRepository(pool)

	var mqttClient *mqtt.Client
	if cfg.MQTTEnabled {
		var err error
		mqttClient, err = mqtt.NewClient(cfg.MQTTBrokerURL, cfg.MQTTTopicPrefix, true)
		if err != nil {
			log.Printf("warning: MQTT client initialization failed: %v (continuing without MQTT)", err)
		} else {
			defer mqttClient.Close()
		}
	} else {
		mqttClient, _ = mqtt.NewClient("", "", false)
	}

	h := handlers.NewHandler(cfg, repo, mqttClient)
	r := h.Router()

	runnerCtx, runnerCancel := context.WithCancel(context.Background())
	defer runnerCancel()
	automationRunner := service.NewAutomationRunner(repo)
	go automationRunner.Start(runnerCtx)
	scheduledRunner := service.NewScheduledActivationsRunner(repo)
	go scheduledRunner.Start(runnerCtx)

	// E.2: Scheduled playbook executions runner
	scheduledPlaybooksRunner := service.NewScheduledPlaybooksRunner(repo)
	go scheduledPlaybooksRunner.Start(runnerCtx)

	// F.3: Alert monitoring
	alertHooks := []alerting.AlertHook{
		alerting.NewLoggerHook(), // Always log alerts
	}
	// Add webhook hook if configured
	if webhookURL := os.Getenv("ALERT_WEBHOOK_URL"); webhookURL != "" {
		alertHooks = append(alertHooks, alerting.NewWebhookHook(webhookURL, nil))
	}
	// Add Slack hook if configured
	if slackWebhookURL := os.Getenv("ALERT_SLACK_WEBHOOK_URL"); slackWebhookURL != "" {
		alertHooks = append(alertHooks, alerting.NewSlackHook(slackWebhookURL))
	}
	alertManager := alerting.NewAlertManager(alertHooks...)
	alertMonitor := service.NewAlertMonitor(repo, alertManager)
	go alertMonitor.Start(runnerCtx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// HTTP server (always started)
	httpSrv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("go-NMOS backend HTTP server started on :%s", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// HTTPS server (if enabled)
	var httpsSrv *http.Server
	if cfg.HTTPSEnabled {
		// Load certificate
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			log.Fatalf("failed to load certificate: %v (check CERT_FILE and KEY_FILE)", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}

		httpsSrv = &http.Server{
			Addr:              ":" + cfg.HTTPSPort,
			Handler:           r,
			TLSConfig:         tlsConfig,
			ReadHeaderTimeout: 10 * time.Second,
		}

		go func() {
			log.Printf("go-NMOS backend HTTPS server started on :%s", cfg.HTTPSPort)
			if err := httpsSrv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS server error: %v", err)
			}
		}()
	}

	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	log.Println("shutting down servers...")
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
	}
	if httpsSrv != nil {
		if err := httpsSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTPS server shutdown failed: %v", err)
		}
	}
}

func ensureAdmin(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.InitPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO users(username, password_hash, role)
		VALUES ($1, $2, 'admin')
		ON CONFLICT (username) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
		    role = 'admin'
	`, cfg.InitAdmin, string(hash))
	return err
}
