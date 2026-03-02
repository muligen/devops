// Agent heartbeat module unit tests
#include <agent/heartbeat_worker.hpp>
#include <agent/common/types.hpp>

#include <catch2/catch_test_macros.hpp>

using namespace agent;

TEST_CASE("SystemMetrics defaults", "[heartbeat]") {
    SystemMetrics metrics;

    REQUIRE(metrics.cpu_usage_percent == 0.0);
    REQUIRE(metrics.memory.total_bytes == 0);
    REQUIRE(metrics.memory.used_bytes == 0);
    REQUIRE(metrics.disk.total_bytes == 0);
    REQUIRE(metrics.uptime_seconds == 0);
}

TEST_CASE("HeartbeatWorker construction", "[heartbeat]") {
    HeartbeatWorker worker;

    // Worker should be constructable without errors
    REQUIRE(true);
}

TEST_CASE("Heartbeat intervals", "[heartbeat]") {
    // Default intervals are defined in types.hpp
    REQUIRE(kDefaultHeartbeatInterval.count() == 1);
    REQUIRE(kDefaultMetricsInterval.count() == 1);  // 1 minute
}
