// Heartbeat worker implementation
#include <agent/heartbeat_worker.hpp>

#include <chrono>
#include <thread>

namespace agent {

namespace {
    constexpr auto kHeartbeatInterval = std::chrono::seconds(1);
    constexpr auto kMetricsInterval = std::chrono::minutes(1);
}

HeartbeatWorker::HeartbeatWorker()
    : last_heartbeat_(std::chrono::steady_clock::now())
    , last_metrics_(std::chrono::steady_clock::now()) {
}

HeartbeatWorker::~HeartbeatWorker() {
    Stop();
}

void HeartbeatWorker::Run() {
    running_ = true;

    while (running_) {
        auto now = std::chrono::steady_clock::now();

        // Send heartbeat every second
        if (now - last_heartbeat_ >= kHeartbeatInterval) {
            SendHeartbeat();
            last_heartbeat_ = now;
        }

        // Collect and send metrics every minute
        if (now - last_metrics_ >= kMetricsInterval) {
            CollectAndSendMetrics();
            last_metrics_ = now;
        }

        // Sleep for a short interval to avoid busy-waiting
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }
}

void HeartbeatWorker::Stop() {
    running_ = false;
}

void HeartbeatWorker::SendHeartbeat() {
    // TODO: Send heartbeat to main process via IPC
    // This will be implemented in the IPC module
}

void HeartbeatWorker::CollectAndSendMetrics() {
    SystemMetrics metrics;

    // Collect all metrics
    metrics.cpu_usage_percent = GetCpuUsage();
    metrics.memory = GetMemoryInfo();
    metrics.disk = GetDiskInfo();
    metrics.uptime_seconds = GetSystemUptime();

    // TODO: Send metrics to main process via IPC
}

}  // namespace agent
