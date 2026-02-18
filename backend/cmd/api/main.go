package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	runner := service.NewAutomationRunner(repo)
	go runner.Start(runnerCtx)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("go-NMOS backend started on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
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
