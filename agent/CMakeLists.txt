cmake_minimum_required(VERSION 3.20)
project(AgentTeams VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Conan settings
set(CMAKE_MODULE_PATH ${CMAKE_BINARY_DIR} ${CMAKE_MODULE_PATH})
set(CMAKE_PREFIX_PATH ${CMAKE_BINARY_DIR} ${CMAKE_PREFIX_PATH})

# Output directories
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)

# Compiler warnings
if(MSVC)
    add_compile_options(/W4 /WX- /utf-8)
    add_compile_definitions(_CRT_SECURE_NO_WARNINGS)
else()
    add_compile_options(-Wall -Wextra -Wpedantic)
endif()

# Find packages from Conan
find_package(nlohmann_json REQUIRED)
find_package(yaml-cpp REQUIRED)
find_package(spdlog REQUIRED)
find_package(Boost COMPONENTS system REQUIRED)
find_package(OpenSSL REQUIRED)

# Common include directory
include_directories(${CMAKE_SOURCE_DIR}/include)

# Common library (shared between executables)
add_subdirectory(src/common)

# Main agent executable
add_subdirectory(src/main)

# Heartbeat worker executable
add_subdirectory(src/heartbeat)

# Task worker executable
add_subdirectory(src/task)
