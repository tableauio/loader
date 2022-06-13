package main

import (
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"google.golang.org/protobuf/compiler/protogen"
)

// generateHub generates related registry files.
func generateHub(gen *protogen.Plugin) {
	hppFilename := "hub." + pcExt + ".h"
	g1 := gen.NewGeneratedFile(hppFilename, "")
	helper.GenerateCommonHeader(gen, g1, version)
	g1.P()
	g1.P(hubHpp)

	cppFilename := "hub." + pcExt + ".cc"
	g2 := gen.NewGeneratedFile(cppFilename, "")
	helper.GenerateCommonHeader(gen, g2, version)
	g2.P()
	g2.P(hubCpp)
}

const hubHpp = `#pragma once
#include <google/protobuf/util/json_util.h>

#include <functional>
#include <mutex>
#include <string>
#include <thread>

namespace tableau {
enum class Format {
  kJSON,
  kText,
  kWire,
};

constexpr const char* kJSONExt = ".json";
constexpr const char* kTextExt = ".text";
constexpr const char* kWireExt = ".wire";

static const std::string kEmpty = "";
const std::string& GetErrMsg();

bool Message2JSON(const google::protobuf::Message& message, std::string& json);
bool JSON2Message(const std::string& json, google::protobuf::Message& message);
bool Text2Message(const std::string& text, google::protobuf::Message& message);
bool Wire2Message(const std::string& wire, google::protobuf::Message& message);

const std::string& GetProtoName(const google::protobuf::Message& message);
bool LoadMessage(const std::string& dir, google::protobuf::Message& message, Format fmt = Format::kJSON);

class Messager {
 public:
  virtual ~Messager() = default;
  static const std::string& Name() { return kEmpty; };
  virtual bool Load(const std::string& dir, Format fmt) = 0;

 protected:
  virtual bool ProcessAfterLoad() { return true; };
};

using MessagerMap = std::unordered_map<std::string, std::shared_ptr<Messager>>;
using MessagerMapPtr = std::shared_ptr<MessagerMap>;
using MMP = MessagerMapPtr;
using Filter = std::function<bool(const std::string& name)>;
using MMPProvider = std::function<MessagerMapPtr()>;  // MMP (MessagerMapPtr) provider.

class Hub {
 public:
  /***** Synchronously Loading *****/
  // Load messagers from dir using the specified format, and store them in MMP.
  bool Load(const std::string& dir, Filter filter = nullptr, Format fmt = Format::kJSON);

  /***** Asynchronously Loading *****/
  // Load configs into temp MMP, and you should call LoopOnce() in you app's main loop,
  // in order to take the temp MMP into effect.
  bool AsyncLoad(const std::string& dir, Filter filter, Format fmt = Format::kJSON);
  int LoopOnce();
  // You'd better initialize the scheduler in the main thread.
  void InitScheduler();

  /***** MMP: Messager Map Ptr *****/
  MessagerMapPtr GetMMP() const { return mmp_; }
  void SetMMPProvider(MMPProvider provider) { mmp_provider_ = provider; }

  /***** Access APIs *****/
  template <typename T>
  const std::shared_ptr<T> Get() const;

  template <typename T, typename U, typename... Args>
  const U* Get(Args... args) const;

  template <typename T, typename U, typename... Args>
  const U* GetOrderedMap(Args... args) const;

 private:
  MMP LoadNewMMP(const std::string& dir, Filter filter = nullptr, Format fmt = Format::kJSON);
  MessagerMapPtr NewMMP(Filter filter = nullptr);
  void SetMMP(MMP mmp);
  MessagerMapPtr GetMMPWithProvider() const;
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const { return (*GetMMPWithProvider())[name]; }

 private:
  // For thread-safe guarantee during configuration updating.
  std::mutex mutex_;
  // All messagers' container.
  MessagerMapPtr mmp_;
  // Provide custom MMP (MessagerMapPtr). For keeping configuration access consistency
  // in a coroutine or a transaction.
  MMPProvider mmp_provider_;
};

template <typename T>
const std::shared_ptr<T> Hub::Get() const {
  auto msg = GetMessager(T::Name());
  return std::dynamic_pointer_cast<T>(msg);
}

template <typename T, typename U, typename... Args>
const U* Hub::Get(Args... args) const {
  auto msg = GetMessager(T::Name());
  auto msger = std::dynamic_pointer_cast<T>(msg);
  return msger ? msger->Get(args...) : nullptr;
}

template <typename T, typename U, typename... Args>
const U* Hub::GetOrderedMap(Args... args) const {
  auto msg = GetMessager(T::Name());
  auto msger = std::dynamic_pointer_cast<T>(msg);
  return msger ? msger->GetOrderedMap(args...) : nullptr;
}

namespace internal {
class Scheduler {
 public:
  typedef std::function<void()> Job;

 public:
  Scheduler() : thread_id_(std::this_thread::get_id()) {}
  static Scheduler& Current();
  // thread-safety
  void Post(const Job& job);
  void Dispatch(const Job& job);
  int LoopOnce();
  bool IsLoopThread() const;
  void AssertInLoopThread() const;

 private:
  std::thread::id thread_id_;
  std::mutex mutex_;
  std::vector<Job> jobs_;
};

}  // namespace internal

}  // namespace tableau
`

const hubCpp = `#include "hub.pc.h"

#include <google/protobuf/text_format.h>

#include <fstream>
#include <sstream>

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
  auto mmp = LoadNewMMP(dir, filter, fmt);
  if (!mmp) {
    return false;
  }
  SetMMP(mmp);
  return true;
}

bool Hub::AsyncLoad(const std::string& dir, Filter filter, Format fmt) {
  auto mmp = LoadNewMMP(dir, filter, fmt);
  if (!mmp) {
    return false;
  }
  ::tableau::internal::Scheduler::Current().Dispatch(std::bind(&Hub::SetMMP, this, mmp));
  return true;
}

int Hub::LoopOnce() { return ::tableau::internal::Scheduler::Current().LoopOnce(); }
void Hub::InitScheduler() { ::tableau::internal::Scheduler::Current(); }

MMP Hub::LoadNewMMP(const std::string& dir, Filter filter, Format fmt) {
  auto mmp = NewMMP(filter);
  for (auto iter : *mmp) {
    auto&& name = iter.first;
    bool ok = iter.second->Load(dir, fmt);
    if (!ok) {
      return nullptr;
    }
  }
  return mmp;
}

MessagerMapPtr Hub::NewMMP(Filter filter) {
  MessagerMapPtr mmp = std::make_shared<MessagerMap>();
  for (auto&& it : Registry::registrar) {
    if (filter == nullptr || filter(it.first)) {
      (*mmp)[it.first] = it.second();
    }
  }
  return mmp;
}

void Hub::SetMMP(MMP mmp) {
  // replace with thread-safe guarantee.
  std::unique_lock<std::mutex> lock(mutex_);
  mmp_ = mmp;
}

MessagerMapPtr Hub::GetMMPWithProvider() const {
  if (mmp_provider_ != nullptr) {
    return mmp_provider_();
  }
  return mmp_;
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
  {  // scoped for auto-release lock.
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
`
