// clean_disk command implementation
#include <agent/command.hpp>
#include <chrono>
#include <vector>
#include <sstream>

#ifdef _WIN32
#include <windows.h>
#include <shlobj.h>
#include <filesystem>
#endif

namespace agent {

class CleanDiskCommand : public CommandExecutor {
public:
    CommandResult Execute(const Command& command) override {
        CommandResult result;
        result.id = command.id;
        result.exit_code = 0;
        result.status = TaskStatus::kSuccess;

        auto start_time = std::chrono::steady_clock::now();

        // Get categories to clean from params
        std::string categories_str = command.params.count("categories") ?
                                     command.params.at("categories") : "temp,cache,logs";

        // Parse categories
        std::vector<std::string> categories;
        std::istringstream iss(categories_str);
        std::string category;
        while (std::getline(iss, category, ',')) {
            // Trim whitespace
            size_t start = category.find_first_not_of(" \t");
            size_t end = category.find_last_not_of(" \t");
            if (start != std::string::npos && end != std::string::npos) {
                categories.push_back(category.substr(start, end - start + 1));
            }
        }

        uint64_t total_bytes_freed = 0;
        std::ostringstream output;

#ifdef _WIN32
        for (const auto& cat : categories) {
            uint64_t bytes_freed = 0;

            if (cat == "temp") {
                bytes_freed += CleanTempFolder();
                bytes_freed += CleanWindowsTemp();
            } else if (cat == "cache") {
                bytes_freed += CleanBrowserCache();
            } else if (cat == "logs") {
                bytes_freed += CleanLogFiles();
            } else if (cat == "recycle") {
                bytes_freed += EmptyRecycleBin();
            }

            total_bytes_freed += bytes_freed;
            output << "Category '" << cat << "': freed " << FormatBytes(bytes_freed) << "\n";
        }
#else
        output << "clean_disk not fully supported on this platform\n";
#endif

        output << "\nTotal freed: " << FormatBytes(total_bytes_freed);
        result.output = output.str();

        auto end_time = std::chrono::steady_clock::now();
        result.duration_seconds = std::chrono::duration<double>(end_time - start_time).count();

        return result;
    }

    bool CanHandle(CommandType type) const override {
        return type == CommandType::kCleanDisk;
    }

private:
#ifdef _WIN32
    uint64_t CleanTempFolder() {
        uint64_t bytes_freed = 0;
        wchar_t temp_path[MAX_PATH];

        if (GetTempPathW(MAX_PATH, temp_path)) {
            bytes_freed += CleanDirectory(temp_path);
        }

        return bytes_freed;
    }

    uint64_t CleanWindowsTemp() {
        uint64_t bytes_freed = 0;
        char windir[MAX_PATH];

        if (GetEnvironmentVariableA("WINDIR", windir, MAX_PATH)) {
            std::string temp_path = std::string(windir) + "\\Temp";
            bytes_freed += CleanDirectory(std::wstring(temp_path.begin(), temp_path.end()));
        }

        return bytes_freed;
    }

    uint64_t CleanBrowserCache() {
        uint64_t bytes_freed = 0;

        // Chrome cache
        char appdata[MAX_PATH];
        if (GetEnvironmentVariableA("LOCALAPPDATA", appdata, MAX_PATH)) {
            std::string chrome_cache = std::string(appdata) + "\\Google\\Chrome\\User Data\\Default\\Cache";
            bytes_freed += CleanDirectory(std::wstring(chrome_cache.begin(), chrome_cache.end()));
        }

        return bytes_freed;
    }

    uint64_t CleanLogFiles() {
        uint64_t bytes_freed = 0;

        char windir[MAX_PATH];
        if (GetEnvironmentVariableA("WINDIR", windir, MAX_PATH)) {
            std::string logs_path = std::string(windir) + "\\Logs";
            bytes_freed += CleanDirectory(std::wstring(logs_path.begin(), logs_path.end()));

            std::string temp_path = std::string(windir) + "\\Temp";
            bytes_freed += CleanFilesByExtension(std::wstring(temp_path.begin(), temp_path.end()), L".log");
        }

        return bytes_freed;
    }

    uint64_t EmptyRecycleBin() {
        uint64_t bytes_freed = 0;
        SHEmptyRecycleBinW(nullptr, nullptr, SHERB_NOCONFIRMATION | SHERB_NOPROGRESSUI | SHERB_NOSOUND);
        return bytes_freed;  // Can't easily determine bytes freed
    }

    uint64_t CleanDirectory(const std::wstring& path) {
        uint64_t bytes_freed = 0;

        try {
            std::filesystem::path dir_path(path);
            if (!std::filesystem::exists(dir_path)) {
                return 0;
            }

            for (const auto& entry : std::filesystem::recursive_directory_iterator(
                dir_path,
                std::filesystem::directory_options::skip_permission_denied
            )) {
                try {
                    if (entry.is_regular_file()) {
                        auto file_size = entry.file_size();
                        std::error_code ec;
                        if (std::filesystem::remove(entry.path(), ec) && !ec) {
                            bytes_freed += file_size;
                        }
                    }
                } catch (...) {
                    // Ignore individual file errors
                }
            }
        } catch (...) {
            // Ignore directory iteration errors
        }

        return bytes_freed;
    }

    uint64_t CleanFilesByExtension(const std::wstring& path, const std::wstring& extension) {
        uint64_t bytes_freed = 0;

        try {
            std::filesystem::path dir_path(path);
            if (!std::filesystem::exists(dir_path)) {
                return 0;
            }

            for (const auto& entry : std::filesystem::recursive_directory_iterator(
                dir_path,
                std::filesystem::directory_options::skip_permission_denied
            )) {
                try {
                    if (entry.is_regular_file() && entry.path().extension() == extension) {
                        auto file_size = entry.file_size();
                        std::error_code ec;
                        if (std::filesystem::remove(entry.path(), ec) && !ec) {
                            bytes_freed += file_size;
                        }
                    }
                } catch (...) {
                    // Ignore individual file errors
                }
            }
        } catch (...) {
            // Ignore directory iteration errors
        }

        return bytes_freed;
    }
#endif

    std::string FormatBytes(uint64_t bytes) const {
        const char* units[] = {"B", "KB", "MB", "GB", "TB"};
        int unit_index = 0;
        double size = static_cast<double>(bytes);

        while (size >= 1024.0 && unit_index < 4) {
            size /= 1024.0;
            unit_index++;
        }

        std::ostringstream oss;
        oss << std::fixed << std::setprecision(2) << size << " " << units[unit_index];
        return oss.str();
    }
};

std::unique_ptr<CommandExecutor> CreateCleanDiskExecutor() {
    return std::make_unique<CleanDiskCommand>();
}

}  // namespace agent
