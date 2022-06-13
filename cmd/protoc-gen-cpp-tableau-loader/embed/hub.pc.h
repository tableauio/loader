#pragma once
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
  internal::Scheduler* sched_;
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

}  // namespace tableau
