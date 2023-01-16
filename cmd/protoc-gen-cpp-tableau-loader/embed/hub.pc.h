#pragma once
#include <google/protobuf/util/json_util.h>

#include <cstddef>
#include <functional>
#include <mutex>
#include <string>
#include <thread>

namespace tableau {
enum class Format {
  kUnknown,
  kJSON,
  kText,
  kBin,
};

constexpr const char* kJSONExt = ".json";
constexpr const char* kTextExt = ".txt";
constexpr const char* kBinExt = ".bin";

static const std::string kEmpty = "";
const std::string& GetErrMsg();
struct LoadOptions {
  // Whether to ignore unknown JSON fields during parsing.
  //
  // https://developers.google.com/protocol-buffers/docs/reference/cpp/google.protobuf.util.json_util#JsonParseOptions.
  bool ignore_unknown_fields;
};

Format Ext2Format(const std::string& ext);
// Empty string will be returned if an unsupported enum value has been passed,
// and the error message could be obtained by GetErrMsg().
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

}  // namespace internal

class Messager {
 public:
  virtual ~Messager() = default;
  static const std::string& Name() { return kEmpty; };
  virtual bool Load(const std::string& dir, Format fmt, const LoadOptions* options = nullptr) = 0;

 protected:
  virtual bool ProcessAfterLoad() { return true; };
};

using MessagerMap = std::unordered_map<std::string, std::shared_ptr<Messager>>;
using MessagerContainer = std::shared_ptr<MessagerMap>;
using Filter = std::function<bool(const std::string& name)>;
using MessagerContainerProvider = std::function<MessagerContainer()>;

class Hub {
 public:
  /***** Synchronously Loading *****/
  // Load messagers from dir using the specified format, and store them in MessagerContainer.
  bool Load(const std::string& dir, Filter filter = nullptr, Format fmt = Format::kJSON,
            const LoadOptions* options = nullptr);

  /***** Asynchronously Loading *****/
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

}  // namespace util

}  // namespace tableau
