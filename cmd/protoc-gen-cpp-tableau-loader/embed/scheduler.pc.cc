#include "scheduler.pc.h"

namespace tableau {
namespace internal {
// Thread-local storage (TLS)
thread_local Scheduler* tls_sched = nullptr;

Scheduler& Scheduler::Current() {
  if (tls_sched == nullptr) {
    tls_sched = new Scheduler;
  }
  return *tls_sched;
}

int Scheduler::LoopOnce() {
  AssertInLoopThread();

  int count = 0;
  std::vector<Job> jobs;
  {
    // scoped for auto-release lock.
    // wake up immediately when there are pending tasks.
    std::unique_lock<std::mutex> lock(mutex_);
    jobs.swap(jobs_);
  }
  for (auto&& job : jobs) {
    job();
  }
  count += jobs.size();
  return count;
}

void Scheduler::Post(const Job& job) {
  std::unique_lock<std::mutex> lock(mutex_);
  jobs_.push_back(job);
}

void Scheduler::Dispatch(const Job& job) {
  if (IsLoopThread()) {
    job();  // run it immediately
  } else {
    Post(job);  // post and run it at next loop
  }
}

bool Scheduler::IsLoopThread() const { return thread_id_ == std::this_thread::get_id(); }

void Scheduler::AssertInLoopThread() const {
  if (!IsLoopThread()) {
    abort();
  }
}
}  // namespace internal
}  // namespace tableau