#ifndef AGENT_COMMAND_HPP
#define AGENT_COMMAND_HPP

#include <agent/common/types.hpp>
#include <chrono>
#include <cstdint>
#include <map>
#include <string>

namespace agent {

// Command structure
struct Command {
    std::string id;
    CommandType type;
    std::map<std::string, std::string> params;
    std::chrono::seconds timeout;

    // Parse from JSON string
    static std::optional<Command> FromJson(const std::string& json);
};

// Command result structure
struct CommandResult {
    std::string id;
    TaskStatus status;
    int exit_code{0};
    std::string output;
    double duration_seconds{0.0};

    // Convert to JSON string
    std::string ToJson() const;
};

// Command executor interface
class CommandExecutor {
public:
    virtual ~CommandExecutor() = default;

    virtual CommandResult Execute(const Command& command) = 0;
    virtual bool CanHandle(CommandType type) const = 0;
};

}  // namespace agent

#endif  // AGENT_COMMAND_HPP
