// Windows system metrics collection
#include <agent/heartbeat_worker.hpp>

#ifdef _WIN32
#include <windows.h>
#include <pdh.h>
#include <psapi.h>
#include <cstdint>

namespace agent {

namespace {
    // CPU usage calculation state
    static ULONGLONG last_cpu_time_ = 0;
    static ULONGLONG last_sys_time_ = 0;
    static bool cpu_initialized_ = false;
}

double HeartbeatWorker::GetCpuUsage() {
    // Get system CPU time
    FILETIME sys_idle, sys_kernel, sys_user;
    if (!GetSystemTimes(&sys_idle, &sys_kernel, &sys_user)) {
        return 0.0;
    }

    // Calculate total system time
    ULONGLONG kernel_time = (static_cast<ULONGLONG>(sys_kernel.dwHighDateTime) << 32) |
                            sys_kernel.dwLowDateTime;
    ULONGLONG user_time = (static_cast<ULONGLONG>(sys_user.dwHighDateTime) << 32) |
                          sys_user.dwLowDateTime;
    ULONGLONG idle_time = (static_cast<ULONGLONG>(sys_idle.dwHighDateTime) << 32) |
                          sys_idle.dwLowDateTime;

    ULONGLONG total_time = kernel_time + user_time;

    if (!cpu_initialized_) {
        last_cpu_time_ = total_time;
        last_sys_time_ = idle_time;
        cpu_initialized_ = true;
        return 0.0;
    }

    ULONGLONG cpu_delta = total_time - last_cpu_time_;
    ULONGLONG idle_delta = idle_time - last_sys_time_;

    last_cpu_time_ = total_time;
    last_sys_time_ = idle_time;

    if (cpu_delta == 0) {
        return 0.0;
    }

    double usage = 100.0 * (1.0 - static_cast<double>(idle_delta) / static_cast<double>(cpu_delta));
    return (std::max)(0.0, (std::min)(100.0, usage));
}

SystemMetrics::MemoryInfo HeartbeatWorker::GetMemoryInfo() {
    SystemMetrics::MemoryInfo info;

    MEMORYSTATUSEX status;
    status.dwLength = sizeof(status);

    if (GlobalMemoryStatusEx(&status)) {
        info.total_bytes = status.ullTotalPhys;
        info.available_bytes = status.ullAvailPhys;
        info.used_bytes = info.total_bytes - info.available_bytes;
        info.percent = static_cast<double>(status.dwMemoryLoad);
    }

    return info;
}

SystemMetrics::DiskInfo HeartbeatWorker::GetDiskInfo() {
    SystemMetrics::DiskInfo info;

    // Get disk space for C: drive
    ULARGE_INTEGER free_bytes_available;
    ULARGE_INTEGER total_bytes;
    ULARGE_INTEGER free_bytes;

    if (GetDiskFreeSpaceExW(L"C:\\", &free_bytes_available, &total_bytes, &free_bytes)) {
        info.total_bytes = total_bytes.QuadPart;
        info.free_bytes = free_bytes.QuadPart;
        info.used_bytes = info.total_bytes - info.free_bytes;
        if (info.total_bytes > 0) {
            info.percent = 100.0 * static_cast<double>(info.used_bytes) /
                          static_cast<double>(info.total_bytes);
        }
    }

    return info;
}

uint64_t HeartbeatWorker::GetSystemUptime() {
    return static_cast<uint64_t>(GetTickCount64() / 1000);
}

}  // namespace agent

#else
// Non-Windows stub implementations
namespace agent {

double HeartbeatWorker::GetCpuUsage() {
    return 0.0;
}

SystemMetrics::MemoryInfo HeartbeatWorker::GetMemoryInfo() {
    return SystemMetrics::MemoryInfo{};
}

SystemMetrics::DiskInfo HeartbeatWorker::GetDiskInfo() {
    return SystemMetrics::DiskInfo{};
}

uint64_t HeartbeatWorker::GetSystemUptime() {
    return 0;
}

}  // namespace agent
#endif
