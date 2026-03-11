cmake_minimum_required(VERSION 3.20)

add_executable(agent
    main.cpp
    config.cpp
    logger.cpp
    process_manager.cpp
    service.cpp
    updater.cpp
)

# Add command executor sources
target_sources(agent PRIVATE
    ../task/command_executor.cpp
    ../task/command_queue.cpp
    ../task/thread_pool.cpp
    ../task/commands/exec_shell.cpp
    ../task/commands/init_machine.cpp
    ../task/commands/clean_disk.cpp
)

target_link_libraries(agent PRIVATE
    agent_common
    nlohmann_json::nlohmann_json
    yaml-cpp::yaml-cpp
    spdlog::spdlog
    Boost::system
)

if(WIN32)
    target_link_libraries(agent PRIVATE
        advapi32.lib
        user32.lib
        shell32.lib
        ole32.lib
        pdh.lib
        psapi.lib
    )
endif()

set_target_properties(agent PROPERTIES
    OUTPUT_NAME "agent"
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin"
)
