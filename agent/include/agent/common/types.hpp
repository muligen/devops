#ifndef AGENT_COMMON_TYPES_HPP
#define AGENT_COMMON_TYPES_HPP

#include <chrono>
#include <cstdint>
#include <optional>
#include <string>

namespace agent {

// Version information
constexpr const char* kAgentVersion = "1.0.0";
constexpr int kVersionMajor = 1;
constexpr int kVersionMinor = 0;
constexpr int kVersionPatch = 0;

// Default configuration values
constexpr auto kDefaultHeartbeatInterval = std::chrono::seconds(1);
constexpr auto kDefaultMetricsInterval = std::chrono::minutes(1);
constexpr auto kDefaultConnectTimeout = std::chrono::seconds(30);
constexpr auto kDefaultReconnectBaseDelay = std::chrono::seconds(5);
constexpr auto kDefaultReconnectMaxDelay = std::chrono::seconds(60);
constexpr auto kDefaultCommandTimeout = std::chrono::minutes(5);
constexpr size_t kDefaultMaxConcurrentCommands = 4;
constexpr size_t kDefaultCommandQueueSize = 100;

// Connection states
enum class ConnectionState {
    kDisconnected,
    kConnecting,
    kConnected,
    kAuthenticating,
    kAuthenticated
};

// Task status
enum class TaskStatus {
    kPending,
    kRunning,
    kSuccess,
    kFailed,
    kTimeout,
    kCancelled
};

// Command types
enum class CommandType {
    kExecShell,
    kInitMachine,
    kCleanDisk
};

// String to enum conversion helpers
std::string CommandTypeToString(CommandType type);
std::optional<CommandType> StringToCommandType(const std::string& str);
std::string TaskStatusToString(TaskStatus status);
std::optional<TaskStatus> StringToTaskStatus(const std::string& str);

}  // namespace agent

#endif  // AGENT_COMMON_TYPES_HPP
