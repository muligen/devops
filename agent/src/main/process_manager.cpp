// Agent process manager implementation (Windows)
#include <agent/process_manager.hpp>
#include <agent/logger.hpp>

#ifdef _WIN32
#include <windows.h>
#include <tlhelp32.h>
#include <psapi.h>
#endif

namespace agent {

ProcessManager::ProcessManager(const ProcessManagerConfig& config)
    : config_(config) {
    // Initialize process info for heartbeat worker
    processes_["heartbeat"] = ProcessInfo{
        .name = "heartbeat",
        .path = config.heartbeat_worker_path,
        .pid = 0,
        .running = false
    };

    // Initialize process info for task worker
    processes_["task"] = ProcessInfo{
        .name = "task",
        .path = config.task_worker_path,
        .pid = 0,
        .running = false
    };
}

ProcessManager::~ProcessManager() {
    StopAll();
}

void ProcessManager::StartAll() {
    std::lock_guard<std::mutex> lock(mutex_);
    for (auto& [name, info] : processes_) {
        StartWorker(name);
    }
}

void ProcessManager::StopAll() {
    std::lock_guard<std::mutex> lock(mutex_);
    for (auto& [name, info] : processes_) {
        StopWorker(name);
    }
}

bool ProcessManager::StartWorker(const std::string& name) {
    auto it = processes_.find(name);
    if (it == processes_.end()) {
        LOG_ERROR("Unknown worker: {}", name);
        return false;
    }

    if (it->second.running) {
        LOG_WARN("Worker {} is already running", name);
        return true;
    }

#ifdef _WIN32
    return StartProcessWin32(name, it->second.path);
#else
    LOG_ERROR("Process management not implemented for this platform");
    return false;
#endif
}

bool ProcessManager::StopWorker(const std::string& name) {
    auto it = processes_.find(name);
    if (it == processes_.end()) {
        LOG_ERROR("Unknown worker: {}", name);
        return false;
    }

    if (!it->second.running) {
        return true;
    }

#ifdef _WIN32
    return StopProcessWin32(name);
#else
    LOG_ERROR("Process management not implemented for this platform");
    return false;
#endif
}

bool ProcessManager::RestartWorker(const std::string& name) {
    StopWorker(name);
    std::this_thread::sleep_for(config_.restart_delay);
    return StartWorker(name);
}

bool ProcessManager::IsRunning(const std::string& name) const {
    std::lock_guard<std::mutex> lock(mutex_);
    auto it = processes_.find(name);
    return it != processes_.end() && it->second.running;
}

ProcessInfo ProcessManager::GetProcessInfo(const std::string& name) const {
    std::lock_guard<std::mutex> lock(mutex_);
    auto it = processes_.find(name);
    if (it != processes_.end()) {
        return it->second;
    }
    return ProcessInfo{};
}

void ProcessManager::SetStateCallback(ProcessStateCallback callback) {
    std::lock_guard<std::mutex> lock(mutex_);
    state_callback_ = callback;
}

void ProcessManager::Monitor() {
    std::lock_guard<std::mutex> lock(mutex_);

    for (auto& [name, info] : processes_) {
#ifdef _WIN32
        auto handle_it = process_handles_.find(name);
        if (handle_it != process_handles_.end()) {
            bool was_running = info.running;
            info.running = IsProcessRunningWin32(handle_it->second);

            // If process stopped unexpectedly
            if (was_running && !info.running) {
                LOG_WARN("Worker {} stopped unexpectedly", name);

                // Close handle
                CloseHandle(handle_it->second);
                process_handles_.erase(handle_it);
                info.pid = 0;

                // Notify callback
                if (state_callback_) {
                    state_callback_(name, false);
                }
            }
        }
#endif
    }
}

void ProcessManager::WaitForAll() {
#ifdef _WIN32
    std::vector<HANDLE> handles;
    {
        std::lock_guard<std::mutex> lock(mutex_);
        for (const auto& [name, handle] : process_handles_) {
            handles.push_back(handle);
        }
    }

    if (!handles.empty()) {
        WaitForMultipleObjects(static_cast<DWORD>(handles.size()),
                               handles.data(), TRUE, INFINITE);
    }
#endif
}

#ifdef _WIN32
bool ProcessManager::StartProcessWin32(const std::string& name, const std::string& path) {
    STARTUPINFOA si;
    PROCESS_INFORMATION pi;

    ZeroMemory(&si, sizeof(si));
    si.cb = sizeof(si);
    ZeroMemory(&pi, sizeof(pi));

    // Create mutable copy of path for CreateProcess
    std::string cmd_line = path;
    char* cmd_line_ptr = cmd_line.empty() ? nullptr : &cmd_line[0];

    // Create the process
    if (!CreateProcessA(
            nullptr,                   // No module name
            cmd_line_ptr,              // Command line
            nullptr,                   // Process handle not inheritable
            nullptr,                   // Thread handle not inheritable
            FALSE,                      // No handle inheritance
            CREATE_NO_WINDOW,          // Creation flags (no console window)
            nullptr,                   // Use parent's environment block
            nullptr,                   // Use parent's starting directory
            &si,                        // Pointer to STARTUPINFO structure
            &pi                         // Pointer to PROCESS_INFORMATION structure
        )) {
        LOG_ERROR("Failed to create process {}: {}", name, GetLastError());
        return false;
    }

    // Update process info
    auto& info = processes_[name];
    info.pid = static_cast<int>(pi.dwProcessId);
    info.running = true;
    info.start_time = std::chrono::system_clock::now();

    // Store handle
    process_handles_[name] = pi.hProcess;

    // Close thread handle (we don't need it)
    CloseHandle(pi.hThread);

    LOG_INFO("Started worker {} with PID {}", name, info.pid);

    // Notify callback
    if (state_callback_) {
        state_callback_(name, true);
    }

    return true;
}

bool ProcessManager::StopProcessWin32(const std::string& name) {
    auto handle_it = process_handles_.find(name);
    if (handle_it == process_handles_.end()) {
        return true;  // Already stopped
    }

    HANDLE handle = handle_it->second;

    // Try graceful shutdown first
    if (!TerminateProcess(handle, 0)) {
        LOG_WARN("Failed to terminate process {}: {}", name, GetLastError());
    }

    // Wait for process to exit
    WaitForSingleObject(handle, 5000);

    // Close handle
    CloseHandle(handle);
    process_handles_.erase(handle_it);

    // Update process info
    auto& info = processes_[name];
    info.running = false;
    info.pid = 0;

    LOG_INFO("Stopped worker {}", name);

    // Notify callback
    if (state_callback_) {
        state_callback_(name, false);
    }

    return true;
}

bool ProcessManager::IsProcessRunningWin32(HANDLE handle) const {
    if (handle == nullptr || handle == INVALID_HANDLE_VALUE) {
        return false;
    }

    DWORD exit_code;
    if (!GetExitCodeProcess(handle, &exit_code)) {
        return false;
    }

    return exit_code == STILL_ACTIVE;
}
#endif

}  // namespace agent
