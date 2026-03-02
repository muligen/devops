// Package mq provides RabbitMQ message queue functionality.
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Config holds RabbitMQ configuration.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
}

// URL returns the RabbitMQ connection URL.
func (c *Config) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s", c.User, c.Password, c.Host, c.Port, c.VHost)
}

// Client wraps RabbitMQ connection and channel.
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
}

// Event types
const (
	// Agent events
	EventAgentOnline  = "agent.online"
	EventAgentOffline = "agent.offline"
	EventAgentHeartbeat = "agent.heartbeat"

	// Task events
	EventTaskCreated  = "task.created"
	EventTaskStarted  = "task.started"
	EventTaskCompleted = "task.completed"
	EventTaskFailed   = "task.failed"

	// Update events
	EventUpdateAvailable = "update.available"
)

// Exchange and queue names
const (
	ExchangeEvents = "agentteams.events"
	QueueTasks     = "agentteams.tasks"
	QueueNotifications = "agentteams.notifications"
)

// Event represents a message queue event.
type Event struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// New creates a new RabbitMQ client.
func New(cfg *Config) (*Client, error) {
	conn, err := amqp.Dial(cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &Client{
		conn:    conn,
		channel: channel,
	}

	// Initialize exchanges and queues
	if err := client.init(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	return client, nil
}

// init initializes exchanges and queues.
func (c *Client) init() error {
	// Declare topic exchange for events
	if err := c.channel.ExchangeDeclare(
		ExchangeEvents,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare task queue
	if _, err := c.channel.QueueDeclare(
		QueueTasks,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("failed to declare task queue: %w", err)
	}

	// Declare notification queue
	if _, err := c.channel.QueueDeclare(
		QueueNotifications,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("failed to declare notification queue: %w", err)
	}

	return nil
}

// PublishEvent publishes an event to the exchange.
func (c *Client) PublishEvent(ctx context.Context, eventType string, data map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	event := Event{
		Type:      eventType,
		Timestamp: 0, // Will be set by consumer
		Data:      data,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return c.channel.PublishWithContext(
		ctx,
		ExchangeEvents,
		eventType, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

// PublishTask publishes a task to the task queue.
func (c *Client) PublishTask(ctx context.Context, taskID string, taskData map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	body, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	return c.channel.PublishWithContext(
		ctx,
		"",        // exchange (direct to queue)
		QueueTasks, // routing key (queue name)
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			MessageId:    taskID,
		},
	)
}

// ConsumeEvents consumes events matching the routing keys.
func (c *Client) ConsumeEvents(ctx context.Context, queueName string, routingKeys []string) (<-chan amqp.Delivery, error) {
	// Declare a unique queue for this consumer
	queue, err := c.channel.QueueDeclare(
		queueName,
		false, // durable (auto-delete after consumer disconnects)
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange with routing keys
	for _, key := range routingKeys {
		if err := c.channel.QueueBind(
			queue.Name,
			key,
			ExchangeEvents,
			false,
			nil,
		); err != nil {
			return nil, fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	return c.channel.Consume(
		queue.Name,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
}

// ConsumeTasks consumes tasks from the task queue.
func (c *Client) ConsumeTasks(ctx context.Context) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		QueueTasks,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
}

// Ack acknowledges a message.
func (c *Client) Ack(tag uint64) error {
	return c.channel.Ack(tag, false)
}

// Nack negatively acknowledges a message.
func (c *Client) Nack(tag uint64, requeue bool) error {
	return c.channel.Nack(tag, false, requeue)
}

// Close closes the RabbitMQ connection.
func (c *Client) Close() error {
	var errs []error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connection: %v", errs)
	}

	return nil
}
