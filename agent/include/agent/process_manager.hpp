#ifndef AGENT_PROCESS_MANAGER_HPP
#define AGENT_PROCESS_MANAGER_HPP

#include <chrono>
#include <functional>
#include <map>
#include <memory>
#include <mutex>
#include <string>

#ifdef _WIN32
#include <windows.h>
#endif

namespace agent {

// Process information
struct ProcessInfo {
    std::string name;
    std::string path;
    int pid{0};
    bool running{false};
    std::chrono::system_clock::time_point start_time;
};

// Process manager configuration
struct ProcessManagerConfig {
    std::string heartbeat_worker_path;
    std::string task_worker_path;
    std::chrono::seconds restart_delay{5};
    int max_restart_attempts{3};
};

// Process state callback
using ProcessStateCallback = std::function<void(const std::string& name, bool running)>;

// Process manager - handles worker process lifecycle
class ProcessManager {
public:
    explicit ProcessManager(const ProcessManagerConfig& config);
    ~ProcessManager();

    // Start all worker processes
    void StartAll();

    // Stop all worker processes
    void StopAll();

    // Start a specific worker
    bool StartWorker(const std::string& name);

    // Stop a specific worker
    bool StopWorker(const std::string& name);

    // Restart a specific worker
    bool RestartWorker(const std::string& name);

    // Check if a worker is running
    bool IsRunning(const std::string& name) const;

    // Get process information
    ProcessInfo GetProcessInfo(const std::string& name) const;

    // Set state change callback
    void SetStateCallback(ProcessStateCallback callback);

    // Monitor processes (call periodically)
    void Monitor();

    // Wait for all processes to exit
    void WaitForAll();

private:
    // Platform-specific implementation
#ifdef _WIN32
    bool StartProcessWin32(const std::string& name, const std::string& path);
    bool StopProcessWin32(const std::string& name);
    bool IsProcessRunningWin32(HANDLE handle) const;
#endif

    ProcessManagerConfig config_;
    mutable std::mutex mutex_;
    std::map<std::string, ProcessInfo> processes_;
    ProcessStateCallback state_callback_;

#ifdef _WIN32
    std::map<std::string, HANDLE> process_handles_;
#endif
};

}  // namespace agent

#endif  // AGENT_PROCESS_MANAGER_HPP
