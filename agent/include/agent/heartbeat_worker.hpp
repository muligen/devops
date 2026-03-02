#ifndef AGENT_HEARTBEAT_WORKER_HPP
#define AGENT_HEARTBEAT_WORKER_HPP

#include <chrono>
#include <cstdint>
#include <string>

namespace agent {

// System metrics structure
struct SystemMetrics {
    double cpu_usage_percent{0.0};

    struct MemoryInfo {
        uint64_t total_bytes{0};
        uint64_t used_bytes{0};
        uint64_t available_bytes{0};
        double percent{0.0};
    } memory;

    struct DiskInfo {
        uint64_t total_bytes{0};
        uint64_t used_bytes{0};
        uint64_t free_bytes{0};
        double percent{0.0};
    } disk;

    uint64_t uptime_seconds{0};
    std::string timestamp;
};

// Heartbeat worker class
class HeartbeatWorker {
public:
    HeartbeatWorker();
    ~HeartbeatWorker();

    // Start the worker loop
    void Run();

    // Stop the worker
    void Stop();

private:
    void SendHeartbeat();
    void CollectAndSendMetrics();

    // Platform-specific metric collection
    double GetCpuUsage();
    SystemMetrics::MemoryInfo GetMemoryInfo();
    SystemMetrics::DiskInfo GetDiskInfo();
    uint64_t GetSystemUptime();

    bool running_{false};
    std::chrono::steady_clock::time_point last_heartbeat_;
    std::chrono::steady_clock::time_point last_metrics_;
};

}  // namespace agent

#endif  // AGENT_HEARTBEAT_WORKER_HPP
