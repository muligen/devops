// Agent auto-updater implementation
#include <agent/config.hpp>
#include <agent/logger.hpp>
#include <agent/common/types.hpp>
#include <agent/process_manager.hpp>

#ifdef _WIN32
#include <windows.h>
#include <winhttp.h>
#include <shlobj.h>
#include <softpub.h>
#include <wintrust.h>
#include <bcrypt.h>
#endif

#include <chrono>
#include <filesystem>
#include <fstream>
#include <functional>
#include <memory>
#include <sstream>

namespace agent {

namespace fs = std::filesystem;

// Updater implementation
class Updater {
public:
    using ProgressCallback = std::function<void(int percent)>;
    using StatusCallback = std::function<void(const std::string& status)>;

    Updater(const Config& config, ProcessManager& process_manager)
        : config_(config), process_manager_(process_manager) {}

    // Check for updates
    struct UpdateInfo {
        bool available{false};
        std::string version;
        std::string download_url;
        std::string file_hash;
        std::string signature;
        std::string release_notes;
    };

    UpdateInfo CheckForUpdate() {
        UpdateInfo info;
        // Implementation would query the server's version endpoint
        // For now, return empty info
        LOG_DEBUG("Checking for updates...");
        return info;
    }

    // Download update package
    bool DownloadUpdate(const std::string& url, const std::string& dest_path,
                       ProgressCallback progress = nullptr) {
#ifdef _WIN32
        // WinHTTP implementation for downloading
        LOG_INFO("Downloading update from: {}", url);

        // Parse URL
        std::string host, path;
        if (!ParseUrl(url, host, path)) {
            LOG_ERROR("Failed to parse URL: {}", url);
            return false;
        }

        // Initialize WinHTTP
        HINTERNET session = WinHttpOpen(
            L"AgentTeams/1.0",
            WINHTTP_ACCESS_TYPE_DEFAULT_PROXY,
            WINHTTP_NO_PROXY_NAME,
            WINHTTP_NO_PROXY_BYPASS, 0);
        if (!session) {
            LOG_ERROR("Failed to initialize WinHTTP: {}", GetLastError());
            return false;
        }

        // Connect to server
        std::wstring whost(host.begin(), host.end());
        HINTERNET connect = WinHttpConnect(session, whost.c_str(),
            INTERNET_DEFAULT_HTTPS_PORT, 0);
        if (!connect) {
            LOG_ERROR("Failed to connect to server: {}", GetLastError());
            WinHttpCloseHandle(session);
            return false;
        }

        // Open request
        std::wstring wpath(path.begin(), path.end());
        HINTERNET request = WinHttpOpenRequest(connect, L"GET", wpath.c_str(),
            nullptr, WINHTTP_NO_REFERER,
            WINHTTP_DEFAULT_ACCEPT_TYPES,
            WINHTTP_FLAG_SECURE);
        if (!request) {
            LOG_ERROR("Failed to create request: {}", GetLastError());
            WinHttpCloseHandle(connect);
            WinHttpCloseHandle(session);
            return false;
        }

        // Send request
        BOOL result = WinHttpSendRequest(request,
            WINHTTP_NO_ADDITIONAL_HEADERS, 0,
            WINHTTP_NO_REQUEST_DATA, 0,
            0, 0);
        if (!result) {
            LOG_ERROR("Failed to send request: {}", GetLastError());
            WinHttpCloseHandle(request);
            WinHttpCloseHandle(connect);
            WinHttpCloseHandle(session);
            return false;
        }

        // Receive response
        result = WinHttpReceiveResponse(request, nullptr);
        if (!result) {
            LOG_ERROR("Failed to receive response: {}", GetLastError());
            WinHttpCloseHandle(request);
            WinHttpCloseHandle(connect);
            WinHttpCloseHandle(session);
            return false;
        }

        // Get content length
        DWORD content_length = 0;
        DWORD buffer_size = sizeof(content_length);
        WinHttpQueryHeaders(request,
            WINHTTP_QUERY_CONTENT_LENGTH | WINHTTP_QUERY_FLAG_NUMBER,
            WINHTTP_HEADER_NAME_BY_INDEX,
            &content_length, &buffer_size, WINHTTP_NO_HEADER_INDEX);

        // Download to file
        std::ofstream file(dest_path, std::ios::binary);
        if (!file) {
            LOG_ERROR("Failed to create file: {}", dest_path);
            WinHttpCloseHandle(request);
            WinHttpCloseHandle(connect);
            WinHttpCloseHandle(session);
            return false;
        }

        DWORD total_read = 0;
        DWORD bytes_read = 0;
        char buffer[8192];

        while (WinHttpReadData(request, buffer, sizeof(buffer), &bytes_read) && bytes_read > 0) {
            file.write(buffer, bytes_read);
            total_read += bytes_read;

            if (progress && content_length > 0) {
                int percent = static_cast<int>((total_read * 100) / content_length);
                progress(percent);
            }
        }

        file.close();
        WinHttpCloseHandle(request);
        WinHttpCloseHandle(connect);
        WinHttpCloseHandle(session);

        LOG_INFO("Download complete: {} bytes", total_read);
        return true;
#else
        LOG_ERROR("Download not implemented for this platform");
        return false;
#endif
    }

    // Verify file hash
    bool VerifyHash(const std::string& file_path, const std::string& expected_hash) {
#ifdef _WIN32
        // Read file
        std::ifstream file(file_path, std::ios::binary);
        if (!file) {
            LOG_ERROR("Failed to open file for hashing: {}", file_path);
            return false;
        }

        // Calculate SHA256
        BCRYPT_ALG_HANDLE alg_handle = nullptr;
        NTSTATUS status = BCryptOpenAlgorithmProvider(&alg_handle,
            BCRYPT_SHA256_ALGORITHM, nullptr, 0);
        if (!BCRYPT_SUCCESS(status)) {
            LOG_ERROR("Failed to open SHA256 algorithm provider");
            return false;
        }

        // Hash the file
        BCRYPT_HASH_HANDLE hash_handle = nullptr;
        status = BCryptCreateHash(alg_handle, &hash_handle, nullptr, 0, nullptr, 0, 0);
        if (!BCRYPT_SUCCESS(status)) {
            BCryptCloseAlgorithmProvider(alg_handle, 0);
            return false;
        }

        char buffer[8192];
        while (file.read(buffer, sizeof(buffer))) {
            status = BCryptHashData(hash_handle,
                reinterpret_cast<PUCHAR>(buffer),
                static_cast<ULONG>(file.gcount()), 0);
            if (!BCRYPT_SUCCESS(status)) {
                BCryptDestroyHash(hash_handle);
                BCryptCloseAlgorithmProvider(alg_handle, 0);
                return false;
            }
        }

        // Finalize hash
        DWORD hash_length = 0;
        DWORD result_length = 0;
        BCryptGetProperty(hash_handle, BCRYPT_HASH_LENGTH,
            reinterpret_cast<PUCHAR>(&hash_length), sizeof(hash_length),
            &result_length, 0);

        std::vector<BYTE> hash_value(hash_length);
        status = BCryptFinishHash(hash_handle, hash_value.data(), hash_length, 0);

        BCryptDestroyHash(hash_handle);
        BCryptCloseAlgorithmProvider(alg_handle, 0);

        if (!BCRYPT_SUCCESS(status)) {
            return false;
        }

        // Convert to hex string
        std::stringstream ss;
        for (BYTE b : hash_value) {
            ss << std::hex << std::setw(2) << std::setfill('0') << static_cast<int>(b);
        }
        std::string actual_hash = ss.str();

        // Compare
        bool match = (actual_hash == expected_hash);
        if (!match) {
            LOG_ERROR("Hash mismatch: expected {}, got {}", expected_hash, actual_hash);
        }
        return match;
#else
        return false;
#endif
    }

    // Verify file signature (Windows)
    bool VerifySignature(const std::string& file_path) {
#ifdef _WIN32
        WINTRUST_FILE_INFO file_info;
        memset(&file_info, 0, sizeof(file_info));
        file_info.cbStruct = sizeof(WINTRUST_FILE_INFO);
        std::wstring wpath(file_path.begin(), file_path.end());
        file_info.pcwszFilePath = wpath.c_str();

        WINTRUST_DATA trust_data;
        memset(&trust_data, 0, sizeof(trust_data));
        trust_data.cbStruct = sizeof(WINTRUST_DATA);
        trust_data.dwUIChoice = WTD_UI_NONE;
        trust_data.fdwRevocationChecks = WTD_REVOKE_NONE;
        trust_data.dwUnionChoice = WTD_CHOICE_FILE;
        trust_data.pFile = &file_info;
        trust_data.dwStateAction = WTD_STATEACTION_VERIFY;

        GUID policy_guid = WINTRUST_ACTION_GENERIC_VERIFY_V2;
        LONG result = WinVerifyTrust(nullptr, &policy_guid, &trust_data);

        // Clean up
        trust_data.dwStateAction = WTD_STATEACTION_CLOSE;
        WinVerifyTrust(nullptr, &policy_guid, &trust_data);

        if (result == ERROR_SUCCESS) {
            LOG_INFO("Signature verification passed");
            return true;
        } else {
            LOG_ERROR("Signature verification failed: {}", result);
            return false;
        }
#else
        return false;
#endif
    }

    // Perform update
    bool PerformUpdate(const UpdateInfo& info, StatusCallback status = nullptr) {
        if (status) status("checking_idle");

        // Check if idle
        if (config_.update.idle_required && !process_manager_.IsRunning("task")) {
            LOG_INFO("Agent not idle, deferring update");
            return false;
        }

        if (status) status("downloading");

        // Download update
        std::string temp_path = std::filesystem::temp_directory_path().string();
        std::string download_path = temp_path + "\\agent_update.zip";

        if (!DownloadUpdate(info.download_url, download_path, [&](int percent) {
            LOG_DEBUG("Download progress: {}%", percent);
        })) {
            if (status) status("download_failed");
            return false;
        }

        if (status) status("verifying");

        // Verify hash
        if (!info.file_hash.empty() && !VerifyHash(download_path, info.file_hash)) {
            if (status) status("hash_failed");
            return false;
        }

        // Verify signature
        if (!VerifySignature(download_path)) {
            if (status) status("signature_failed");
            return false;
        }

        if (status) status("installing");

        // Stop workers
        process_manager_.StopAll();

        // Backup and replace files
        // TODO: Implement actual file replacement with rollback

        if (status) status("completed");

        // Restart workers
        process_manager_.StartAll();

        LOG_INFO("Update completed successfully to version {}", info.version);
        return true;
    }

private:
    const Config& config_;
    ProcessManager& process_manager_;

    bool ParseUrl(const std::string& url, std::string& host, std::string& path) {
        // Simple URL parsing
        size_t pos = url.find("://");
        if (pos == std::string::npos) {
            return false;
        }

        std::string after_scheme = url.substr(pos + 3);
        size_t slash_pos = after_scheme.find('/');
        if (slash_pos == std::string::npos) {
            host = after_scheme;
            path = "/";
        } else {
            host = after_scheme.substr(0, slash_pos);
            path = after_scheme.substr(slash_pos);
        }

        return true;
    }
};

}  // namespace agent
