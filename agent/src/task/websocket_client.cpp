// Task worker websocket_client stub
// The task worker communicates with the main process via IPC, not WebSocket directly.
// This file is kept for build compatibility.

#include <agent/websocket_client.hpp>

namespace agent {

std::unique_ptr<WebSocketClient> WebSocketClient::Create(const WebSocketConfig& config) {
    // Task worker doesn't use WebSocket directly
    return nullptr;
}

}  // namespace agent
