// Package e2e provides end-to-end testing infrastructure
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// HTTPClient provides helper methods for HTTP requests
type HTTPClient struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetAuthToken sets the authorization token
func (c *HTTPClient) SetAuthToken(token string) {
	c.authToken = token
}

// Get performs a GET request
func (c *HTTPClient) Get(path string) (*HTTPResponse, error) {
	return c.doRequest("GET", path, nil)
}

// Post performs a POST request
func (c *HTTPClient) Post(path string, body interface{}) (*HTTPResponse, error) {
	return c.doRequest("POST", path, body)
}

// Delete performs a DELETE request
func (c *HTTPClient) Delete(path string) (*HTTPResponse, error) {
	return c.doRequest("DELETE", path, nil)
}

// doRequest performs an HTTP request
func (c *HTTPClient) doRequest(method, path string, body interface{}) (*HTTPResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// JSON unmarshals the response body into v
func (r *HTTPResponse) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String returns the response body as a string
func (r *HTTPResponse) String() string {
	return string(r.Body)
}

// WSClient provides WebSocket client functionality for testing
type WSClient struct {
	conn       *websocket.Conn
	url        string
	msgChan    chan []byte
	errChan    chan error
	closeChan  chan struct{}
}

// NewWSClient creates a new WebSocket client
func NewWSClient(baseURL string) *WSClient {
	// Convert HTTP URL to WebSocket URL
	wsURL := baseURL
	if len(wsURL) > 4 && wsURL[:4] == "http" {
		wsURL = "ws" + wsURL[4:]
	}
	return &WSClient{
		url:       wsURL,
		msgChan:   make(chan []byte, 100),
		errChan:   make(chan error, 10),
		closeChan: make(chan struct{}),
	}
}

// Connect establishes a WebSocket connection
func (c *WSClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, resp, err := dialer.Dial(c.url+"/ws", nil)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	go c.readLoop()

	return nil
}

// readLoop reads messages from the WebSocket connection
func (c *WSClient) readLoop() {
	for {
		select {
		case <-c.closeChan:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				select {
				case c.errChan <- err:
				default:
				}
				return
			}
			select {
			case c.msgChan <- message:
			default:
			}
		}
	}
}

// Send sends a JSON message
func (c *WSClient) Send(v interface{}) error {
	return c.conn.WriteJSON(v)
}

// SendMessage sends a raw message
func (c *WSClient) SendMessage(msg []byte) error {
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

// Receive waits for a message with timeout
func (c *WSClient) Receive(timeout time.Duration) ([]byte, error) {
	select {
	case msg := <-c.msgChan:
		return msg, nil
	case err := <-c.errChan:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for message")
	}
}

// ReceiveJSON waits for a JSON message and unmarshals it
func (c *WSClient) ReceiveJSON(timeout time.Duration, v interface{}) error {
	msg, err := c.Receive(timeout)
	if err != nil {
		return err
	}
	return json.Unmarshal(msg, v)
}

// WaitForMessage waits for a specific message type
func (c *WSClient) WaitForMessage(timeout time.Duration, expectedType string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for message type %s", expectedType)
		case msg := <-c.msgChan:
			var result map[string]interface{}
			if err := json.Unmarshal(msg, &result); err != nil {
				continue
			}
			if msgType, ok := result["type"].(string); ok && msgType == expectedType {
				return result, nil
			}
		}
	}
}

// Close closes the WebSocket connection
func (c *WSClient) Close() error {
	close(c.closeChan)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
