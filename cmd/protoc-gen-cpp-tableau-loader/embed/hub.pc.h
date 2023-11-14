#pragma once
#include <google/protobuf/util/json_util.h>

#include <cstddef>
#include <functional>
#include <mutex>
#include <string>
#include <thread>
#include <unordered_map>

namespace tableau {
enum class Format {
  kUnknown,
  kJSON,
  kText,
  kBin,
};

extern const std::string kJSONExt;
extern const std::string kTextExt;
extern const std::string kBinExt;

static const std::string kEmpty = "";
const std::string& GetErrMsg();

class Messager;
class Hub;

using MessagerMap = std::unordered_map<std::string, std::shared_ptr<Messager>>;
using MessagerContainer = std::shared_ptr<MessagerMap>;
using Filter = std::function<bool(const std::string& name)>;
using MessagerContainerProvider = std::function<MessagerContainer()>;
using Postprocessor = std::function<bool(const Hub& hub)>;

struct LoadOptions {
  // Whether to ignore unknown JSON fields during parsing.
  //
  // Refer https://protobuf.dev/reference/cpp/api-docs/google.protobuf.util.json_util/#JsonParseOptions.
  bool ignore_unknown_fields = false;
  // Paths maps each messager name to a corresponding config file path.
  // If a messager name is existed, then the messager will be parsed from
  // the config file directly.
  // NOTE: only JSON, bin, and text formats are supported.
  std::unordered_map<std::string, std::string> paths;
  // postprocessor is called after loading all configurations.
  Postprocessor postprocessor;
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
bool Message2JSON(const google::protobuf::Message& message, std::string& json);
bool JSON2Message(const std::string& json, google::protobuf::Message& message, const LoadOptions* options = nullptr);
bool Text2Message(const std::string& text, google::protobuf::Message& message);
bool Bin2Message(const std::string& bin, google::protobuf::Message& message);
void ProtobufLogHandler(google::protobuf::LogLevel level, const char* filename, int line, const std::string& message);
const std::string& GetProtoName(const google::protobuf::Message& message);
bool LoadMessage(const std::string& dir, google::protobuf::Message& message, Format fmt = Format::kJSON,
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

bool Postprocess(Postprocessor postprocessor, MessagerContainer container);

}  // namespace internal

class Messager {
 public:
  virtual ~Messager() = default;
  static const std::string& Name() { return kEmpty; };
  virtual bool Load(const std::string& dir, Format fmt, const LoadOptions* options = nullptr) = 0;

 protected:
  // callback after this messager loaded.
  virtual bool ProcessAfterLoad() { return true; };

 public:
  // callback after all messagers loaded.
  virtual bool ProcessAfterLoadAll(const Hub& hub) { return true; };
};

class Hub {
 public:
  Hub() = default;
  Hub(MessagerContainer container) { SetMessagerContainer(container); }
  /***** Synchronous Loading *****/
  // Load messagers from dir using the specified format, and store them in MessagerContainer.
  bool Load(const std::string& dir, Filter filter = nullptr, Format fmt = Format::kJSON,
            const LoadOptions* options = nullptr);

  /***** Asynchronous Loading *****/
  // Load configs into temp MessagerContainer, and you should call LoopOnce() in you app's main loop,
  // in order to take the temp MessagerContainer into effect.
  bool AsyncLoad(const std::string& dir, Filter filter = nullptr, Format fmt = Format::kJSON,
                 const LoadOptions* options = nullptr);
  int LoopOnce();
  // You'd better initialize the scheduler in the main thread.
  void InitScheduler();

  /***** MessagerContainer *****/
  MessagerContainer GetMessagerContainer() const { return msger_container_; }
  void SetMessagerContainerProvider(MessagerContainerProvider provider) { msger_container_provider_ = provider; }

  /***** Access APIs *****/
  template <typename T>
  const std::shared_ptr<T> Get() const;

  template <typename T, typename U, typename... Args>
  const U* Get(Args... args) const;

  template <typename T, typename U, typename... Args>
  const U* GetOrderedMap(Args... args) const;

 private:
  MessagerContainer LoadNewMessagerContainer(const std::string& dir, Filter filter = nullptr,
                                             Format fmt = Format::kJSON, const LoadOptions* options = nullptr);
  MessagerContainer NewMessagerContainer(Filter filter = nullptr);
  void SetMessagerContainer(MessagerContainer msger_container);
  MessagerContainer GetMessagerContainerWithProvider() const;
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const;

 private:
  // For thread-safe guarantee during configuration updating.
  std::mutex mutex_;
  // All messagers' container.
  MessagerContainer msger_container_;
  // Provide custom MessagerContainer. For keeping configuration access
  // consistent in a coroutine or a transaction.
  MessagerContainerProvider msger_container_provider_;
  // Loading scheduler.
  internal::Scheduler* sched_ = nullptr;
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

namespace util {

// Combine hash values
//
// References:
//  - https://stackoverflow.com/questions/2590677/how-do-i-combine-hash-values-in-c0x
//  - https://stackoverflow.com/questions/17016175/c-unordered-map-using-a-custom-class-type-as-the-key
inline void HashCombine(std::size_t& seed) {}

template <typename T, typename... O>
inline void HashCombine(std::size_t& seed, const T& v, O... others) {
  std::hash<T> hasher;
  seed ^= hasher(v) + 0x9e3779b9 + (seed << 6) + (seed >> 2);
  HashCombine(seed, others...);
}

template <typename T, typename... O>
inline std::size_t SugaredHashCombine(const T& v, O... others) {
  std::size_t seed = 0;  // start with a hash value 0
  HashCombine(seed, v, others...);
  return seed;
}

// Mkdir makes dir recursively.
int Mkdir(const std::string& path);
// GetDir returns all but the last element of path, typically the path's
// directory.
std::string GetDir(const std::string& path);
// GetExt returns the file name extension used by path.
// The extension is the suffix beginning at the final dot
// in the final element of path; it is empty if there is
// no dot.
std::string GetExt(const std::string& path);

}  // namespace util

}  // namespace tableau
