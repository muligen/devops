// Agent logger implementation
#include <agent/logger.hpp>
#include <spdlog/sinks/rotating_file_sink.h>
#include <spdlog/sinks/stdout_color_sinks.h>
#include <spdlog/spdlog.h>

namespace agent {

std::shared_ptr<spdlog::logger> Logger::logger_;

void Logger::Init(const LoggerConfig& config) {
    std::vector<spdlog::sink_ptr> sinks;

    // Console sink
    if (config.console_output) {
        auto console_sink = std::make_shared<spdlog::sinks::stdout_color_sink_mt>();
        console_sink->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [%^%l%$] [%t] %v");
        sinks.push_back(console_sink);
    }

    // File sink with rotation
    if (!config.file.empty()) {
        auto file_sink = std::make_shared<spdlog::sinks::rotating_file_sink_mt>(
            config.file, config.max_size, config.max_files);
        file_sink->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [%l] [%t] %v");
        sinks.push_back(file_sink);
    }

    // Create logger
    if (sinks.empty()) {
        // Default to console if no sinks configured
        auto console_sink = std::make_shared<spdlog::sinks::stdout_color_sink_mt>();
        console_sink->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [%^%l%$] [%t] %v");
        sinks.push_back(console_sink);
    }

    logger_ = std::make_shared<spdlog::logger>(config.name, sinks.begin(), sinks.end());
    logger_->set_level(spdlog::level::from_str(config.level));
    logger_->flush_on(spdlog::level::warn);

    // Register as default logger
    spdlog::register_logger(logger_);
    spdlog::set_default_logger(logger_);
}

std::shared_ptr<spdlog::logger> Logger::Get() {
    if (!logger_) {
        // Initialize with defaults if not already done
        LoggerConfig default_config;
        Init(default_config);
    }
    return logger_;
}

void Logger::SetLevel(LogLevel level) {
    if (logger_) {
        logger_->set_level(static_cast<spdlog::level::level_enum>(level));
    }
}

void Logger::SetLevel(const std::string& level) {
    if (logger_) {
        logger_->set_level(spdlog::level::from_str(level));
    }
}

void Logger::Flush() {
    if (logger_) {
        logger_->flush();
    }
}

void Logger::Shutdown() {
    if (logger_) {
        logger_->flush();
        spdlog::drop(logger_->name());
        logger_.reset();
    }
}

}  // namespace agent
