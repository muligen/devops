# Agent Test CMakeLists.txt

# Include directories for tests
include_directories(${CMAKE_SOURCE_DIR}/include)
include_directories(${CMAKE_SOURCE_DIR}/src)

# Connection tests
add_executable(connection_test connection_test.cpp)
target_link_libraries(connection_test
    PRIVATE
    agent_common
    GTest::gtest
    GTest::gtest_main
    nlohmann_json::nlohmann_json
    spdlog::spdlog
    OpenSSL::SSL
    OpenSSL::Crypto
)
add_test(NAME ConnectionTest COMMAND connection_test)

# Task tests
add_executable(task_test task_test.cpp)
target_link_libraries(task_test
    PRIVATE
    agent_common
    GTest::gtest
    GTest::gtest_main
    nlohmann_json::nlohmann_json
    spdlog::spdlog
    yaml-cpp::yaml-cpp
)
add_test(NAME TaskTest COMMAND task_test)
