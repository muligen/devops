// Inter-process communication implementation (shared)
#include <agent/ipc.hpp>

#include <nlohmann/json.hpp>

#ifdef _WIN32
#include <windows.h>
// Undef Windows macros that conflict with our methods
#ifdef SendMessage
#undef SendMessage
#endif
#endif

namespace agent {

class IpcClient::IpcClientImpl {
public:
    IpcClientImpl() = default;
    ~IpcClientImpl() { Disconnect(); }

    bool Connect(const std::string& pipe_name) {
#ifdef _WIN32
        std::wstring wide_name(pipe_name.begin(), pipe_name.end());

        pipe_ = CreateFileW(
            wide_name.c_str(),
            GENERIC_READ | GENERIC_WRITE,
            0,
            nullptr,
            OPEN_EXISTING,
            0,
            nullptr
        );

        if (pipe_ == INVALID_HANDLE_VALUE) {
            return false;
        }

        connected_ = true;
        return true;
#else
        return false;
#endif
    }

    void Disconnect() {
#ifdef _WIN32
        if (pipe_ != INVALID_HANDLE_VALUE && pipe_ != nullptr) {
            CloseHandle(pipe_);
            pipe_ = INVALID_HANDLE_VALUE;
        }
#endif
        connected_ = false;
    }

    bool SendMessage(const std::string& message) {
#ifdef _WIN32
        if (!connected_ || pipe_ == INVALID_HANDLE_VALUE) {
            return false;
        }

        DWORD bytes_written = 0;
        DWORD message_size = static_cast<DWORD>(message.size());

        // Write message length first
        if (!WriteFile(pipe_, &message_size, sizeof(message_size), &bytes_written, nullptr)) {
            return false;
        }

        // Write message content
        if (!WriteFile(pipe_, message.data(), message_size, &bytes_written, nullptr)) {
            return false;
        }

        FlushFileBuffers(pipe_);
        return true;
#else
        return false;
#endif
    }

    std::string ReceiveMessage() {
#ifdef _WIN32
        if (!connected_ || pipe_ == INVALID_HANDLE_VALUE) {
            return "";
        }

        DWORD message_size = 0;
        DWORD bytes_read = 0;

        // Read message length
        if (!ReadFile(pipe_, &message_size, sizeof(message_size), &bytes_read, nullptr)) {
            return "";
        }

        if (message_size == 0 || message_size > 1024 * 1024) {
            return "";
        }

        // Read message content
        std::string buffer(message_size, '\0');
        if (!ReadFile(pipe_, buffer.data(), message_size, &bytes_read, nullptr)) {
            return "";
        }

        return buffer;
#else
        return "";
#endif
    }

    bool IsConnected() const { return connected_; }

private:
#ifdef _WIN32
    HANDLE pipe_{INVALID_HANDLE_VALUE};
#endif
    bool connected_{false};
};

// IpcClient implementation
IpcClient::IpcClient() : impl_(std::make_unique<IpcClientImpl>()) {
}

IpcClient::~IpcClient() = default;

bool IpcClient::Connect(const std::string& pipe_name) {
    return impl_->Connect(pipe_name);
}

void IpcClient::Disconnect() {
    impl_->Disconnect();
}

bool IpcClient::SendMessage(const std::string& message) {
    return impl_->SendMessage(message);
}

bool IpcClient::IsConnected() const {
    return impl_->IsConnected();
}

}  // namespace agent
