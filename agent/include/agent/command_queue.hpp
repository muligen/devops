#ifndef AGENT_COMMAND_QUEUE_HPP
#define AGENT_COMMAND_QUEUE_HPP

#include <agent/command.hpp>
#include <mutex>
#include <optional>
#include <queue>

namespace agent {

// Thread-safe command queue
class CommandQueue {
public:
    explicit CommandQueue(size_t max_size = 100);

    // Add command to queue
    // Returns false if queue is full
    bool Push(const Command& command);

    // Get next command from queue
    std::optional<Command> Pop();

    // Check if queue is empty
    bool Empty() const;

    // Check if queue is full
    bool Full() const;

    // Get current queue size
    size_t Size() const;

    // Get max queue size
    size_t MaxSize() const { return max_size_; }

private:
    mutable std::mutex mutex_;
    std::queue<Command> queue_;
    size_t max_size_;
};

}  // namespace agent

#endif  // AGENT_COMMAND_QUEUE_HPP
