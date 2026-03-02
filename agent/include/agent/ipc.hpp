#ifndef AGENT_IPC_HPP
#define AGENT_IPC_HPP

#include <memory>
#include <string>

namespace agent {

// Inter-process communication client
class IpcClient {
public:
    IpcClient();
    ~IpcClient();

    // Non-copyable
    IpcClient(const IpcClient&) = delete;
    IpcClient& operator=(const IpcClient&) = delete;

    // Connect to named pipe
    bool Connect(const std::string& pipe_name);

    // Disconnect from pipe
    void Disconnect();

    // Send message to server
    bool SendMessage(const std::string& message);

    // Check if connected
    bool IsConnected() const;

private:
    class IpcClientImpl;
    std::unique_ptr<IpcClientImpl> impl_;
};

}  // namespace agent

#endif  // AGENT_IPC_HPP
