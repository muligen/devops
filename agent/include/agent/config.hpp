#ifndef AGENT_CONFIG_HPP
#define AGENT_CONFIG_HPP

#include <agent/common/types.hpp>
#include <chrono>
#include <optional>
#include <string>

namespace agent {

// Agent configuration
struct AgentConfig {
    std::string id;
    std::string token;
    std::string server_url;
};

// Connection configuration
struct ConnectionConfig {
    std::chrono::seconds retry_interval{5};
    std::chrono::seconds max_retry_interval{60};
    std::chrono::seconds ping_interval{10};
    std::chrono::seconds pong_timeout{5};
};

// Heartbeat configuration
struct HeartbeatConfig {
    std::chrono::milliseconds interval{1000};
};

// Metrics configuration
struct MetricsConfig {
    std::chrono::minutes interval{1};
};

// Task configuration
struct TaskConfig {
    size_t max_concurrent{4};
    size_t queue_size{100};
    std::chrono::minutes default_timeout{5};
};

// Update configuration
struct UpdateConfig {
    std::chrono::hours check_interval{1};
    bool idle_required{true};
};

// Logging configuration
struct LoggingConfig {
    std::string level{"info"};
    std::string file;
    size_t max_size{100 * 1024 * 1024};  // 100MB
    size_t max_files{5};
};

// Main configuration structure
struct Config {
    AgentConfig agent;
    ConnectionConfig connection;
    HeartbeatConfig heartbeat;
    MetricsConfig metrics;
    TaskConfig task;
    UpdateConfig update;
    LoggingConfig logging;

    // Load configuration from file
    static std::optional<Config> LoadFromFile(const std::string& path);

    // Load configuration from YAML string
    static std::optional<Config> LoadFromYaml(const std::string& yaml_content);

    // Validate configuration
    bool Validate() const;
};

}  // namespace agent

#endif  // AGENT_CONFIG_HPP
