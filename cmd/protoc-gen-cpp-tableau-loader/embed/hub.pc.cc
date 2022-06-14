#include "hub.pc.h"

#include <google/protobuf/text_format.h>

#include <fstream>
#include <sstream>

#include "logger.pc.h"
#include "registry.pc.h"

namespace tableau {
static thread_local std::string g_err_msg;
const std::string& GetErrMsg() { return g_err_msg; }
bool Message2JSON(const google::protobuf::Message& message, std::string& json) {
  google::protobuf::util::JsonPrintOptions options;
  options.add_whitespace = true;
  options.always_print_primitive_fields = true;
  options.preserve_proto_field_names = true;
  return google::protobuf::util::MessageToJsonString(message, &json, options).ok();
}

bool JSON2Message(const std::string& json, google::protobuf::Message& message) {
  if (!google::protobuf::util::JsonStringToMessage(json, &message).ok()) {
    g_err_msg = "failed to parse json file: " + GetProtoName(message) + kJSONExt;
    return false;
  }
  return true;
}

bool Text2Message(const std::string& text, google::protobuf::Message& message) {
  if (!google::protobuf::TextFormat::ParseFromString(text, &message)) {
    g_err_msg = "failed to parse text file: " + GetProtoName(message) + kTextExt;
    return false;
  }
  return true;
}
bool Wire2Message(const std::string& wire, google::protobuf::Message& message) {
  if (!message.ParseFromString(wire)) {
    g_err_msg = "failed to parse wire file: " + GetProtoName(message) + kWireExt;
    return false;
  }
  return true;
}

const std::string& GetProtoName(const google::protobuf::Message& message) {
  const auto* md = message.GetDescriptor();
  return md != nullptr ? md->name() : kEmpty;
}

bool ReadFile(const std::string& filename, std::string& content) {
  std::ifstream file(filename);
  if (!file.is_open()) {
    g_err_msg = "failed to open file: " + filename;
    return false;
  }
  std::stringstream ss;
  ss << file.rdbuf();
  content = std::move(ss.str());
  return true;
}

bool LoadMessage(const std::string& dir, google::protobuf::Message& message, Format fmt) {
  message.Clear();
  std::string basepath = dir + GetProtoName(message);
  // TODO: support 3 formats: json, text, and wire.
  std::string content;
  switch (fmt) {
    case Format::kJSON: {
      bool ok = ReadFile(basepath + kJSONExt, content);
      if (!ok) {
        return false;
      }
      return JSON2Message(content, message);
    }
    case Format::kText: {
      bool ok = ReadFile(basepath + kTextExt, content);
      if (!ok) {
        return false;
      }
      return Text2Message(content, message);
    }
    case Format::kWire: {
      bool ok = ReadFile(basepath + kWireExt, content);
      if (!ok) {
        return false;
      }
      return Wire2Message(content, message);
    }
    default: {
      g_err_msg = "unsupported format: %d" + static_cast<int>(fmt);
      return false;
    }
  }
}

bool StoreMessage(const std::string& dir, google::protobuf::Message& message, Format fmt) {
  // TODO: write protobuf message to file, support 3 formats: json, text, and wire.
  return false;
}

bool Hub::Load(const std::string& dir, Filter filter, Format fmt) {
  auto msger_container = LoadNewMessagerContainer(dir, filter, fmt);
  if (!msger_container) {
    return false;
  }
  SetMessagerContainer(msger_container);
  return true;
}

bool Hub::AsyncLoad(const std::string& dir, Filter filter, Format fmt) {
  auto msger_container = LoadNewMessagerContainer(dir, filter, fmt);
  if (!msger_container) {
    return false;
  }
  sched_->Dispatch(std::bind(&Hub::SetMessagerContainer, this, msger_container));
  return true;
}

int Hub::LoopOnce() { return sched_->LoopOnce(); }
void Hub::InitScheduler() {
  sched_ = new internal::Scheduler();
  sched_->Current();
}

MessagerContainer Hub::LoadNewMessagerContainer(const std::string& dir, Filter filter, Format fmt) {
  auto msger_container = NewMessagerContainer(filter);
  for (auto iter : *msger_container) {
    auto&& name = iter.first;
    ATOM_TRACE("loading %s", name.c_str());
    bool ok = iter.second->Load(dir, fmt);
    if (!ok) {
      ATOM_ERROR("load %s failed: %s", name.c_str(), GetErrMsg().c_str());
      return nullptr;
    }
    ATOM_TRACE("loaded %s", name.c_str());
  }
  return msger_container;
}

MessagerContainer Hub::NewMessagerContainer(Filter filter) {
  MessagerContainer msger_container = std::make_shared<MessagerMap>();
  for (auto&& it : Registry::registrar) {
    if (filter == nullptr || filter(it.first)) {
      (*msger_container)[it.first] = it.second();
    }
  }
  return msger_container;
}

void Hub::SetMessagerContainer(MessagerContainer msger_container) {
  // replace with thread-safe guarantee.
  std::unique_lock<std::mutex> lock(mutex_);
  msger_container_ = msger_container;
}

MessagerContainer Hub::GetMessagerContainerWithProvider() const {
  if (msger_container_provider_ != nullptr) {
    return msger_container_provider_();
  }
  return msger_container_;
}

const std::shared_ptr<Messager> Hub::GetMessager(const std::string& name) const {
  auto container = GetMessagerContainerWithProvider();
  if (container) {
    auto it = container->find(name);
    if (it != container->end()) {
      return it->second;
    }
  }
  return nullptr;
}

namespace internal {
// Thread-local storage (TLS)
thread_local Scheduler* tls_sched = nullptr;
Scheduler& Scheduler::Current() {
  if (tls_sched == nullptr) tls_sched = new Scheduler;
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
