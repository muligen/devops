// Command executor implementation
#include <agent/command.hpp>
#include <nlohmann/json.hpp>

namespace agent {

std::optional<Command> Command::FromJson(const std::string& json) {
    try {
        auto j = nlohmann::json::parse(json);

        Command cmd;
        cmd.id = j.value("id", "");

        std::string type_str = j.value("command_type", j.value("type", ""));
        auto type_opt = StringToCommandType(type_str);
        if (!type_opt) {
            return std::nullopt;
        }
        cmd.type = *type_opt;

        // Parse params
        if (j.contains("params") && j["params"].is_object()) {
            for (auto& [key, value] : j["params"].items()) {
                if (value.is_string()) {
                    cmd.params[key] = value.get<std::string>();
                } else {
                    cmd.params[key] = value.dump();
                }
            }
        }

        // Parse timeout (default 5 minutes)
        int timeout_seconds = j.value("timeout", 300);
        cmd.timeout = std::chrono::seconds(timeout_seconds);

        return cmd;
    } catch (const std::exception&) {
        return std::nullopt;
    }
}

std::string CommandResult::ToJson() const {
    nlohmann::json j;
    j["type"] = "result";
    j["id"] = id;
    j["status"] = TaskStatusToString(status);
    j["exit_code"] = exit_code;
    j["output"] = output;
    j["duration"] = duration_seconds;

    return j.dump();
}

}  // namespace agent
