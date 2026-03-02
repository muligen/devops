// Agent configuration implementation
#include <agent/config.hpp>
#include <agent/logger.hpp>
#include <fstream>
#include <yaml-cpp/yaml.h>

namespace agent {

namespace {

// YAML conversion specializations
template<typename T>
std::optional<T> GetOptional(const YAML::Node& node, const std::string& key) {
    if (node[key]) {
        try {
            return node[key].as<T>();
        } catch (...) {
            return std::nullopt;
        }
    }
    return std::nullopt;
}

template<typename T>
T GetValue(const YAML::Node& node, const std::string& key, T default_value) {
    auto opt = GetOptional<T>(node, key);
    return opt.value_or(default_value);
}

}  // namespace

std::optional<Config> Config::LoadFromFile(const std::string& path) {
    std::ifstream file(path);
    if (!file.is_open()) {
        LOG_ERROR("Failed to open config file: {}", path);
        return std::nullopt;
    }

    std::string content((std::istreambuf_iterator<char>(file)),
                        std::istreambuf_iterator<char>());
    return LoadFromYaml(content);
}

std::optional<Config> Config::LoadFromYaml(const std::string& yaml_content) {
    try {
        YAML::Node root = YAML::Load(yaml_content);
        Config config;

        // Parse server configuration
        if (root["server"]) {
            auto server = root["server"];
            config.agent.server_url = GetValue<std::string>(server, "url", "");
        }

        // Parse auth configuration
        if (root["auth"]) {
            auto auth = root["auth"];
            config.agent.id = GetValue<std::string>(auth, "agent_id", "");
            config.agent.token = GetValue<std::string>(auth, "token", "");
        }

        // Parse connection configuration
        if (root["connection"]) {
            auto conn = root["connection"];
            config.connection.retry_interval = std::chrono::seconds(
                GetValue<int>(conn, "retry_interval", 5));
            config.connection.max_retry_interval = std::chrono::seconds(
                GetValue<int>(conn, "max_retry_interval", 60));
            config.connection.ping_interval = std::chrono::seconds(
                GetValue<int>(conn, "ping_interval", 10));
            config.connection.pong_timeout = std::chrono::seconds(
                GetValue<int>(conn, "pong_timeout", 5));
        }

        // Parse heartbeat configuration
        if (root["heartbeat"]) {
            auto hb = root["heartbeat"];
            config.heartbeat.interval = std::chrono::milliseconds(
                GetValue<int>(hb, "interval", 1000));
        }

        // Parse metrics configuration
        if (root["metrics"]) {
            auto metrics = root["metrics"];
            config.metrics.interval = std::chrono::minutes(
                GetValue<int>(metrics, "interval", 1));
        }

        // Parse task configuration
        if (root["task"]) {
            auto task = root["task"];
            config.task.max_concurrent = GetValue<size_t>(task, "max_concurrent", 4);
            config.task.queue_size = GetValue<size_t>(task, "queue_size", 100);
            config.task.default_timeout = std::chrono::minutes(
                GetValue<int>(task, "default_timeout", 5));
        }

        // Parse update configuration
        if (root["update"]) {
            auto update = root["update"];
            config.update.check_interval = std::chrono::hours(
                GetValue<int>(update, "check_interval", 1));
            config.update.idle_required = GetValue<bool>(update, "idle_required", true);
        }

        // Parse logging configuration
        if (root["logging"]) {
            auto logging = root["logging"];
            config.logging.level = GetValue<std::string>(logging, "level", "info");
            config.logging.file = GetValue<std::string>(logging, "file", "");
            config.logging.max_size = GetValue<size_t>(logging, "max_size", 100 * 1024 * 1024);
            config.logging.max_files = GetValue<size_t>(logging, "max_files", 5);
        }

        if (!config.Validate()) {
            return std::nullopt;
        }

        return config;
    } catch (const YAML::Exception& e) {
        LOG_ERROR("YAML parsing error: {}", e.what());
        return std::nullopt;
    } catch (const std::exception& e) {
        LOG_ERROR("Config parsing error: {}", e.what());
        return std::nullopt;
    }
}

bool Config::Validate() const {
    if (agent.id.empty()) {
        LOG_ERROR("Agent ID is required");
        return false;
    }

    if (agent.token.empty()) {
        LOG_ERROR("Agent token is required");
        return false;
    }

    if (agent.server_url.empty()) {
        LOG_ERROR("Server URL is required");
        return false;
    }

    if (agent.server_url.find("wss://") != 0 && agent.server_url.find("ws://") != 0) {
        LOG_ERROR("Server URL must start with ws:// or wss://");
        return false;
    }

    return true;
}

}  // namespace agent
