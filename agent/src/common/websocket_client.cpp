// WebSocket client implementation with TLS support
#include <agent/websocket_client.hpp>
#include <agent/logger.hpp>

#include <boost/asio.hpp>
#include <boost/asio/ssl.hpp>
#include <boost/beast/core.hpp>
#include <boost/beast/websocket.hpp>
#include <boost/beast/websocket/ssl.hpp>
#include <nlohmann/json.hpp>

#include <openssl/hmac.h>
#include <openssl/sha.h>

#include <atomic>
#include <condition_variable>
#include <mutex>
#include <queue>
#include <thread>
#include <variant>
#include <vector>
#include <iomanip>
#include <sstream>

namespace agent {

namespace beast = boost::beast;
namespace websocket = beast::websocket;
namespace net = boost::asio;
namespace ssl = net::ssl;

using tcp = net::ip::tcp;

// WebSocket stream types
using WsStream = websocket::stream<beast::tcp_stream>;
using WssStream = websocket::stream<ssl::stream<beast::tcp_stream>>;
using WsVariant = std::variant<std::unique_ptr<WsStream>, std::unique_ptr<WssStream>>;

// HMAC-SHA256 computation helper
static std::string ComputeHmacSha256(const std::string& key, const std::string& message) {
    std::vector<unsigned char> digest(SHA256_DIGEST_LENGTH);

    unsigned int digest_len = 0;
    HMAC(EVP_sha256(),
         key.data(), static_cast<int>(key.size()),
         reinterpret_cast<const unsigned char*>(message.data()), message.size(),
         digest.data(), &digest_len);

    // Convert to hex string
    std::ostringstream oss;
    for (unsigned char c : digest) {
        oss << std::hex << std::setw(2) << std::setfill('0') << static_cast<int>(c);
    }
    return oss.str();
}

// SHA256 hash helper
static std::string Sha256Hash(const std::string& input) {
    std::vector<unsigned char> digest(SHA256_DIGEST_LENGTH);
    SHA256(reinterpret_cast<const unsigned char*>(input.data()), input.size(), digest.data());

    // Convert to hex string
    std::ostringstream oss;
    for (unsigned char c : digest) {
        oss << std::hex << std::setw(2) << std::setfill('0') << static_cast<int>(c);
    }
    return oss.str();
}

// Internal implementation using PIMPL pattern
class WebSocketClientImpl : public std::enable_shared_from_this<WebSocketClientImpl> {
public:
    WebSocketClientImpl(net::io_context& io_context, const WebSocketConfig& config)
        : io_context_(io_context)
        , ssl_context_(ssl::context::tlsv12_client)
        , config_(config)
        , state_(ConnectionState::kDisconnected)
        , reconnect_attempts_(0)
        , write_in_progress_(false)
        , secure_(false) {

        // Configure SSL context
        ssl_context_.set_default_verify_paths();
        // Disable certificate verification for local testing
        ssl_context_.set_verify_mode(ssl::verify_none);

        // Allow TLS 1.2 and above
        ssl_context_.set_options(
            ssl::context::default_workarounds |
            ssl::context::no_sslv2 |
            ssl::context::no_sslv3 |
            ssl::context::no_tlsv1 |
            ssl::context::no_tlsv1_1
        );
    }

    ~WebSocketClientImpl() {
        Disconnect();
    }

    void Connect(const std::string& url) {
        if (state_ != ConnectionState::kDisconnected) {
            return;
        }

        server_url_ = url;
        SetState(ConnectionState::kConnecting);

        // Parse URL
        auto parsed = ParseUrl(url);
        if (!parsed) {
            HandleError("Invalid URL: " + url);
            return;
        }

        secure_ = parsed->secure;

        // Create appropriate WebSocket stream
        if (secure_) {
            ws_ = std::make_unique<WssStream>(net::make_strand(io_context_), ssl_context_);
        } else {
            ws_ = std::make_unique<WsStream>(net::make_strand(io_context_));
        }

        // Start async resolve
        auto self = shared_from_this();
        resolver_.async_resolve(
            parsed->host,
            parsed->port,
            [self](const beast::error_code& ec, const tcp::resolver::results_type& results) {
                self->OnResolve(ec, results);
            }
        );
    }

    void Disconnect() {
        if (state_ == ConnectionState::kDisconnected) {
            return;
        }

        // Stop keepalive
        StopKeepalive();

        // Close WebSocket
        std::visit([](auto& ws) {
            if (ws && ws->is_open()) {
                beast::error_code ec;
                ws->close(websocket::close_code::normal, ec);
            }
        }, ws_);

        SetState(ConnectionState::kDisconnected);
        session_id_.clear();
        reconnect_attempts_ = 0;

        // Clear message queue
        std::lock_guard<std::mutex> lock(write_mutex_);
        while (!write_queue_.empty()) {
            write_queue_.pop();
        }
        write_in_progress_ = false;
    }

    ConnectionState GetState() const {
        return state_.load();
    }

    bool IsConnected() const {
        return state_ == ConnectionState::kAuthenticated;
    }

    std::string GetSessionId() const {
        return session_id_;
    }

    bool IsAuthenticated() const {
        return state_ == ConnectionState::kAuthenticated && !session_id_.empty();
    }

    void Send(const std::string& message) {
        if (!IsConnected()) {
            HandleError("Not connected");
            return;
        }

        net::post(io_context_, [self = shared_from_this(), message]() {
            self->DoWrite(message);
        });
    }

    void SetMessageCallback(MessageCallback callback) {
        message_callback_ = std::move(callback);
    }

    void SetErrorCallback(ErrorCallback callback) {
        error_callback_ = std::move(callback);
    }

    void SetConnectionCallback(ConnectionCallback callback) {
        connection_callback_ = std::move(callback);
    }

    void SetAuthCallback(AuthCallback callback) {
        auth_callback_ = std::move(callback);
    }

private:
    struct ParsedUrl {
        std::string host;
        std::string port;
        std::string path;
        bool secure;
    };

    std::optional<ParsedUrl> ParseUrl(const std::string& url) {
        // Parse wss://host:port/path
        ParsedUrl result;

        size_t pos = 0;
        if (url.substr(0, 6) == "wss://") {
            result.secure = true;
            result.port = "443";
            pos = 6;
        } else if (url.substr(0, 5) == "ws://") {
            result.secure = false;
            result.port = "80";
            pos = 5;
        } else {
            return std::nullopt;
        }

        // Find host:port
        size_t slash_pos = url.find('/', pos);
        std::string host_port;

        if (slash_pos == std::string::npos) {
            host_port = url.substr(pos);
            result.path = "/";
        } else {
            host_port = url.substr(pos, slash_pos - pos);
            result.path = url.substr(slash_pos);
        }

        // Parse host and port
        size_t colon_pos = host_port.find(':');
        if (colon_pos != std::string::npos) {
            result.host = host_port.substr(0, colon_pos);
            result.port = host_port.substr(colon_pos + 1);
        } else {
            result.host = host_port;
        }

        return result;
    }

    void SetState(ConnectionState new_state) {
        ConnectionState old_state = state_.exchange(new_state);
        if (old_state != new_state && connection_callback_) {
            connection_callback_(new_state);
        }
    }

    void HandleError(const std::string& error) {
        if (error_callback_) {
            error_callback_(error);
        }
        ScheduleReconnect();
    }

    void OnResolve(const beast::error_code& ec, const tcp::resolver::results_type& results) {
        if (ec) {
            HandleError("Resolve failed: " + ec.message());
            return;
        }

        // Connect to endpoint
        auto self = shared_from_this();

        if (secure_) {
            auto& wss = std::get<std::unique_ptr<WssStream>>(ws_);
            beast::get_lowest_layer(*wss).async_connect(
                results,
                [self](const beast::error_code& ec, const tcp::endpoint& endpoint) {
                    self->OnConnect(ec);
                }
            );
        } else {
            auto& ws = std::get<std::unique_ptr<WsStream>>(ws_);
            beast::get_lowest_layer(*ws).async_connect(
                results,
                [self](const beast::error_code& ec, const tcp::endpoint& endpoint) {
                    self->OnConnect(ec);
                }
            );
        }
    }

    void OnConnect(const beast::error_code& ec) {
        if (ec) {
            HandleError("Connect failed: " + ec.message());
            return;
        }

        auto self = shared_from_this();

        if (secure_) {
            // Perform SSL handshake for secure connections
            auto& wss = std::get<std::unique_ptr<WssStream>>(ws_);
            wss->next_layer().async_handshake(
                ssl::stream_base::client,
                [self](const beast::error_code& ec) {
                    self->OnSslHandshake(ec);
                }
            );
        } else {
            // Skip SSL handshake for non-secure connections
            DoWebSocketHandshake();
        }
    }

    void OnSslHandshake(const beast::error_code& ec) {
        if (ec) {
            HandleError("SSL handshake failed: " + ec.message());
            return;
        }

        DoWebSocketHandshake();
    }

    void DoWebSocketHandshake() {
        auto parsed = ParseUrl(server_url_);
        if (!parsed) {
            HandleError("Invalid URL during handshake");
            return;
        }

        auto self = shared_from_this();

        if (secure_) {
            auto& wss = std::get<std::unique_ptr<WssStream>>(ws_);
            wss->set_option(websocket::stream_base::decorator(
                [](websocket::request_type& req) {
                    req.set(beast::http::field::user_agent,
                           std::string(BOOST_BEAST_VERSION_STRING) + " AgentTeams-Agent");
                }
            ));

            wss->control_callback(
                [self](beast::websocket::frame_type type, beast::string_view payload) {
                    if (type == beast::websocket::frame_type::pong) {
                        self->OnPong();
                    }
                }
            );

            wss->async_handshake(
                parsed->host,
                parsed->path,
                [self](const beast::error_code& ec) {
                    self->OnHandshake(ec);
                }
            );
        } else {
            auto& ws = std::get<std::unique_ptr<WsStream>>(ws_);
            ws->set_option(websocket::stream_base::decorator(
                [](websocket::request_type& req) {
                    req.set(beast::http::field::user_agent,
                           std::string(BOOST_BEAST_VERSION_STRING) + " AgentTeams-Agent");
                }
            ));

            ws->control_callback(
                [self](beast::websocket::frame_type type, beast::string_view payload) {
                    if (type == beast::websocket::frame_type::pong) {
                        self->OnPong();
                    }
                }
            );

            ws->async_handshake(
                parsed->host,
                parsed->path,
                [self](const beast::error_code& ec) {
                    self->OnHandshake(ec);
                }
            );
        }
    }

    void OnHandshake(const beast::error_code& ec) {
        if (ec) {
            HandleError("WebSocket handshake failed: " + ec.message());
            return;
        }

        SetState(ConnectionState::kAuthenticating);
        reconnect_attempts_ = 0;

        // Start reading messages
        DoRead();

        // Send authentication init message
        SendAuthInit();
    }

    void SendAuthInit() {
        // Send authentication init with agent_id
        nlohmann::json msg;
        msg["type"] = "auth";
        msg["data"] = {
            {"agent_id", config_.agent_id}
        };

        DoWrite(msg.dump());
    }

    void HandleAuthChallenge(const nlohmann::json& msg) {
        std::string nonce = msg["nonce"];

        // Compute HMAC(SHA256(token), nonce)
        // Server stores token_hash = SHA256(token), so we need to hash token first
        std::string token_hash = Sha256Hash(config_.token);
        std::string response = ComputeHmacSha256(token_hash, nonce);

        // Send auth_response
        nlohmann::json reply;
        reply["type"] = "challenge";
        reply["data"] = {
            {"response", response}
        };

        DoWrite(reply.dump());
    }

    void HandleAuthSuccess(const nlohmann::json& msg) {
        session_id_ = msg.value("session_id", "");

        SetState(ConnectionState::kAuthenticated);

        // Start keepalive
        StartKeepalive();

        if (auth_callback_) {
            auth_callback_(true, session_id_);
        }
    }

    void HandleAuthFailure(const nlohmann::json& msg) {
        std::string error = msg.value("error", "Authentication failed");

        if (auth_callback_) {
            auth_callback_(false, "");
        }

        HandleError("Authentication failed: " + error);
    }

    void StartKeepalive() {
        if (config_.ping_interval.count() > 0) {
            SchedulePing();
        }
    }

    void StopKeepalive() {
        ping_timer_.cancel();
        pong_timeout_timer_.cancel();
        waiting_pong_ = false;
    }

    void SchedulePing() {
        auto self = shared_from_this();
        ping_timer_.expires_after(config_.ping_interval);
        ping_timer_.async_wait([self](const beast::error_code& ec) {
            if (!ec && self->state_ == ConnectionState::kAuthenticated) {
                self->SendPing();
            }
        });
    }

    void SendPing() {
        if (waiting_pong_) {
            // Previous ping not acknowledged
            HandleError("Pong timeout - connection may be dead");
            return;
        }

        auto self = shared_from_this();

        if (secure_) {
            auto& wss = std::get<std::unique_ptr<WssStream>>(ws_);
            wss->async_ping(
                beast::websocket::ping_data{},
                [self](const beast::error_code& ec) {
                    if (!ec) {
                        self->waiting_pong_ = true;
                        self->StartPongTimeout();
                    } else {
                        self->HandleError("Ping failed: " + ec.message());
                    }
                }
            );
        } else {
            auto& ws = std::get<std::unique_ptr<WsStream>>(ws_);
            ws->async_ping(
                beast::websocket::ping_data{},
                [self](const beast::error_code& ec) {
                    if (!ec) {
                        self->waiting_pong_ = true;
                        self->StartPongTimeout();
                    } else {
                        self->HandleError("Ping failed: " + ec.message());
                    }
                }
            );
        }
    }

    void StartPongTimeout() {
        auto self = shared_from_this();
        pong_timeout_timer_.expires_after(config_.pong_timeout);
        pong_timeout_timer_.async_wait([self](const beast::error_code& ec) {
            if (!ec && self->waiting_pong_) {
                self->HandleError("Pong timeout");
            }
        });
    }

    void OnPong() {
        waiting_pong_ = false;
        pong_timeout_timer_.cancel();

        // Schedule next ping
        if (state_ == ConnectionState::kAuthenticated) {
            SchedulePing();
        }
    }

    void DoRead() {
        auto self = shared_from_this();

        if (secure_) {
            auto& wss = std::get<std::unique_ptr<WssStream>>(ws_);
            wss->async_read(
                read_buffer_,
                [self](const beast::error_code& ec, std::size_t bytes_transferred) {
                    self->OnRead(ec, bytes_transferred);
                }
            );
        } else {
            auto& ws = std::get<std::unique_ptr<WsStream>>(ws_);
            ws->async_read(
                read_buffer_,
                [self](const beast::error_code& ec, std::size_t bytes_transferred) {
                    self->OnRead(ec, bytes_transferred);
                }
            );
        }
    }

    void OnRead(const beast::error_code& ec, std::size_t bytes_transferred) {
        if (ec) {
            if (ec == websocket::error::closed) {
                // Connection closed by server
                session_id_.clear();
                SetState(ConnectionState::kDisconnected);
                ScheduleReconnect();
            } else {
                HandleError("Read error: " + ec.message());
            }
            return;
        }

        // Process message
        std::string message = beast::buffers_to_string(read_buffer_.data());
        read_buffer_.consume(bytes_transferred);

        // Handle authentication messages
        try {
            auto msg = nlohmann::json::parse(message);
            std::string type = msg.value("type", "");

            if (state_ == ConnectionState::kAuthenticating) {
                // Handle authentication flow
                if (type == "challenge") {
                    auto data = msg["data"];
                    HandleAuthChallenge(data);
                } else if (type == "auth_result") {
                    auto data = msg["data"];
                    bool success = data.value("success", false);
                    if (success) {
                        HandleAuthSuccess(data);
                    } else {
                        HandleAuthFailure(data);
                    }
                } else {
                    // Unexpected message during auth
                    HandleError("Unexpected message during authentication: " + type);
                }
            } else {
                // Normal message handling
                if (message_callback_) {
                    message_callback_(message);
                }
            }
        } catch (const nlohmann::json::exception& e) {
            HandleError("Failed to parse message: " + std::string(e.what()));
        }

        // Continue reading
        DoRead();
    }

    void DoWrite(const std::string& message) {
        bool write_in_progress = write_in_progress_.exchange(true);
        if (write_in_progress) {
            // Queue message
            std::lock_guard<std::mutex> lock(write_mutex_);
            write_queue_.push(message);
            return;
        }

        // Write message directly
        auto self = shared_from_this();

        if (secure_) {
            auto& wss = std::get<std::unique_ptr<WssStream>>(ws_);
            wss->async_write(
                net::buffer(message),
                [self](const beast::error_code& ec, std::size_t bytes_transferred) {
                    self->OnWrite(ec, bytes_transferred);
                }
            );
        } else {
            auto& ws = std::get<std::unique_ptr<WsStream>>(ws_);
            ws->async_write(
                net::buffer(message),
                [self](const beast::error_code& ec, std::size_t bytes_transferred) {
                    self->OnWrite(ec, bytes_transferred);
                }
            );
        }
    }

    void OnWrite(const beast::error_code& ec, std::size_t bytes_transferred) {
        if (ec) {
            write_in_progress_ = false;
            HandleError("Write error: " + ec.message());
            return;
        }

        // Check for more messages in queue
        std::lock_guard<std::mutex> lock(write_mutex_);
        if (write_queue_.empty()) {
            write_in_progress_ = false;
        } else {
            std::string message = std::move(write_queue_.front());
            write_queue_.pop();

            auto self = shared_from_this();

            if (secure_) {
                auto& wss = std::get<std::unique_ptr<WssStream>>(ws_);
                wss->async_write(
                    net::buffer(message),
                    [self](const beast::error_code& ec, std::size_t bytes_transferred) {
                        self->OnWrite(ec, bytes_transferred);
                    }
                );
            } else {
                auto& ws = std::get<std::unique_ptr<WsStream>>(ws_);
                ws->async_write(
                    net::buffer(message),
                    [self](const beast::error_code& ec, std::size_t bytes_transferred) {
                        self->OnWrite(ec, bytes_transferred);
                    }
                );
            }
        }
    }

    void ScheduleReconnect() {
        SetState(ConnectionState::kDisconnected);

        // Calculate backoff delay
        auto delay = CalculateReconnectDelay();
        reconnect_attempts_++;

        // Schedule reconnect
        auto self = shared_from_this();
        reconnect_timer_.expires_after(delay);
        reconnect_timer_.async_wait([self](const beast::error_code& ec) {
            if (!ec) {
                self->Connect(self->server_url_);
            }
        });
    }

    std::chrono::seconds CalculateReconnectDelay() const {
        // Exponential backoff: 5s, 10s, 20s, 40s, 60s (max)
        auto base = config_.reconnect_base_delay.count();
        auto max = config_.reconnect_max_delay.count();

        auto delay = base * (1 << std::min(reconnect_attempts_.load(), 4));
        return std::chrono::seconds(std::min(delay, max));
    }

    // Members
    net::io_context& io_context_;
    ssl::context ssl_context_;
    WsVariant ws_;
    tcp::resolver resolver_{net::make_strand(io_context_)};
    net::steady_timer reconnect_timer_{io_context_};
    net::steady_timer ping_timer_{io_context_};
    net::steady_timer pong_timeout_timer_{io_context_};
    WebSocketConfig config_;

    std::atomic<ConnectionState> state_;
    std::string server_url_;
    std::string session_id_;
    std::atomic<int> reconnect_attempts_;
    std::atomic<bool> waiting_pong_{false};
    bool secure_;

    beast::flat_buffer read_buffer_;
    std::mutex write_mutex_;
    std::queue<std::string> write_queue_;
    std::atomic<bool> write_in_progress_;

    MessageCallback message_callback_;
    ErrorCallback error_callback_;
    ConnectionCallback connection_callback_;
    AuthCallback auth_callback_;
};

// Factory method implementation
std::unique_ptr<WebSocketClient> WebSocketClient::Create(const WebSocketConfig& config) {
    class WebSocketClientWrapper : public WebSocketClient {
    public:
        WebSocketClientWrapper(const WebSocketConfig& config)
            : io_context_()
            , work_guard_(net::make_work_guard(io_context_))
            , impl_(std::make_shared<WebSocketClientImpl>(io_context_, config)) {
            // Start IO thread
            io_thread_ = std::thread([this]() {
                io_context_.run();
            });
        }

        ~WebSocketClientWrapper() override {
            impl_->Disconnect();
            work_guard_.reset();
            io_context_.stop();
            if (io_thread_.joinable()) {
                io_thread_.join();
            }
        }

        void Connect(const std::string& url) override {
            net::post(io_context_, [this, url]() {
                impl_->Connect(url);
            });
        }

        void Disconnect() override {
            impl_->Disconnect();
        }

        ConnectionState GetState() const override {
            return impl_->GetState();
        }

        bool IsConnected() const override {
            return impl_->IsConnected();
        }

        std::string GetSessionId() const override {
            return impl_->GetSessionId();
        }

        bool IsAuthenticated() const override {
            return impl_->IsAuthenticated();
        }

        void Send(const std::string& message) override {
            impl_->Send(message);
        }

        void SetMessageCallback(MessageCallback callback) override {
            impl_->SetMessageCallback(std::move(callback));
        }

        void SetErrorCallback(ErrorCallback callback) override {
            impl_->SetErrorCallback(std::move(callback));
        }

        void SetConnectionCallback(ConnectionCallback callback) override {
            impl_->SetConnectionCallback(std::move(callback));
        }

        void SetAuthCallback(AuthCallback callback) override {
            impl_->SetAuthCallback(std::move(callback));
        }

    private:
        net::io_context io_context_;
        net::executor_work_guard<net::io_context::executor_type> work_guard_;
        std::shared_ptr<WebSocketClientImpl> impl_;
        std::thread io_thread_;
    };

    return std::make_unique<WebSocketClientWrapper>(config);
}

}  // namespace agent
