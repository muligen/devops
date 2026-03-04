// Package service provides notification dispatching functionality.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/agentteams/server/internal/pkg/logger"
)

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeAlert   NotificationType = "alert"
	NotificationTypeOffline NotificationType = "agent_offline"
	NotificationTypeOnline  NotificationType = "agent_online"
)

// Notification represents a notification to be sent.
type Notification struct {
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Severity  string                 `json:"severity"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// WebhookConfig represents webhook notification configuration.
type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Enabled bool              `json:"enabled"`
}

// EmailConfig represents email notification configuration.
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	FromAddress  string `json:"from_address"`
	Enabled      bool   `json:"enabled"`
}

// NotificationDispatcher handles sending notifications.
type NotificationDispatcher struct {
	webhookConfig *WebhookConfig
	emailConfig   *EmailConfig
	httpClient    *http.Client
	logger        *logger.Logger
	queue         chan Notification
	wg            sync.WaitGroup
	stopCh        chan struct{}
}

// NewNotificationDispatcher creates a new notification dispatcher.
func NewNotificationDispatcher(webhook *WebhookConfig, email *EmailConfig, log *logger.Logger) *NotificationDispatcher {
	if webhook == nil {
		webhook = &WebhookConfig{}
	}
	if email == nil {
		email = &EmailConfig{}
	}

	return &NotificationDispatcher{
		webhookConfig: webhook,
		emailConfig:   email,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: log,
		queue:  make(chan Notification, 100),
		stopCh: make(chan struct{}),
	}
}

// Start starts the notification dispatcher workers.
func (d *NotificationDispatcher) Start(ctx context.Context) {
	// Start worker goroutines
	for i := 0; i < 3; i++ {
		d.wg.Add(1)
		go d.worker(ctx)
	}
}

// Stop stops the notification dispatcher.
func (d *NotificationDispatcher) Stop() {
	close(d.stopCh)
	d.wg.Wait()
}

// worker processes notifications from the queue.
func (d *NotificationDispatcher) worker(ctx context.Context) {
	defer d.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case notification := <-d.queue:
			if err := d.dispatch(ctx, notification); err != nil && d.logger != nil {
				d.logger.Errorw("Failed to dispatch notification",
					"error", err,
					"type", notification.Type,
					"title", notification.Title,
				)
			}
		}
	}
}

// Queue queues a notification for sending.
func (d *NotificationDispatcher) Queue(notification Notification) error {
	select {
	case d.queue <- notification:
		return nil
	default:
		return fmt.Errorf("notification queue is full")
	}
}

// dispatch sends the notification through configured channels.
func (d *NotificationDispatcher) dispatch(ctx context.Context, notification Notification) error {
	var errs []error

	// Send webhook notification
	if d.webhookConfig.Enabled && d.webhookConfig.URL != "" {
		if err := d.sendWebhook(ctx, notification); err != nil {
			errs = append(errs, fmt.Errorf("webhook: %w", err))
		}
	}

	// Send email notification
	if d.emailConfig.Enabled {
		// Email sending would be implemented here
		// For now, just log it
		if d.logger != nil {
			d.logger.Infow("Email notification",
				"type", notification.Type,
				"title", notification.Title,
				"message", notification.Message,
			)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("notification dispatch errors: %v", errs)
	}

	return nil
}

// sendWebhook sends a notification via webhook.
func (d *NotificationDispatcher) sendWebhook(ctx context.Context, notification Notification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.webhookConfig.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range d.webhookConfig.Headers {
		req.Header.Set(key, value)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	if d.logger != nil {
		d.logger.Infow("Webhook notification sent",
			"type", notification.Type,
			"title", notification.Title,
			"status", resp.StatusCode,
		)
	}

	return nil
}

// NotifyAlert sends an alert notification.
func (d *NotificationDispatcher) NotifyAlert(ctx context.Context, ruleName, agentID, severity string, value, threshold float64) error {
	notification := Notification{
		Type:     NotificationTypeAlert,
		Title:    fmt.Sprintf("Alert: %s", ruleName),
		Message:  fmt.Sprintf("Agent %s: value %.2f exceeds threshold %.2f", agentID, value, threshold),
		Severity: severity,
		Data: map[string]interface{}{
			"rule_name": ruleName,
			"agent_id":  agentID,
			"value":     value,
			"threshold": threshold,
		},
		Timestamp: time.Now(),
	}

	return d.Queue(notification)
}

// NotifyAgentOffline sends an agent offline notification.
func (d *NotificationDispatcher) NotifyAgentOffline(ctx context.Context, agentID, agentName string) error {
	notification := Notification{
		Type:     NotificationTypeOffline,
		Title:    "Agent Offline",
		Message:  fmt.Sprintf("Agent %s (%s) has gone offline", agentName, agentID),
		Severity: "warning",
		Data: map[string]interface{}{
			"agent_id":   agentID,
			"agent_name": agentName,
		},
		Timestamp: time.Now(),
	}

	return d.Queue(notification)
}

// NotifyAgentOnline sends an agent online notification.
func (d *NotificationDispatcher) NotifyAgentOnline(ctx context.Context, agentID, agentName string) error {
	notification := Notification{
		Type:     NotificationTypeOnline,
		Title:    "Agent Online",
		Message:  fmt.Sprintf("Agent %s (%s) is now online", agentName, agentID),
		Severity: "info",
		Data: map[string]interface{}{
			"agent_id":   agentID,
			"agent_name": agentName,
		},
		Timestamp: time.Now(),
	}

	return d.Queue(notification)
}
