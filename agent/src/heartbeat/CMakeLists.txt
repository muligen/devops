cmake_minimum_required(VERSION 3.20)

add_executable(worker_heartbeat
    main.cpp
    heartbeat.cpp
    metrics.cpp
)

target_include_directories(worker_heartbeat PRIVATE
    ${CMAKE_SOURCE_DIR}/include
)

target_link_libraries(worker_heartbeat PRIVATE
    agent_common
    nlohmann_json::nlohmann_json
    spdlog::spdlog
)

if(WIN32)
    target_link_libraries(worker_heartbeat PRIVATE
        pdh.lib
        psapi.lib
    )
endif()

set_target_properties(worker_heartbeat PROPERTIES
    OUTPUT_NAME "worker_heartbeat"
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin"
)
