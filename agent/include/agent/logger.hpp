#ifndef AGENT_LOGGER_HPP
#define AGENT_LOGGER_HPP

#include <memory>
#include <spdlog/spdlog.h>
#include <string>

namespace agent {

// Logger levels
enum class LogLevel {
    kTrace = SPDLOG_LEVEL_TRACE,
    kDebug = SPDLOG_LEVEL_DEBUG,
    kInfo = SPDLOG_LEVEL_INFO,
    kWarn = SPDLOG_LEVEL_WARN,
    kError = SPDLOG_LEVEL_ERROR,
    kCritical = SPDLOG_LEVEL_CRITICAL,
    kOff = SPDLOG_LEVEL_OFF
};

// Logger configuration
struct LoggerConfig {
    std::string name{"agent"};
    std::string level{"info"};
    std::string file;
    size_t max_size{100 * 1024 * 1024};  // 100MB
    size_t max_files{5};
    bool console_output{true};
};

// Logger wrapper class
class Logger {
public:
    // Initialize logger with configuration
    static void Init(const LoggerConfig& config);

    // Get the underlying spdlog logger
    static std::shared_ptr<spdlog::logger> Get();

    // Set log level
    static void SetLevel(LogLevel level);
    static void SetLevel(const std::string& level);

    // Flush logs
    static void Flush();

    // Shutdown logger
    static void Shutdown();

private:
    Logger() = default;
    static std::shared_ptr<spdlog::logger> logger_;
};

// Convenience macros for logging
#define LOG_TRACE(...) SPDLOG_LOGGER_TRACE(agent::Logger::Get(), __VA_ARGS__)
#define LOG_DEBUG(...) SPDLOG_LOGGER_DEBUG(agent::Logger::Get(), __VA_ARGS__)
#define LOG_INFO(...) SPDLOG_LOGGER_INFO(agent::Logger::Get(), __VA_ARGS__)
#define LOG_WARN(...) SPDLOG_LOGGER_WARN(agent::Logger::Get(), __VA_ARGS__)
#define LOG_ERROR(...) SPDLOG_LOGGER_ERROR(agent::Logger::Get(), __VA_ARGS__)
#define LOG_CRITICAL(...) SPDLOG_LOGGER_CRITICAL(agent::Logger::Get(), __VA_ARGS__)

}  // namespace agent

#endif  // AGENT_LOGGER_HPP
