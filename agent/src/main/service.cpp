// Agent Windows service implementation
#include <agent/config.hpp>
#include <agent/logger.hpp>
#include <agent/process_manager.hpp>
#include <agent/common/types.hpp>

#ifdef _WIN32
#include <windows.h>
#include <winsvc.h>
#endif

#include <atomic>
#include <memory>
#include <string>

namespace agent {

#ifdef _WIN32

// Service name and display name
constexpr const char* kServiceName = "AgentTeams";
constexpr const char* kServiceDisplayName = "AgentTeams Agent";

// Global service status
SERVICE_STATUS g_service_status = {};
SERVICE_STATUS_HANDLE g_status_handle = nullptr;
std::atomic<bool> g_running{false};

// Forward declarations
void WINAPI ServiceMain(DWORD argc, LPSTR* argv);
void WINAPI ServiceCtrlHandler(DWORD ctrl);
void ServiceReportStatus(DWORD current_state, DWORD win32_exit_code, DWORD wait_hint);

// Service context
struct ServiceContext {
    Config config;
    std::unique_ptr<ProcessManager> process_manager;
};

std::unique_ptr<ServiceContext> g_context;

void ServiceRun(const Config& config) {
    LOG_INFO("Agent starting...");

    // Initialize process manager
    ProcessManagerConfig pm_config;
    pm_config.heartbeat_worker_path = "worker_heartbeat.exe";
    pm_config.task_worker_path = "worker_task.exe";
    pm_config.restart_delay = std::chrono::seconds(5);
    pm_config.max_restart_attempts = 3;

    g_context = std::make_unique<ServiceContext>();
    g_context->config = config;
    g_context->process_manager = std::make_unique<ProcessManager>(pm_config);

    // Set up state callback
    g_context->process_manager->SetStateCallback(
        [](const std::string& name, bool running) {
            if (!running) {
                LOG_WARN("Worker {} stopped", name);
            }
        });

    // Start workers
    g_context->process_manager->StartAll();

    g_running = true;
    ServiceReportStatus(SERVICE_RUNNING, NO_ERROR, 0);

    LOG_INFO("Agent started successfully");

    // Main service loop
    while (g_running) {
        // Monitor worker processes
        g_context->process_manager->Monitor();
        std::this_thread::sleep_for(std::chrono::seconds(5));
    }

    // Shutdown
    LOG_INFO("Agent shutting down...");
    g_context->process_manager->StopAll();
    ServiceReportStatus(SERVICE_STOPPED, NO_ERROR, 0);
}

void ServiceInstall() {
    SC_HANDLE sc_manager = OpenSCManagerA(nullptr, nullptr, SC_MANAGER_ALL_ACCESS);
    if (!sc_manager) {
        LOG_ERROR("Failed to open service manager: {}", GetLastError());
        return;
    }

    char path[MAX_PATH];
    if (!GetModuleFileNameA(nullptr, path, MAX_PATH)) {
        LOG_ERROR("Failed to get module path: {}", GetLastError());
        CloseServiceHandle(sc_manager);
        return;
    }

    SC_HANDLE service = CreateServiceA(
        sc_manager,
        kServiceName,
        kServiceDisplayName,
        SERVICE_ALL_ACCESS,
        SERVICE_WIN32_OWN_PROCESS,
        SERVICE_AUTO_START,
        SERVICE_ERROR_NORMAL,
        path,
        nullptr, nullptr, nullptr, nullptr, nullptr);

    if (service) {
        LOG_INFO("Service installed successfully");
        CloseServiceHandle(service);
    } else {
        LOG_ERROR("Failed to install service: {}", GetLastError());
    }

    CloseServiceHandle(sc_manager);
}

void ServiceUninstall() {
    SC_HANDLE sc_manager = OpenSCManagerA(nullptr, nullptr, SC_MANAGER_ALL_ACCESS);
    if (!sc_manager) {
        LOG_ERROR("Failed to open service manager: {}", GetLastError());
        return;
    }

    SC_HANDLE service = OpenServiceA(sc_manager, kServiceName, DELETE);
    if (service) {
        if (DeleteService(service)) {
            LOG_INFO("Service uninstalled successfully");
        } else {
            LOG_ERROR("Failed to uninstall service: {}", GetLastError());
        }
        CloseServiceHandle(service);
    } else {
        LOG_ERROR("Failed to open service: {}", GetLastError());
    }

    CloseServiceHandle(sc_manager);
}

void WINAPI ServiceMain(DWORD argc, LPSTR* argv) {
    g_status_handle = RegisterServiceCtrlHandlerA(kServiceName, ServiceCtrlHandler);
    if (!g_status_handle) {
        LOG_ERROR("Failed to register service handler: {}", GetLastError());
        return;
    }

    g_service_status.dwServiceType = SERVICE_WIN32_OWN_PROCESS;
    g_service_status.dwControlsAccepted = 0;
    g_service_status.dwCurrentState = SERVICE_START_PENDING;
    g_service_status.dwWin32ExitCode = 0;
    g_service_status.dwServiceSpecificExitCode = 0;
    g_service_status.dwCheckPoint = 0;
    g_service_status.dwWaitHint = 0;

    ServiceReportStatus(SERVICE_START_PENDING, NO_ERROR, 3000);

    // Load configuration
    auto config = Config::LoadFromFile("agent.yaml");
    if (!config) {
        LOG_ERROR("Failed to load configuration");
        ServiceReportStatus(SERVICE_STOPPED, ERROR_INVALID_DATA, 0);
        return;
    }

    // Initialize logging
    LoggerConfig log_config;
    log_config.level = config->logging.level;
    log_config.file = config->logging.file;
    log_config.max_size = config->logging.max_size;
    log_config.max_files = config->logging.max_files;
    Logger::Init(log_config);

    ServiceRun(*config);
}

void WINAPI ServiceCtrlHandler(DWORD ctrl) {
    switch (ctrl) {
        case SERVICE_CONTROL_STOP:
            ServiceReportStatus(SERVICE_STOP_PENDING, NO_ERROR, 0);
            g_running = false;
            break;

        case SERVICE_CONTROL_PAUSE:
            ServiceReportStatus(SERVICE_PAUSED, NO_ERROR, 0);
            break;

        case SERVICE_CONTROL_CONTINUE:
            ServiceReportStatus(SERVICE_RUNNING, NO_ERROR, 0);
            break;

        case SERVICE_CONTROL_SHUTDOWN:
            g_running = false;
            break;

        default:
            break;
    }
}

void ServiceReportStatus(DWORD current_state, DWORD win32_exit_code, DWORD wait_hint) {
    g_service_status.dwCurrentState = current_state;
    g_service_status.dwWin32ExitCode = win32_exit_code;
    g_service_status.dwWaitHint = wait_hint;

    if (current_state == SERVICE_START_PENDING) {
        g_service_status.dwControlsAccepted = 0;
    } else {
        g_service_status.dwControlsAccepted = SERVICE_ACCEPT_STOP |
                                               SERVICE_ACCEPT_PAUSE_CONTINUE |
                                               SERVICE_ACCEPT_SHUTDOWN;
    }

    if (current_state == SERVICE_RUNNING || current_state == SERVICE_STOPPED) {
        g_service_status.dwCheckPoint = 0;
    } else {
        g_service_status.dwCheckPoint++;
    }

    SetServiceStatus(g_status_handle, &g_service_status);
}

#endif  // _WIN32

// Public API

bool ServiceInstallPlatform() {
#ifdef _WIN32
    ServiceInstall();
    return true;
#else
    return false;
#endif
}

bool ServiceUninstallPlatform() {
#ifdef _WIN32
    ServiceUninstall();
    return true;
#else
    return false;
#endif
}

int ServiceMainPlatform(int argc, char* argv[]) {
#ifdef _WIN32
    SERVICE_TABLE_ENTRYA service_table[] = {
        {const_cast<LPSTR>(kServiceName), (LPSERVICE_MAIN_FUNCTIONA)ServiceMain},
        {nullptr, nullptr}
    };

    if (!StartServiceCtrlDispatcherA(service_table)) {
        // Not running as service, run in console mode
        if (GetLastError() == ERROR_FAILED_SERVICE_CONTROLLER_CONNECT) {
            // Load configuration
            auto config = Config::LoadFromFile("agent.yaml");
            if (!config) {
                fprintf(stderr, "Failed to load configuration\n");
                return 1;
            }

            // Initialize logging
            LoggerConfig log_config;
            log_config.level = config->logging.level;
            log_config.file = config->logging.file;
            log_config.console_output = true;
            Logger::Init(log_config);

            LOG_INFO("Running in console mode");

            ServiceRun(*config);
            return 0;
        }

        LOG_ERROR("Failed to start service dispatcher: {}", GetLastError());
        return 1;
    }

    return 0;
#else
    // TODO: Implement for other platforms
    return 1;
#endif
}

}  // namespace agent
