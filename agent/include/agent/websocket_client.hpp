#ifndef AGENT_WEBSOCKET_CLIENT_HPP
#define AGENT_WEBSOCKET_CLIENT_HPP

// Undef Windows macros that conflict with our methods
#ifdef SendMessage
#undef SendMessage
#endif

#include <agent/common/types.hpp>
#include <chrono>
#include <functional>
#include <memory>
#include <string>

namespace agent {

// Forward declarations
class Message;

// Callback types
using MessageCallback = std::function<void(const std::string&)>;
using ErrorCallback = std::function<void(const std::string&)>;
using ConnectionCallback = std::function<void(ConnectionState)>;
using AuthCallback = std::function<void(bool success, const std::string& session_id)>;

// WebSocket client configuration
struct WebSocketConfig {
    std::string server_url;
    std::string agent_id;
    std::string token;
    std::chrono::seconds connect_timeout{kDefaultConnectTimeout};
    std::chrono::seconds reconnect_base_delay{kDefaultReconnectBaseDelay};
    std::chrono::seconds reconnect_max_delay{kDefaultReconnectMaxDelay};
    std::chrono::seconds ping_interval{std::chrono::seconds(10)};
    std::chrono::seconds pong_timeout{std::chrono::seconds(5)};
};

// WebSocket client interface
class WebSocketClient {
public:
    virtual ~WebSocketClient() = default;

    // Connection management
    virtual void Connect(const std::string& url) = 0;
    virtual void Disconnect() = 0;
    virtual ConnectionState GetState() const = 0;
    virtual bool IsConnected() const = 0;

    // Session management
    virtual std::string GetSessionId() const = 0;
    virtual bool IsAuthenticated() const = 0;

    // Message sending
    virtual void Send(const std::string& message) = 0;

    // Callbacks
    virtual void SetMessageCallback(MessageCallback callback) = 0;
    virtual void SetErrorCallback(ErrorCallback callback) = 0;
    virtual void SetConnectionCallback(ConnectionCallback callback) = 0;
    virtual void SetAuthCallback(AuthCallback callback) = 0;

    // Factory method
    static std::unique_ptr<WebSocketClient> Create(const WebSocketConfig& config);
};

}  // namespace agent

#endif  // AGENT_WEBSOCKET_CLIENT_HPP
