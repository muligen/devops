// Agent connection module unit tests
#include <agent/websocket_client.hpp>
#include <agent/common/types.hpp>

#include <catch2/catch_test_macros.hpp>

using namespace agent;

TEST_CASE("WebSocketConfig defaults", "[connection]") {
    WebSocketConfig config;

    REQUIRE(config.connect_timeout.count() == 30);
    REQUIRE(config.reconnect_base_delay.count() == 5);
    REQUIRE(config.reconnect_max_delay.count() == 60);
}

TEST_CASE("ConnectionState transitions", "[connection]") {
    SECTION("Initial state is disconnected") {
        WebSocketConfig config;
        auto client = WebSocketClient::Create(config);
        REQUIRE(client->GetState() == ConnectionState::kDisconnected);
    }
}

TEST_CASE("URL parsing", "[connection]") {
    // URL parsing is internal, tested through Connect behavior
}
