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

const std::string kUnknownExt = ".unknown";
const std::string kJSONExt = ".json";
const std::string kTextExt = ".txt";
const std::string kBinExt = ".bin";

Format Ext2Format(const std::string& ext) {
  if (ext == kJSONExt) {
    return Format::kJSON;
  } else if (ext == kTextExt) {
    return Format::kText;
  } else if (ext == kBinExt) {
    return Format::kBin;
  } else {
    return Format::kUnknown;
  }
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
      return kUnknownExt;
  }
}

// refer: https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/stubs/logging.h
void ProtobufLogHandler(google::protobuf::LogLevel level, const char* filename, int line, const std::string& msg) {
  static const std::unordered_map<int, log::Level> kLevelMap = {{google::protobuf::LOGLEVEL_INFO, log::kInfo},
                                                                {google::protobuf::LOGLEVEL_WARNING, log::kWarn},
                                                                {google::protobuf::LOGLEVEL_ERROR, log::kError},
                                                                {google::protobuf::LOGLEVEL_FATAL, log::kFatal}};
  log::Level lvl = log::kWarn;  // default
  auto iter = kLevelMap.find(level);
  if (iter != kLevelMap.end()) {
    lvl = iter->second;
  }
  ATOM_LOGGER_CALL(tableau::log::DefaultLogger(), lvl, "[libprotobuf %s:%d] %s", filename, line, msg.c_str());
}

bool Message2JSON(const google::protobuf::Message& msg, std::string& json) {
  google::protobuf::util::JsonPrintOptions options;
  options.add_whitespace = true;
  options.always_print_primitive_fields = true;
  options.preserve_proto_field_names = true;
  return google::protobuf::util::MessageToJsonString(msg, &json, options).ok();
}

bool JSON2Message(const std::string& json, google::protobuf::Message& msg, const LoadOptions* options /* = nullptr */) {
  google::protobuf::util::Status status;
  if (options != nullptr) {
    google::protobuf::util::JsonParseOptions parse_options;
    parse_options.ignore_unknown_fields = options->ignore_unknown_fields;
    status = google::protobuf::util::JsonStringToMessage(json, &msg, parse_options);
  } else {
    status = google::protobuf::util::JsonStringToMessage(json, &msg);
  }
  if (!status.ok()) {
    g_err_msg = "failed to parse " + GetProtoName(msg) + kJSONExt + ": " + status.ToString();
    return false;
  }
  return true;
}

bool Text2Message(const std::string& text, google::protobuf::Message& msg) {
  if (!google::protobuf::TextFormat::ParseFromString(text, &msg)) {
    g_err_msg = "failed to parse " + GetProtoName(msg) + kTextExt;
    return false;
  }
  return true;
}
bool Bin2Message(const std::string& bin, google::protobuf::Message& msg) {
  if (!msg.ParseFromString(bin)) {
    g_err_msg = "failed to parse " + GetProtoName(msg) + kBinExt;
    return false;
  }
  return true;
}

// MergeMessage merges patch into main message, which must be the same descriptor.
//
// 1. top none-map field (field value):
//   - `replace`: if field present in both **main** and **patch** sheet
//
// 2. top map field patch (key-value pair):
//   - `add`: if key not exists in **main** sheet
//   - `replace`: if key exists in both **main** and **patch** sheet
bool MergeMessage(google::protobuf::Message& main, google::protobuf::Message& patch) {
  // Ensure both messages are of the same type
  if (main.GetDescriptor() != patch.GetDescriptor()) {
    std::cerr << "Messages are not of the same type" << std::endl;
    return false;
  }

  // Get the reflection and descriptor for the messages
  const google::protobuf::Reflection* main_reflection = main.GetReflection();
  const google::protobuf::Reflection* patch_reflection = patch.GetReflection();
  const google::protobuf::Descriptor* descriptor = main.GetDescriptor();

  // Iterate over all fields in the message
  for (int i = 0; i < descriptor->field_count(); ++i) {
    const google::protobuf::FieldDescriptor* field = descriptor->field(i);

    if (field->is_repeated() && patch_reflection->FieldSize(patch, field) > 0) {
      // if repeated field is set in patch message, then clear the repeated field
      // in the main message ahead.
      main_reflection->ClearField(&main, field);
    }
  }

  // 1. Singular fields will be overwritten, if specified in from, except for
  //    embedded messages which will be merged.
  // 2. Repeated fields will be concatenated.
  //
  // References:
  //  - https://protobuf.dev/reference/cpp/api-docs/google.protobuf.message/#Reflection
  //  - https://protobuf.dev/reference/cpp/api-docs/google.protobuf.descriptor/#FieldDescriptor
  //  - https://protobuf.dev/reference/cpp/api-docs/google.protobuf.message/#Message.MergeFrom.details
  main.MergeFrom(patch);
  return true;
}

const std::string& GetProtoName(const google::protobuf::Message& msg) {
  const auto* md = msg.GetDescriptor();
  return md != nullptr ? md->name() : kEmpty;
}

bool ExistsFile(const std::string& filename) {
  std::ifstream file(filename);
  // returns true if the file exists and is accessible
  return file.good();
}

bool ReadFile(const std::string& filename, std::string& content) {
  std::ifstream file(filename);
  if (!file.is_open()) {
    g_err_msg = "failed to open " + filename + ": " + strerror(errno);
    return false;
  }
  std::stringstream ss;
  ss << file.rdbuf();
  content = ss.str();
  return true;
}

bool LoadMessageWithPatch(google::protobuf::Message& msg, const std::string& path, Format fmt, tableau::Patch patch,
                          const LoadOptions* options /* = nullptr*/) {
  std::string name = GetProtoName(msg);
  std::string patch_path;
  Format patch_fmt = fmt;
  if (options) {
    auto iter = options->patch_paths.find(name);
    if (iter != options->patch_paths.end()) {
      // patch_path specified in PatchPaths, then use it instead of PatchDir.
      patch_path = iter->second;
      patch_fmt = Ext2Format(util::GetExt(iter->second));
    }
  }
  if (path.empty()) {
    if (!options || options->patch_dir.empty()) {
      // patch_dir not provided, then just load from file in the "main" dir.
      return LoadMessageByPath(msg, path, fmt, options);
    }
    patch_path = options->patch_dir + name + Format2Ext(fmt);
  }
  if (!ExistsFile(patch_path)) {
    // If patch file not exists, then just load from the "main" file.
    return LoadMessageByPath(msg, path, fmt, options);
  }
  switch (patch) {
    case tableau::PATCH_REPLACE: {
      return LoadMessageByPath(msg, patch_path, patch_fmt, options);
    }
    case tableau::PATCH_MERGE: {
      // Create a new instance of the same type as the original message
      google::protobuf::Message& patch_msg = *(msg.New());
      // load msg from the "main" file
      if (!LoadMessageByPath(msg, path, fmt, options)) {
        return false;
      }
      // load patch_msg from the "patch" file
      if (!LoadMessageByPath(patch_msg, patch_path, patch_fmt, options)) {
        return false;
      }
      return MergeMessage(msg, patch_msg);
    }
    default: {
      g_err_msg = "unknown patch type: " + std::to_string(static_cast<int>(patch));
      return false;
    }
  }
}

bool LoadMessageByPath(google::protobuf::Message& msg, const std::string& path, Format fmt,
                       const LoadOptions* options /* = nullptr*/) {
  std::string content;
  bool ok = ReadFile(path, content);
  if (!ok) {
    return false;
  }
  switch (fmt) {
    case Format::kJSON: {
      return JSON2Message(content, msg, options);
    }
    case Format::kText: {
      return Text2Message(content, msg);
    }
    case Format::kBin: {
      return Bin2Message(content, msg);
    }
    default: {
      g_err_msg = "unknown format: " + std::to_string(static_cast<int>(fmt));
      return false;
    }
  }
}

bool LoadMessage(google::protobuf::Message& msg, const std::string& dir, Format fmt,
                 const LoadOptions* options /* = nullptr*/) {
  std::string name = GetProtoName(msg);
  std::string path;
  if (options) {
    auto iter = options->paths.find(name);
    if (iter != options->paths.end()) {
      // path specified in Paths, then use it instead of dir.
      path = iter->second;
      fmt = Ext2Format(util::GetExt(iter->second));
    }
  }
  if (path.empty()) {
    path = dir + name + Format2Ext(fmt);
  }

  const google::protobuf::Descriptor* descriptor = msg.GetDescriptor();
  if (!descriptor) {
    g_err_msg = "failed to get descriptor of message: " + name;
    return false;
  }
  // access the extension directly using the generated identifier
  const tableau::WorksheetOptions worksheet_options = descriptor->options().GetExtension(tableau::worksheet);
  if (worksheet_options.patch() != tableau::PATCH_NONE) {
    return LoadMessageWithPatch(msg, path, fmt, worksheet_options.patch(), options);
  }

  return LoadMessageByPath(msg, path, fmt, options);
}

bool StoreMessage(google::protobuf::Message& msg, const std::string& dir, Format fmt) {
  // TODO: write protobuf message to file, support 3 formats: json, text, and bin.
  return false;
}

bool Hub::Load(const std::string& dir, Format fmt /* = Format::kJSON */, const LoadOptions* options /* = nullptr */) {
  auto msger_container = LoadNewMessagerContainer(dir, fmt, options);
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

bool Hub::AsyncLoad(const std::string& dir, Format fmt /* = Format::kJSON */,
                    const LoadOptions* options /* = nullptr */) {
  auto msger_container = LoadNewMessagerContainer(dir, fmt, options);
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

MessagerContainer Hub::LoadNewMessagerContainer(const std::string& dir, Format fmt /* = Format::kJSON */,
                                                const LoadOptions* options /* = nullptr */) {
  // intercept protobuf error logs
  auto old_handler = google::protobuf::SetLogHandler(ProtobufLogHandler);
  Filter filter = options != nullptr ? options->filter : nullptr;
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
  last_loaded_time_ = std::time(nullptr);
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
