// Agent main entry point
#include <agent/config.hpp>
#include <agent/logger.hpp>
#include <agent/common/types.hpp>

#include <cstdlib>
#include <iostream>
#include <string>

#ifdef _WIN32
#include <windows.h>
#endif

// Forward declarations
namespace agent {
bool ServiceInstallPlatform();
bool ServiceUninstallPlatform();
int ServiceMainPlatform(int argc, char* argv[]);
}

void PrintUsage(const char* program) {
    std::cout << "Usage: " << program << " [options]\n\n"
              << "Options:\n"
              << "  --install       Install as Windows service\n"
              << "  --uninstall     Uninstall Windows service\n"
              << "  --config FILE   Specify configuration file (default: agent.yaml)\n"
              << "  --version       Show version information\n"
              << "  --help          Show this help message\n";
}

void PrintVersion() {
    std::cout << "AgentTeams Agent v" << agent::kAgentVersion << "\n";
}

int main(int argc, char* argv[]) {
#ifdef _WIN32
    // Set console code page to UTF-8
    SetConsoleOutputCP(CP_UTF8);
    SetConsoleCP(CP_UTF8);
#endif

    // Parse command line arguments
    std::string config_file = "agent.yaml";
    bool install_service = false;
    bool uninstall_service = false;

    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];

        if (arg == "--install") {
            install_service = true;
        } else if (arg == "--uninstall") {
            uninstall_service = true;
        } else if (arg == "--config" || arg == "-c") {
            if (i + 1 < argc) {
                config_file = argv[++i];
            }
        } else if (arg == "--version" || arg == "-v") {
            PrintVersion();
            return 0;
        } else if (arg == "--help" || arg == "-h") {
            PrintUsage(argv[0]);
            return 0;
        }
    }

    // Handle service install/uninstall
    if (install_service) {
        agent::ServiceInstallPlatform();
        return 0;
    }

    if (uninstall_service) {
        agent::ServiceUninstallPlatform();
        return 0;
    }

    // Run as service or console application
    return agent::ServiceMainPlatform(argc, argv);
}
