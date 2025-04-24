#pragma once
#include <google/protobuf/message.h>
#include <google/protobuf/util/json_util.h>
#include <tableau/protobuf/tableau.pb.h>

#include <algorithm>
#include <chrono>
#include <cstddef>
#include <ctime>
#include <functional>
#include <map>
#include <mutex>
#include <string>
#include <thread>
#include <unordered_map>
#include <vector>

#include "messager.pc.h"

namespace tableau {
extern const std::string kUnknownExt;
extern const std::string kJSONExt;
extern const std::string kTextExt;
extern const std::string kBinExt;

const std::string& GetErrMsg();

class MessagerContainer;
class Hub;

// Auto-generated declarations below

using MessagerMap = std::unordered_map<std::string, std::shared_ptr<Messager>>;
// FilterFunc filter in messagers if returned value is true.
// NOTE: name is the protobuf message name, e.g.: "message ItemConf{...}".
using Filter = std::function<bool(const std::string& name)>;
using MessagerContainerProvider = std::function<std::shared_ptr<MessagerContainer>()>;

struct HubOptions {
  // Filter can only filter in certain specific messagers based on the
  // condition that you provide.
  Filter filter;
  // Provide custom MessagerContainer. For keeping configuration access
  // consistent in a coroutine or a transaction.
  MessagerContainerProvider provider;
};

// Convert file extension to Format type.
// NOTE: ext includes dot ".", such as:
//  - kJSONExt：".json"
//  - kTextExt".txt"
//  - kBinExt".bin"
Format Ext2Format(const std::string& ext);
// Empty string will be returned if an unsupported enum value has been passed,
// and the error message can be obtained by GetErrMsg().
const std::string& Format2Ext(Format fmt);
bool Message2JSON(const google::protobuf::Message& msg, std::string& json);
bool JSON2Message(const std::string& json, google::protobuf::Message& msg, const LoadOptions* options = nullptr);
bool Text2Message(const std::string& text, google::protobuf::Message& msg);
bool Bin2Message(const std::string& bin, google::protobuf::Message& msg);
void ProtobufLogHandler(google::protobuf::LogLevel level, const char* filename, int line, const std::string& msg);
const std::string& GetProtoName(const google::protobuf::Message& msg);
bool LoadMessageByPath(google::protobuf::Message& msg, const std::string& path, Format fmt = Format::kJSON,
                       const LoadOptions* options = nullptr);
bool LoadMessage(google::protobuf::Message& msg, const std::string& dir, Format fmt = Format::kJSON,
                 const LoadOptions* options = nullptr);

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

bool Postprocess(Postprocessor postprocessor, std::shared_ptr<MessagerMap> msger_map);

}  // namespace internal

class Hub {
 public:
  Hub(const HubOptions* options = nullptr)
      : msger_container_(std::make_shared<MessagerContainer>()), options_(options ? *options : HubOptions{}) {}
  /***** Synchronous Loading *****/
  // Load fills messages (in MessagerContainer) from files in the specified directory and format.
  bool Load(const std::string& dir, Format fmt = Format::kJSON, const LoadOptions* options = nullptr);

  /***** Asynchronous Loading *****/
  // Load configs into temp MessagerContainer, and you should call LoopOnce() in you app's main loop,
  // in order to take the temp MessagerContainer into effect.
  bool AsyncLoad(const std::string& dir, Format fmt = Format::kJSON, const LoadOptions* options = nullptr);
  int LoopOnce();
  // You'd better initialize the scheduler in the main thread.
  void InitScheduler();

  /***** MessagerMap *****/
  std::shared_ptr<MessagerMap> GetMessagerMap() const;
  void SetMessagerMap(std::shared_ptr<MessagerMap> msger_map);

  /***** MessagerContainer *****/
  // This function is exposed only for use in MessagerContainerProvider.
  std::shared_ptr<MessagerContainer> GetMessagerContainer() const {
    if (options_.provider != nullptr) {
      return options_.provider();
    }
    return msger_container_;
  }

  /***** Access APIs *****/
  template <typename T>
  const std::shared_ptr<T> Get() const;

  template <typename T, typename U, typename... Args>
  const U* Get(Args... args) const;

  template <typename T, typename U, typename... Args>
  const U* GetOrderedMap(Args... args) const;

  // GetLastLoadedTime returns the time when hub's msger_container_ was last set.
  inline std::time_t GetLastLoadedTime() const;

 private:
  std::shared_ptr<MessagerMap> InternalLoad(const std::string& dir, Format fmt = Format::kJSON,
                                            const LoadOptions* options = nullptr) const;
  std::shared_ptr<MessagerMap> NewMessagerMap() const;
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const;

 private:
  // For thread-safe guarantee during configuration updating.
  std::mutex mutex_;
  // All messagers' container.
  std::shared_ptr<MessagerContainer> msger_container_;
  // Loading scheduler.
  internal::Scheduler* sched_ = nullptr;
  // Hub options
  const HubOptions options_;
};

template <typename T>
const std::shared_ptr<T> Hub::Get() const {
  auto msg = GetMessager(T::Name());
  return std::dynamic_pointer_cast<T>(msg);
}

template <typename T, typename U, typename... Args>
const U* Hub::Get(Args... args) const {
  auto msger = Get<T>();
  return msger ? msger->Get(args...) : nullptr;
}

template <typename T, typename U, typename... Args>
const U* Hub::GetOrderedMap(Args... args) const {
  auto msger = Get<T>();
  return msger ? msger->GetOrderedMap(args...) : nullptr;
}

// Auto-generated specializations below
class MessagerContainer {
 public:
  MessagerContainer(std::shared_ptr<MessagerMap> msger_map = nullptr);

 public:
  std::shared_ptr<MessagerMap> msger_map_;
  std::time_t last_loaded_time_;
  // Auto-generated fields below
};

}  // namespace tableau