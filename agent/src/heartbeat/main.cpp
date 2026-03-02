// Heartbeat worker main entry point
#include <agent/heartbeat_worker.hpp>
#include <cstdlib>
#include <iostream>

#ifdef _WIN32
#include <windows.h>
#endif

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

    agent::HeartbeatWorker worker;
    worker.Run();

    return 0;
}
