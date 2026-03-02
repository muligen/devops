// Command queue implementation
#include <agent/command_queue.hpp>

namespace agent {

CommandQueue::CommandQueue(size_t max_size)
    : max_size_(max_size) {
}

bool CommandQueue::Push(const Command& command) {
    std::lock_guard<std::mutex> lock(mutex_);

    if (queue_.size() >= max_size_) {
        return false;
    }

    queue_.push(command);
    return true;
}

std::optional<Command> CommandQueue::Pop() {
    std::lock_guard<std::mutex> lock(mutex_);

    if (queue_.empty()) {
        return std::nullopt;
    }

    Command command = std::move(queue_.front());
    queue_.pop();
    return command;
}

bool CommandQueue::Empty() const {
    std::lock_guard<std::mutex> lock(mutex_);
    return queue_.empty();
}

bool CommandQueue::Full() const {
    std::lock_guard<std::mutex> lock(mutex_);
    return queue_.size() >= max_size_;
}

size_t CommandQueue::Size() const {
    std::lock_guard<std::mutex> lock(mutex_);
    return queue_.size();
}

}  // namespace agent
