cmake_minimum_required(VERSION 3.20)

add_executable(worker_task
    main.cpp
    command_executor.cpp
    command_queue.cpp
    thread_pool.cpp
    websocket_client.cpp
    commands/exec_shell.cpp
    commands/init_machine.cpp
    commands/clean_disk.cpp
)

target_include_directories(worker_task PRIVATE
    ${CMAKE_SOURCE_DIR}/include
)

target_link_libraries(worker_task PRIVATE
    agent_common
    nlohmann_json::nlohmann_json
    yaml-cpp::yaml-cpp
    spdlog::spdlog
    Boost::system
)

if(WIN32)
    target_link_libraries(worker_task PRIVATE
        shell32.lib
        advapi32.lib
        ole32.lib
    )
endif()

set_target_properties(worker_task PROPERTIES
    OUTPUT_NAME "worker_task"
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin"
)
