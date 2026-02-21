package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Alert represents a system alert/notification
type Alert struct {
	Severity    string            `json:"severity"`    // "info", "warning", "error", "critical"
	Title       string            `json:"title"`
	Message     string            `json:"message"`
	Component   string            `json:"component"`   // "registry", "ptp", "connection", etc.
	Condition   string            `json:"condition"`   // "registry_empty", "ptp_mismatch", etc.
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

// AlertHook defines the interface for alert delivery mechanisms
type AlertHook interface {
	Send(alert Alert) error
	Name() string
}

// WebhookHook sends alerts to a webhook URL
type WebhookHook struct {
	URL     string
	Headers map[string]string
	Client  *http.Client
}

func NewWebhookHook(url string, headers map[string]string) *WebhookHook {
	return &WebhookHook{
		URL:     url,
		Headers: headers,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *WebhookHook) Name() string {
	return "webhook"
}

func (h *WebhookHook) Send(alert Alert) error {
	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	req, err := http.NewRequest("POST", h.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SlackHook sends alerts to Slack via webhook URL
type SlackHook struct {
	WebhookURL string
	Client     *http.Client
}

func NewSlackHook(webhookURL string) *SlackHook {
	return &SlackHook{
		WebhookURL: webhookURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *SlackHook) Name() string {
	return "slack"
}

func (h *SlackHook) Send(alert Alert) error {
	severityEmoji := map[string]string{
		"info":     "ℹ️",
		"warning":  "⚠️",
		"error":    "❌",
		"critical": "🚨",
	}

	color := map[string]string{
		"info":     "#36a64f",
		"warning":  "#ff9900",
		"error":    "#ff0000",
		"critical": "#8b0000",
	}

	emoji := severityEmoji[alert.Severity]
	if emoji == "" {
		emoji = "ℹ️"
	}

	slackPayload := map[string]any{
		"text": fmt.Sprintf("%s %s", emoji, alert.Title),
		"attachments": []map[string]any{
			{
				"color":     color[alert.Severity],
				"title":     alert.Title,
				"text":      alert.Message,
				"fields": []map[string]any{
					{"title": "Component", "value": alert.Component, "short": true},
					{"title": "Condition", "value": alert.Condition, "short": true},
					{"title": "Severity", "value": alert.Severity, "short": true},
					{"title": "Timestamp", "value": alert.Timestamp.Format(time.RFC3339), "short": true},
				},
			},
		},
	}

	payload, err := json.Marshal(slackPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequest("POST", h.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.Client.Do(req)
	if err != nil {
		return fmt.Errorf("slack webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// EmailHook is a placeholder for email alerts (not implemented)
type EmailHook struct {
	SMTPHost     string
	SMTPPort     int
	From         string
	To           []string
	AuthUsername string
	AuthPassword string
}

func NewEmailHook(smtpHost string, smtpPort int, from string, to []string, username, password string) *EmailHook {
	return &EmailHook{
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		From:         from,
		To:           to,
		AuthUsername: username,
		AuthPassword: password,
	}
}

func (h *EmailHook) Name() string {
	return "email"
}

func (h *EmailHook) Send(alert Alert) error {
	// Placeholder: Email sending not implemented
	// In production, use a library like net/smtp or gomail
	log.Printf("[EMAIL PLACEHOLDER] Alert: %s - %s (To: %v)", alert.Severity, alert.Title, h.To)
	return nil
}

// LoggerHook logs alerts to the application log
type LoggerHook struct{}

func NewLoggerHook() *LoggerHook {
	return &LoggerHook{}
}

func (h *LoggerHook) Name() string {
	return "logger"
}

func (h *LoggerHook) Send(alert Alert) error {
	log.Printf("[ALERT] %s [%s] %s: %s", alert.Severity, alert.Component, alert.Title, alert.Message)
	return nil
}

// AlertManager manages multiple alert hooks
type AlertManager struct {
	hooks []AlertHook
}

func NewAlertManager(hooks ...AlertHook) *AlertManager {
	return &AlertManager{
		hooks: hooks,
	}
}

// Send sends an alert to all configured hooks
func (m *AlertManager) Send(alert Alert) {
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now().UTC()
	}

	for _, hook := range m.hooks {
		if err := hook.Send(alert); err != nil {
			log.Printf("alert hook %s failed: %v", hook.Name(), err)
		}
	}
}

// SendCritical sends a critical alert
func (m *AlertManager) SendCritical(component, condition, title, message string, metadata map[string]any) {
	m.Send(Alert{
		Severity:  "critical",
		Title:     title,
		Message:   message,
		Component: component,
		Condition: condition,
		Metadata:  metadata,
	})
}

// SendError sends an error alert
func (m *AlertManager) SendError(component, condition, title, message string, metadata map[string]any) {
	m.Send(Alert{
		Severity:  "error",
		Title:     title,
		Message:   message,
		Component: component,
		Condition: condition,
		Metadata:  metadata,
	})
}

// SendWarning sends a warning alert
func (m *AlertManager) SendWarning(component, condition, title, message string, metadata map[string]any) {
	m.Send(Alert{
		Severity:  "warning",
		Title:     title,
		Message:   message,
		Component: component,
		Condition: condition,
		Metadata:  metadata,
	})
}
