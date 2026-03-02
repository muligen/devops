// Agent task execution module unit tests
#include <agent/command.hpp>
#include <agent/command_queue.hpp>
#include <agent/common/types.hpp>

#include <catch2/catch_test_macros.hpp>

using namespace agent;

TEST_CASE("CommandQueue basic operations", "[task]") {
    CommandQueue queue(10);

    SECTION("Empty queue") {
        REQUIRE(queue.Empty());
        REQUIRE(queue.Size() == 0);
    }

    SECTION("Push and pop") {
        Command cmd;
        cmd.id = "test-cmd-1";
        cmd.type = CommandType::kExecShell;
        cmd.timeout = std::chrono::seconds(60);

        REQUIRE(queue.Push(cmd));
        REQUIRE_FALSE(queue.Empty());
        REQUIRE(queue.Size() == 1);

        auto popped = queue.Pop();
        REQUIRE(popped.has_value());
        REQUIRE(popped->id == "test-cmd-1");
        REQUIRE(queue.Empty());
    }

    SECTION("Queue full") {
        CommandQueue small_queue(2);
        Command cmd;
        cmd.id = "cmd-1";

        REQUIRE(small_queue.Push(cmd));
        REQUIRE(small_queue.Push(cmd));
        REQUIRE_FALSE(small_queue.Push(cmd));  // Should fail, queue full
    }
}

TEST_CASE("Command JSON parsing", "[task]") {
    SECTION("Valid command") {
        std::string json = R"({
            "id": "cmd-123",
            "type": "exec_shell",
            "params": {
                "command": "dir C:\\",
                "shell": "cmd.exe"
            },
            "timeout": 300
        })";

        auto cmd = Command::FromJson(json);
        REQUIRE(cmd.has_value());
        REQUIRE(cmd->id == "cmd-123");
        REQUIRE(cmd->type == CommandType::kExecShell);
        REQUIRE(cmd->timeout.count() == 300);
    }

    SECTION("Invalid JSON") {
        std::string json = "{ invalid json }";
        auto cmd = Command::FromJson(json);
        REQUIRE_FALSE(cmd.has_value());
    }

    SECTION("Missing required fields") {
        std::string json = R"({"id": "cmd-123"})";
        auto cmd = Command::FromJson(json);
        // Should fail due to missing type
        REQUIRE_FALSE(cmd.has_value());
    }
}

TEST_CASE("CommandResult JSON serialization", "[task]") {
    CommandResult result;
    result.id = "cmd-123";
    result.status = TaskStatus::kSuccess;
    result.exit_code = 0;
    result.output = "Command completed successfully";
    result.duration_seconds = 1.5;

    std::string json = result.ToJson();

    REQUIRE(json.find("\"id\":\"cmd-123\"") != std::string::npos);
    REQUIRE(json.find("\"status\":\"success\"") != std::string::npos);
    REQUIRE(json.find("\"exit_code\":0") != std::string::npos);
}
