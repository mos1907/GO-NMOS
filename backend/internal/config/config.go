package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	JWTSecret        string
	InitAdmin        string
	InitPassword     string
	CORSOrigin       string
	LogDir           string
	RateLimitRPM     int
	MQTTEnabled      bool
	MQTTBrokerURL    string
	MQTTTopicPrefix  string
	MQTTWSPort       string
	DisableAuth      bool
	HTTPSEnabled     bool
	HTTPSPort        string
	CertFile         string
	KeyFile          string
	SDNControllerURL string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Port:             getenv("APP_PORT", "8080"),
		DatabaseURL:      getenv("DATABASE_URL", "postgres://postgres:postgres@db:5432/go_nmos?sslmode=disable"),
		JWTSecret:        getenv("JWT_SECRET", "change-me-in-production"),
		InitAdmin:        getenv("INIT_ADMIN_USER", "admin"),
		InitPassword:     getenv("INIT_ADMIN_PASSWORD", "admin"),
		CORSOrigin:       getenv("CORS_ORIGIN", "*"),
		LogDir:           getenv("LOG_DIR", "/tmp/go-nmos-logs"),
		RateLimitRPM:     getenvInt("RATE_LIMIT_RPM", 600),
		MQTTEnabled:      getenvBool("MQTT_ENABLED", true),
		MQTTBrokerURL:    getenv("MQTT_BROKER_URL", "tcp://mqtt:1883"),
		MQTTTopicPrefix:  getenv("MQTT_TOPIC_PREFIX", "go-nmos/flows/events"),
		MQTTWSPort:       getenv("MQTT_WS_PORT", "9001"),
		DisableAuth:      getenvBool("DISABLE_AUTH", false),
		HTTPSEnabled:     getenvBool("HTTPS_ENABLED", false),
		HTTPSPort:        getenv("HTTPS_PORT", "8443"),
		CertFile:         getenv("CERT_FILE", "/certs/server.crt"),
		KeyFile:          getenv("KEY_FILE", "/certs/server.key"),
		SDNControllerURL: getenv("SDN_CONTROLLER_URL", ""),
	}

	if cfg.JWTSecret == "change-me-in-production" {
		log.Println("warning: default JWT_SECRET is in use; change this before production")
	}

	return cfg
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}
