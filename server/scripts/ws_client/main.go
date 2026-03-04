package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	agentID := "f7d2a25c-92eb-4b49-ae07-62b052918b19"
	token := "6ebed1148e901bb254873761254c2492be25cc3ad4d1ed5fea60bbefbdf77356" //nolint:gosec // test token

	// Compute token hash
	tokenHashBytes := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	// Connect
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/api/v1/agent/ws"}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial:", err)
	}
	defer c.Close()

	var mu sync.Mutex
	_ = mu // mutex for future use

	// Read handler
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Connection closed:", err)
				return
			}
			log.Printf("<< %s", message)

			var msg struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}
			json.Unmarshal(message, &msg)

			switch msg.Type {
			case "challenge":
				var data struct {
					Nonce string `json:"nonce"`
				}
				json.Unmarshal(msg.Data, &data)

				// Compute response
				h := hmac.New(sha256.New, []byte(tokenHash))
				h.Write([]byte(data.Nonce))
				response := hex.EncodeToString(h.Sum(nil))

				resp := map[string]interface{}{
					"type": "challenge",
					"data": map[string]string{"response": response},
				}
				respBytes, _ := json.Marshal(resp)
				c.WriteMessage(websocket.TextMessage, respBytes)
				log.Printf(">> Sent challenge response")

			case "auth_result":
				var data struct {
					Success bool   `json:"success"`
					Message string `json:"message"`
				}
				json.Unmarshal(msg.Data, &data)
				if data.Success {
					log.Println("=== AUTHENTICATED ===")

					// Send heartbeat
					hb := map[string]interface{}{
						"type": "heartbeat",
						"data": map[string]interface{}{"timestamp": time.Now().Unix()},
					}
					hbBytes, _ := json.Marshal(hb)
					c.WriteMessage(websocket.TextMessage, hbBytes)

					// Send metrics
					m := map[string]interface{}{
						"type": "metrics",
						"data": map[string]interface{}{
							"cpu_usage": 25.5, "memory_total": 16000000000,
							"memory_used": 8000000000, "memory_percent": 50.0,
							"disk_total": 500000000000, "disk_used": 250000000000,
							"disk_percent": 50.0, "uptime": 7200,
						},
					}
					mBytes, _ := json.Marshal(m)
					c.WriteMessage(websocket.TextMessage, mBytes)
					log.Println(">> Sent heartbeat and metrics")
				}

			case "command":
				var cmd struct {
					CommandID   string                 `json:"command_id"`
					CommandType string                 `json:"command_type"`
					Params      map[string]interface{} `json:"params"`
					Timeout     int                    `json:"timeout"`
				}
				json.Unmarshal(msg.Data, &cmd)
				log.Printf(">>> EXECUTING: %s (id: %s)", cmd.CommandType, cmd.CommandID)

				// Simulate execution
				time.Sleep(500 * time.Millisecond)

				// Send result
				result := map[string]interface{}{
					"type": "result",
					"data": map[string]interface{}{
						"command_id": cmd.CommandID,
						"status":     "success",
						"exit_code":  0,
						"output":     "Hello from agent! Command executed successfully.\n",
						"duration":   0.5,
					},
				}
				resultBytes, _ := json.Marshal(result)
				c.WriteMessage(websocket.TextMessage, resultBytes)
				log.Printf(">> Sent result for task: %s", cmd.CommandID)

			case "heartbeat":
				// Heartbeat ack
			}
		}
	}()

	// Send auth request
	authReq := map[string]interface{}{
		"type": "auth",
		"data": map[string]string{"agent_id": agentID},
	}
	authBytes, _ := json.Marshal(authReq)
	c.WriteMessage(websocket.TextMessage, authBytes)
	log.Println(">> Sent auth request")

	// Wait for authentication
	time.Sleep(1 * time.Second)

	// Keep running for 30 seconds to receive commands
	log.Println("Waiting for commands (30s)...")
	time.Sleep(30 * time.Second)

	log.Println("Test completed!")
}
