// exec_shell command implementation
#include <agent/command.hpp>
#include <chrono>
#include <thread>
#include <sstream>

#ifdef _WIN32
#include <windows.h>
#endif

namespace agent {

class ExecShellCommand : public CommandExecutor {
public:
    CommandResult Execute(const Command& command) override {
        CommandResult result;
        result.id = command.id;

        auto start_time = std::chrono::steady_clock::now();

#ifdef _WIN32
        // Get shell and command from params
        std::string shell = command.params.count("shell") ? command.params.at("shell") : "cmd.exe";
        std::string cmd = command.params.count("command") ? command.params.at("command") : "";

        // Build command line
        std::string full_cmd;
        if (shell.find("powershell") != std::string::npos || shell.find("pwsh") != std::string::npos) {
            full_cmd = shell + " -Command \"" + cmd + "\"";
        } else {
            full_cmd = shell + " /c " + cmd;
        }

        // Set up security attributes for pipes
        SECURITY_ATTRIBUTES sa;
        sa.nLength = sizeof(SECURITY_ATTRIBUTES);
        sa.bInheritHandle = TRUE;
        sa.lpSecurityDescriptor = nullptr;

        // Create pipes for stdout and stderr
        HANDLE stdout_read = nullptr;
        HANDLE stdout_write = nullptr;
        HANDLE stderr_read = nullptr;
        HANDLE stderr_write = nullptr;

        if (!CreatePipe(&stdout_read, &stdout_write, &sa, 0) ||
            !CreatePipe(&stderr_read, &stderr_write, &sa, 0)) {
            result.status = TaskStatus::kFailed;
            result.output = "Failed to create pipes";
            result.exit_code = -1;
            return result;
        }

        // Make read handles non-inheritable
        SetHandleInformation(stdout_read, HANDLE_FLAG_INHERIT, 0);
        SetHandleInformation(stderr_read, HANDLE_FLAG_INHERIT, 0);

        // Set up startup info
        STARTUPINFOA si = {};
        si.cb = sizeof(STARTUPINFOA);
        si.dwFlags = STARTF_USESTDHANDLES;
        si.hStdOutput = stdout_write;
        si.hStdError = stderr_write;
        si.hStdInput = GetStdHandle(STD_INPUT_HANDLE);

        PROCESS_INFORMATION pi = {};

        // Create the process
        std::vector<char> cmd_buf(full_cmd.size() + 1);
        strcpy_s(cmd_buf.data(), cmd_buf.size(), full_cmd.c_str());

        BOOL success = CreateProcessA(
            nullptr,
            cmd_buf.data(),
            nullptr,
            nullptr,
            TRUE,
            CREATE_NO_WINDOW,
            nullptr,
            nullptr,
            &si,
            &pi
        );

        if (!success) {
            CloseHandle(stdout_read);
            CloseHandle(stderr_read);
            CloseHandle(stdout_write);
            CloseHandle(stderr_write);

            result.status = TaskStatus::kFailed;
            result.output = "Failed to create process: " + std::to_string(GetLastError());
            result.exit_code = -1;
            return result;
        }

        // Close write ends of pipes
        CloseHandle(stdout_write);
        CloseHandle(stderr_write);

        // Read output with timeout
        std::string stdout_output;
        std::string stderr_output;
        auto deadline = start_time + command.timeout;

        char buffer[4096];
        DWORD bytes_read;

        while (true) {
            // Check timeout
            if (std::chrono::steady_clock::now() > deadline) {
                TerminateProcess(pi.hProcess, 1);
                result.status = TaskStatus::kTimeout;
                break;
            }

            bool has_data = false;

            // Check if process has exited
            DWORD exit_code;
            if (GetExitCodeProcess(pi.hProcess, &exit_code) && exit_code != STILL_ACTIVE) {
                has_data = true;
            }

            // Read from stdout
            DWORD available = 0;
            if (PeekNamedPipe(stdout_read, nullptr, 0, nullptr, &available, nullptr) && available > 0) {
                if (ReadFile(stdout_read, buffer, sizeof(buffer) - 1, &bytes_read, nullptr) && bytes_read > 0) {
                    buffer[bytes_read] = '\0';
                    stdout_output.append(buffer, bytes_read);
                    has_data = true;
                }
            }

            // Read from stderr
            if (PeekNamedPipe(stderr_read, nullptr, 0, nullptr, &available, nullptr) && available > 0) {
                if (ReadFile(stderr_read, buffer, sizeof(buffer) - 1, &bytes_read, nullptr) && bytes_read > 0) {
                    buffer[bytes_read] = '\0';
                    stderr_output.append(buffer, bytes_read);
                    has_data = true;
                }
            }

            // Check if process has exited
            if (GetExitCodeProcess(pi.hProcess, &exit_code) && exit_code != STILL_ACTIVE) {
                // Read remaining output
                while (PeekNamedPipe(stdout_read, nullptr, 0, nullptr, &available, nullptr) && available > 0) {
                    if (ReadFile(stdout_read, buffer, sizeof(buffer) - 1, &bytes_read, nullptr) && bytes_read > 0) {
                        buffer[bytes_read] = '\0';
                        stdout_output.append(buffer, bytes_read);
                    } else {
                        break;
                    }
                }
                while (PeekNamedPipe(stderr_read, nullptr, 0, nullptr, &available, nullptr) && available > 0) {
                    if (ReadFile(stderr_read, buffer, sizeof(buffer) - 1, &bytes_read, nullptr) && bytes_read > 0) {
                        buffer[bytes_read] = '\0';
                        stderr_output.append(buffer, bytes_read);
                    } else {
                        break;
                    }
                }

                result.exit_code = static_cast<int>(exit_code);
                result.status = (exit_code == 0) ? TaskStatus::kSuccess : TaskStatus::kFailed;
                break;
            }

            if (!has_data) {
                std::this_thread::sleep_for(std::chrono::milliseconds(10));
            }
        }

        // Cleanup handles
        CloseHandle(stdout_read);
        CloseHandle(stderr_read);
        CloseHandle(pi.hProcess);
        CloseHandle(pi.hThread);

        // Combine stdout and stderr
        if (!stdout_output.empty() && !stderr_output.empty()) {
            result.output = stdout_output + "\n[stderr]\n" + stderr_output;
        } else if (!stdout_output.empty()) {
            result.output = stdout_output;
        } else if (!stderr_output.empty()) {
            result.output = stderr_output;
        }

#else
        // Non-Windows stub
        result.status = TaskStatus::kFailed;
        result.output = "exec_shell not supported on this platform";
        result.exit_code = -1;
#endif

        auto end_time = std::chrono::steady_clock::now();
        result.duration_seconds = std::chrono::duration<double>(end_time - start_time).count();

        return result;
    }

    bool CanHandle(CommandType type) const override {
        return type == CommandType::kExecShell;
    }
};

// Factory function
std::unique_ptr<CommandExecutor> CreateExecShellExecutor() {
    return std::make_unique<ExecShellCommand>();
}

}  // namespace agent
