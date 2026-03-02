cmake_minimum_required(VERSION 3.20)

find_package(Catch2 REQUIRED)

add_executable(agent_tests
    connection_test.cpp
    heartbeat_test.cpp
    task_test.cpp
)

target_include_directories(agent_tests PRIVATE
    ${CMAKE_SOURCE_DIR}/include
)

# Link common library
target_link_libraries(agent_tests PRIVATE
    agent_common
    Catch2::Catch2
    Catch2::Catch2WithMain
)

# Add heartbeat worker sources
target_sources(agent_tests PRIVATE
    ${CMAKE_SOURCE_DIR}/src/heartbeat/heartbeat.cpp
    ${CMAKE_SOURCE_DIR}/src/heartbeat/metrics.cpp
)

# Add task worker sources
target_sources(agent_tests PRIVATE
    ${CMAKE_SOURCE_DIR}/src/task/command_queue.cpp
    ${CMAKE_SOURCE_DIR}/src/task/command_executor.cpp
)

if(WIN32)
    target_link_libraries(agent_tests PRIVATE
        ws2_32.lib
        crypt32.lib
        pdh.lib
        psapi.lib
    )
endif()

include(CTest)
include(Catch)
catch_discover_tests(agent_tests)
