// Agent Windows service implementation
#include <agent/config.hpp>
#include <agent/logger.hpp>
#include <agent/process_manager.hpp>
#include <agent/websocket_client.hpp>
#include <agent/common/types.hpp>
#include <agent/heartbeat_worker.hpp>
#include <nlohmann/json.hpp>

#ifdef _WIN32
#include <windows.h>
#include <winsvc.h>
#include <pdh.h>
#include <pdhmsg.h>
#include <psapi.h>
#pragma comment(lib, "pdh.lib")
#endif

#include <atomic>
#include <memory>
#include <string>
#include <thread>
#include <chrono>
#include <sstream>
#include <iomanip>

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
    std::unique_ptr<WebSocketClient> ws_client;
    std::atomic<bool> ws_connected{false};
};

std::unique_ptr<ServiceContext> g_context;

// Helper functions for system metrics
#ifdef _WIN32
double GetCpuUsage() {
    static PDH_HQUERY cpuQuery = nullptr;
    static PDH_HCOUNTER cpuCounter = nullptr;

    if (cpuQuery == nullptr) {
        PdhOpenQuery(nullptr, 0, &cpuQuery);
        PdhAddEnglishCounterA(cpuQuery, "\\Processor(_Total)\\% Processor Time", 0, &cpuCounter);
        PdhCollectQueryData(cpuQuery);
        return 0.0;
    }

    PdhCollectQueryData(cpuQuery);
    PDH_FMT_COUNTERVALUE counterVal;
    PdhGetFormattedCounterValue(cpuCounter, PDH_FMT_DOUBLE, nullptr, &counterVal);
    return counterVal.doubleValue;
}

uint64_t GetTotalMemory() {
    MEMORYSTATUSEX memInfo;
    memInfo.dwLength = sizeof(MEMORYSTATUSEX);
    GlobalMemoryStatusEx(&memInfo);
    return memInfo.ullTotalPhys;
}

uint64_t GetUsedMemory() {
    MEMORYSTATUSEX memInfo;
    memInfo.dwLength = sizeof(MEMORYSTATUSEX);
    GlobalMemoryStatusEx(&memInfo);
    return memInfo.ullTotalPhys - memInfo.ullAvailPhys;
}

double GetMemoryPercent() {
    MEMORYSTATUSEX memInfo;
    memInfo.dwLength = sizeof(MEMORYSTATUSEX);
    GlobalMemoryStatusEx(&memInfo);
    return memInfo.dwMemoryLoad;
}

uint64_t GetTotalDisk() {
    ULARGE_INTEGER freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes;
    GetDiskFreeSpaceExA("C:\\", &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes);
    return totalNumberOfBytes.QuadPart;
}

uint64_t GetUsedDisk() {
    ULARGE_INTEGER freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes;
    GetDiskFreeSpaceExA("C:\\", &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes);
    return totalNumberOfBytes.QuadPart - totalNumberOfFreeBytes.QuadPart;
}

double GetDiskPercent() {
    ULARGE_INTEGER freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes;
    GetDiskFreeSpaceExA("C:\\", &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes);
    if (totalNumberOfBytes.QuadPart == 0) return 0.0;
    return 100.0 * (totalNumberOfBytes.QuadPart - totalNumberOfFreeBytes.QuadPart) / totalNumberOfBytes.QuadPart;
}

uint64_t GetSystemUptime() {
    return GetTickCount64() / 1000;
}
#else
double GetCpuUsage() { return 0.0; }
uint64_t GetTotalMemory() { return 0; }
uint64_t GetUsedMemory() { return 0; }
double GetMemoryPercent() { return 0.0; }
uint64_t GetTotalDisk() { return 0; }
uint64_t GetUsedDisk() { return 0; }
double GetDiskPercent() { return 0.0; }
uint64_t GetSystemUptime() { return 0; }
#endif

// Send heartbeat message
void SendHeartbeat() {
    if (!g_context || !g_context->ws_client || !g_context->ws_connected) {
        return;
    }

    nlohmann::json msg;
    msg["type"] = "heartbeat";
    msg["data"] = {
        {"timestamp", std::chrono::duration_cast<std::chrono::seconds>(
            std::chrono::system_clock::now().time_since_epoch()).count()}
    };

    g_context->ws_client->Send(msg.dump());
}

// Send metrics message
void SendMetrics() {
    if (!g_context || !g_context->ws_client || !g_context->ws_connected) {
        return;
    }

    nlohmann::json msg;
    msg["type"] = "metrics";
    msg["data"] = {
        {"cpu_usage", GetCpuUsage()},
        {"memory", {
            {"total", GetTotalMemory()},
            {"used", GetUsedMemory()},
            {"percent", GetMemoryPercent()}
        }},
        {"disk", {
            {"total", GetTotalDisk()},
            {"used", GetUsedDisk()},
            {"percent", GetDiskPercent()}
        }},
        {"uptime", GetSystemUptime()}
    };

    g_context->ws_client->Send(msg.dump());
    LOG_DEBUG("Sent metrics: cpu={:.1f}%, mem={:.1f}%, disk={:.1f}%",
              GetCpuUsage(), GetMemoryPercent(), GetDiskPercent());
}

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

    // Initialize WebSocket client
    WebSocketConfig ws_config;
    ws_config.server_url = config.agent.server_url;
    ws_config.agent_id = config.agent.id;
    ws_config.token = config.agent.token;
    ws_config.connect_timeout = config.connection.retry_interval;
    ws_config.reconnect_base_delay = config.connection.retry_interval;
    ws_config.reconnect_max_delay = config.connection.max_retry_interval;
    ws_config.ping_interval = config.connection.ping_interval;
    ws_config.pong_timeout = config.connection.pong_timeout;

    g_context->ws_client = WebSocketClient::Create(ws_config);

    // Set up WebSocket callbacks
    g_context->ws_client->SetConnectionCallback(
        [](ConnectionState state) {
            switch (state) {
                case ConnectionState::kConnected:
                    LOG_INFO("WebSocket connected");
                    break;
                case ConnectionState::kAuthenticated:
                    LOG_INFO("WebSocket authenticated");
                    g_context->ws_connected = true;
                    break;
                case ConnectionState::kDisconnected:
                    LOG_WARN("WebSocket disconnected");
                    g_context->ws_connected = false;
                    break;
                case ConnectionState::kConnecting:
                    LOG_INFO("WebSocket connecting...");
                    break;
                case ConnectionState::kAuthenticating:
                    LOG_INFO("WebSocket authenticating...");
                    break;
            }
        });

    g_context->ws_client->SetErrorCallback(
        [](const std::string& error) {
            LOG_ERROR("WebSocket error: {}", error);
        });

    g_context->ws_client->SetMessageCallback(
        [](const std::string& message) {
            LOG_DEBUG("WebSocket message received: {}", message.substr(0, 100));
            // TODO: Handle incoming messages (commands from server)
        });

    g_context->ws_client->SetAuthCallback(
        [](bool success, const std::string& session_id) {
            if (success) {
                LOG_INFO("Authentication successful, session: {}", session_id);
            } else {
                LOG_ERROR("Authentication failed");
            }
        });

    // Connect to server
    LOG_INFO("Connecting to server: {}", config.agent.server_url);
    g_context->ws_client->Connect(config.agent.server_url);

    // Start workers
    g_context->process_manager->StartAll();

    g_running = true;
    ServiceReportStatus(SERVICE_RUNNING, NO_ERROR, 0);

    LOG_INFO("Agent started successfully");

    // Timing for heartbeat and metrics
    auto last_heartbeat = std::chrono::steady_clock::now();
    auto last_metrics = std::chrono::steady_clock::now();
    const auto heartbeat_interval = std::chrono::seconds(10);  // Send heartbeat every 10s
    const auto metrics_interval = std::chrono::seconds(30);    // Send metrics every 30s

    // Main service loop
    while (g_running) {
        auto now = std::chrono::steady_clock::now();

        // Send heartbeat periodically
        if (now - last_heartbeat >= heartbeat_interval) {
            SendHeartbeat();
            last_heartbeat = now;
        }

        // Send metrics periodically
        if (now - last_metrics >= metrics_interval) {
            SendMetrics();
            last_metrics = now;
        }

        // Monitor worker processes
        g_context->process_manager->Monitor();
        std::this_thread::sleep_for(std::chrono::seconds(1));
    }

    // Shutdown
    LOG_INFO("Agent shutting down...");
    if (g_context->ws_client) {
        g_context->ws_client->Disconnect();
    }
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
