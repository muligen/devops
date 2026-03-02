// Agent common types implementation
#include <agent/common/types.hpp>
#include <unordered_map>

namespace agent {

std::string CommandTypeToString(CommandType type) {
    static const std::unordered_map<CommandType, std::string> kCommandTypeNames = {
        {CommandType::kExecShell, "exec_shell"},
        {CommandType::kInitMachine, "init_machine"},
        {CommandType::kCleanDisk, "clean_disk"}
    };

    auto it = kCommandTypeNames.find(type);
    if (it != kCommandTypeNames.end()) {
        return it->second;
    }
    return "unknown";
}

std::optional<CommandType> StringToCommandType(const std::string& str) {
    static const std::unordered_map<std::string, CommandType> kStringToCommandType = {
        {"exec_shell", CommandType::kExecShell},
        {"init_machine", CommandType::kInitMachine},
        {"clean_disk", CommandType::kCleanDisk}
    };

    auto it = kStringToCommandType.find(str);
    if (it != kStringToCommandType.end()) {
        return it->second;
    }
    return std::nullopt;
}

std::string TaskStatusToString(TaskStatus status) {
    static const std::unordered_map<TaskStatus, std::string> kTaskStatusNames = {
        {TaskStatus::kPending, "pending"},
        {TaskStatus::kRunning, "running"},
        {TaskStatus::kSuccess, "success"},
        {TaskStatus::kFailed, "failed"},
        {TaskStatus::kTimeout, "timeout"},
        {TaskStatus::kCancelled, "cancelled"}
    };

    auto it = kTaskStatusNames.find(status);
    if (it != kTaskStatusNames.end()) {
        return it->second;
    }
    return "unknown";
}

std::optional<TaskStatus> StringToTaskStatus(const std::string& str) {
    static const std::unordered_map<std::string, TaskStatus> kStringToTaskStatus = {
        {"pending", TaskStatus::kPending},
        {"running", TaskStatus::kRunning},
        {"success", TaskStatus::kSuccess},
        {"failed", TaskStatus::kFailed},
        {"timeout", TaskStatus::kTimeout},
        {"cancelled", TaskStatus::kCancelled}
    };

    auto it = kStringToTaskStatus.find(str);
    if (it != kStringToTaskStatus.end()) {
        return it->second;
    }
    return std::nullopt;
}

}  // namespace agent
