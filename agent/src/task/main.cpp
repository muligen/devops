// Task worker main entry point
#include <agent/command.hpp>
#include <agent/command_queue.hpp>
#include <agent/thread_pool.hpp>
#include <agent/ipc.hpp>

#include <cstdlib>
#include <iostream>
#include <chrono>
#include <thread>

#ifdef _WIN32
#include <windows.h>
#endif

namespace {

constexpr size_t kDefaultMaxConcurrent = 4;
constexpr size_t kDefaultQueueSize = 100;

}  // namespace

void PrintUsage(const char* program) {
    std::cout << "Usage: " << program << " [options]\n\n"
              << "This is a worker process managed by the main agent process.\n"
              << "It should not be run directly.\n";
}

int main(int argc, char* argv[]) {
#ifdef _WIN32
    SetConsoleOutputCP(CP_UTF8);
    SetConsoleCP(CP_UTF8);
#endif

    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        if (arg == "--help" || arg == "-h") {
            PrintUsage(argv[0]);
            return 0;
        }
    }

    // Create command queue
    agent::CommandQueue queue(kDefaultQueueSize);

    // Create thread pool for concurrent execution
    agent::ThreadPool pool(kDefaultMaxConcurrent);

    // TODO: Connect to main process via IPC
    // TODO: Receive commands from main process
    // TODO: Execute commands and send results

    // Keep running
    while (true) {
        std::this_thread::sleep_for(std::chrono::seconds(1));
    }

    return 0;
}
