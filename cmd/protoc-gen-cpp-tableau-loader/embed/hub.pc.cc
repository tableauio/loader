#include "hub.pc.h"

#include <google/protobuf/stubs/logging.h>
#include <google/protobuf/stubs/status.h>
#include <google/protobuf/text_format.h>

#include <fstream>
#include <sstream>
#include <unordered_map>

#include "logger.pc.h"
#include "registry.pc.h"

namespace tableau {
static thread_local std::string g_err_msg;
const std::string& GetErrMsg() { return g_err_msg; }

const std::string kJSONExt = ".json";
const std::string kTextExt = ".txt";
const std::string kBinExt = ".bin";

Format Ext2Format(const std::string& ext) {
  if (ext == kJSONExt) {
    return Format::kJSON;
  }
  if (ext == kTextExt) {
    return Format::kText;
  }
  if (ext == kBinExt) {
    return Format::kBin;
  }
  return Format::kUnknown;
}

const std::string& Format2Ext(Format fmt) {
  switch (fmt) {
    case Format::kJSON:
      return kJSONExt;
    case Format::kText:
      return kTextExt;
    case Format::kBin:
      return kBinExt;
    default:
      g_err_msg = "unsupported format: " + std::to_string(static_cast<int>(fmt));
      return kEmpty;
  }
}

// refer: https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/stubs/logging.h
void ProtobufLogHandler(google::protobuf::LogLevel level, const char* filename, int line, const std::string& message) {
  static const std::unordered_map<int, log::Level> kLevelMap = {{google::protobuf::LOGLEVEL_INFO, log::kInfo},
                                                                {google::protobuf::LOGLEVEL_WARNING, log::kWarn},
                                                                {google::protobuf::LOGLEVEL_ERROR, log::kError},
                                                                {google::protobuf::LOGLEVEL_FATAL, log::kFatal}};
  log::Level lvl = log::kWarn;  // default
  auto iter = kLevelMap.find(level);
  if (iter != kLevelMap.end()) {
    lvl = iter->second;
  }
  ATOM_LOGGER_CALL(tableau::log::DefaultLogger(), lvl, "[libprotobuf %s:%d] %s", filename, line, message.c_str());
}

bool Message2JSON(const google::protobuf::Message& message, std::string& json) {
  google::protobuf::util::JsonPrintOptions options;
  options.add_whitespace = true;
  options.always_print_primitive_fields = true;
  options.preserve_proto_field_names = true;
  return google::protobuf::util::MessageToJsonString(message, &json, options).ok();
}

bool JSON2Message(const std::string& json, google::protobuf::Message& message,
                  const LoadOptions* options /* = nullptr */) {
  google::protobuf::util::Status status;
  if (options != nullptr) {
    google::protobuf::util::JsonParseOptions parse_options;
    parse_options.ignore_unknown_fields = options->ignore_unknown_fields;
    status = google::protobuf::util::JsonStringToMessage(json, &message, parse_options);
  } else {
    status = google::protobuf::util::JsonStringToMessage(json, &message);
  }
  if (!status.ok()) {
    g_err_msg = "failed to parse " + GetProtoName(message) + kJSONExt + ": " + status.ToString();
    return false;
  }
  return true;
}

bool Text2Message(const std::string& text, google::protobuf::Message& message) {
  if (!google::protobuf::TextFormat::ParseFromString(text, &message)) {
    g_err_msg = "failed to parse " + GetProtoName(message) + kTextExt;
    return false;
  }
  return true;
}
bool Bin2Message(const std::string& bin, google::protobuf::Message& message) {
  if (!message.ParseFromString(bin)) {
    g_err_msg = "failed to parse " + GetProtoName(message) + kBinExt;
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
    g_err_msg = "failed to open " + filename + ": " + strerror(errno);
    return false;
  }
  std::stringstream ss;
  ss << file.rdbuf();
  content = std::move(ss.str());
  return true;
}

bool LoadMessage(const std::string& dir, google::protobuf::Message& message, Format fmt,
                 const LoadOptions* options /* = nullptr*/) {
  message.Clear();
  std::string path;
  std::string name = GetProtoName(message);
  if (options) {
    auto iter = options->paths.find(name);
    if (iter != options->paths.end()) {
      // path specified explicitly, then use it directly
      path = iter->second;
      fmt = Ext2Format(util::GetExt(iter->second));
    }
  }
  // Support 3 formats: json, text, and bin
  if (path.empty()) {
    switch (fmt) {
      case Format::kJSON: {
        path = dir + name + kJSONExt;
        break;
      }
      case Format::kText: {
        path = dir + name + kTextExt;
        break;
      }
      case Format::kBin: {
        path = dir + name + kBinExt;
        break;
      }
      default: {
        g_err_msg = "unsupported format: %d" + static_cast<int>(fmt);
        return false;
      }
    }
  }
  std::string content;
  bool ok = ReadFile(path, content);
  if (!ok) {
    return false;
  }
  switch (fmt) {
    case Format::kJSON: {
      return JSON2Message(content, message, options);
    }
    case Format::kText: {
      return Text2Message(content, message);
    }
    case Format::kBin: {
      return Bin2Message(content, message);
    }
    default: {
      g_err_msg = "unsupported format: %d" + static_cast<int>(fmt);
      return false;
    }
  }
}

bool StoreMessage(const std::string& dir, google::protobuf::Message& message, Format fmt) {
  // TODO: write protobuf message to file, support 3 formats: json, text, and bin.
  return false;
}

bool Hub::Load(const std::string& dir, Filter filter /* = nullptr */, Format fmt /* = Format::kJSON */,
               const LoadOptions* options /* = nullptr */) {
  auto msger_container = LoadNewMessagerContainer(dir, filter, fmt, options);
  if (!msger_container) {
    return false;
  }
  bool ok = internal::Postprocess(options->postprocessor, msger_container);
  if (!ok) {
    return false;
  }
  SetMessagerContainer(msger_container);
  return true;
}

bool Hub::AsyncLoad(const std::string& dir, Filter filter /* = nullptr */, Format fmt /* = Format::kJSON */,
                    const LoadOptions* options /* = nullptr */) {
  auto msger_container = LoadNewMessagerContainer(dir, filter, fmt, options);
  if (!msger_container) {
    return false;
  }
  bool ok = internal::Postprocess(options->postprocessor, msger_container);
  if (!ok) {
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

MessagerContainer Hub::LoadNewMessagerContainer(const std::string& dir, Filter filter /* = nullptr */,
                                                Format fmt /* = Format::kJSON */,
                                                const LoadOptions* options /* = nullptr */) {
  // intercept protobuf error logs
  auto old_handler = google::protobuf::SetLogHandler(ProtobufLogHandler);

  auto msger_container = NewMessagerContainer(filter);
  for (auto iter : *msger_container) {
    auto&& name = iter.first;
    ATOM_DEBUG("loading %s", name.c_str());
    bool ok = iter.second->Load(dir, fmt, options);
    if (!ok) {
      ATOM_ERROR("load %s failed: %s", name.c_str(), GetErrMsg().c_str());
      // restore to old protobuf log hanlder
      google::protobuf::SetLogHandler(old_handler);
      return nullptr;
    }
    ATOM_DEBUG("loaded %s", name.c_str());
  }

  // restore to old protobuf log hanlder
  google::protobuf::SetLogHandler(old_handler);
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

bool Postprocess(Postprocessor postprocessor, MessagerContainer container) {
  // create a temporary hub with messager container for post process
  Hub hub(container);

  // messager-level postprocess
  for (auto iter : *container) {
    auto msger = iter.second;
    bool ok = msger->ProcessAfterLoadAll(hub);
    if (!ok) {
      g_err_msg = "hub call ProcessAfterLoadAll failed, messager: " + msger->Name();
      return false;
    }
  }

  // hub-level postprocess
  if (postprocessor != nullptr) {
    bool ok = postprocessor(hub);
    if (!ok) {
      g_err_msg = "hub call Postprocesser failed, you'd better check your custom 'postprocessor' load option";
      return false;
    }
  }

  return true;
}

}  // namespace internal

namespace util {
int Mkdir(const std::string& path) {
  std::string path_ = path + "/";
  struct stat info;
  for (size_t pos = path_.find('/', 0); pos != std::string::npos; pos = path_.find('/', pos)) {
    ++pos;
    auto sub_dir = path_.substr(0, pos);
    if (stat(sub_dir.c_str(), &info) == 0 && info.st_mode & S_IFDIR) {
      continue;
    }
    int status = mkdir(sub_dir.c_str(), 0755);
    if (status != 0) {
      std::cerr << "system error: " << strerror(errno) << std::endl;
      return -1;
    }
  }
  return 0;
}

std::string GetDir(const std::string& path) {
  std::size_t pos = path.find_last_of("/\\");
  if (pos != std::string::npos) {
    return path.substr(0, pos);
  }
  return kEmpty;
}

std::string GetExt(const std::string& path) {
  std::size_t pos = path.find_last_of(".");
  if (pos != std::string::npos) {
    return path.substr(pos);
  }
  return kEmpty;
}

}  // namespace util

}  // namespace tableau
