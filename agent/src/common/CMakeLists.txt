cmake_minimum_required(VERSION 3.20)

# Common library for shared code between executables
add_library(agent_common STATIC
    types.cpp
    websocket_client.cpp
    ipc.cpp
)

target_include_directories(agent_common PUBLIC
    ${CMAKE_SOURCE_DIR}/include
)

target_link_libraries(agent_common PUBLIC
    nlohmann_json::nlohmann_json
    spdlog::spdlog
    Boost::system
)

# Add OpenSSL for TLS support
find_package(OpenSSL REQUIRED)
target_link_libraries(agent_common PUBLIC OpenSSL::SSL OpenSSL::Crypto)

if(WIN32)
    target_link_libraries(agent_common PRIVATE
        ws2_32.lib
        crypt32.lib
    )
endif()
