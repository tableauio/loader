#pragma once
#include <ctime>
#include <functional>
#include <memory>
#include <mutex>
#include <string>
#include <unordered_map>

#include "load.pc.h"
#include "scheduler.pc.h"

namespace tableau {
class MessagerContainer;
class Hub;

using MessagerMap = std::unordered_map<std::string, std::shared_ptr<Messager>>;
// FilterFunc filter in messagers if returned value is true.
// NOTE: name is the protobuf message name, e.g.: "message ItemConf{...}".
using Filter = std::function<bool(const std::string& name)>;
// MessagerContainerProvider provides a custom MessagerContainer for hub.
// If not specified, the hub's default MessagerContainer will be used.
// NOTE: This func must return non-nil MessagerContainer.
using MessagerContainerProvider = std::function<std::shared_ptr<MessagerContainer>(const Hub&)>;

struct HubOptions {
  // Filter can only filter in certain specific messagers based on the
  // condition that you provide.
  Filter filter;
  // Provide custom MessagerContainer. For keeping configuration access
  // consistent in a coroutine or a transaction.
  MessagerContainerProvider provider;
};

class Hub {
 public:
  Hub();

  // InitOnce inits the hub only once, and the subsequent calls will not take effect.
  void InitOnce(std::shared_ptr<const HubOptions> options);

  /***** Synchronous Loading *****/
  // Load fills messages (in MessagerContainer) from files in the specified directory and format.
  bool Load(const std::filesystem::path& dir, Format fmt = Format::kJSON,
            std::shared_ptr<const load::Options> options = nullptr);

  /***** Asynchronous Loading *****/
  // Load configs into temp MessagerContainer, and you should call LoopOnce() in you app's main loop,
  // in order to take the temp MessagerContainer into effect.
  bool AsyncLoad(const std::filesystem::path& dir, Format fmt = Format::kJSON,
                 std::shared_ptr<const load::Options> options = nullptr);
  int LoopOnce();
  // You'd better initialize the scheduler in the main thread.
  void InitScheduler();

  /***** MessagerMap *****/
  std::shared_ptr<MessagerMap> GetMessagerMap() const;
  void SetMessagerMap(std::shared_ptr<MessagerMap> msger_map);

  /***** MessagerContainer *****/
  // This function is exposed only for use in MessagerContainerProvider.
  std::shared_ptr<MessagerContainer> GetMessagerContainer() const { return msger_container_; }

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
  std::shared_ptr<MessagerMap> InternalLoad(const std::filesystem::path& dir, Format fmt = Format::kJSON,
                                            std::shared_ptr<const load::Options> options = nullptr) const;
  std::shared_ptr<MessagerMap> NewMessagerMap() const;
  std::shared_ptr<MessagerContainer> GetMessagerContainerWithProvider() const;
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const;

  bool Postprocess(std::shared_ptr<MessagerMap> msger_map);

 private:
  // For thread-safe guarantee during configuration updating.
  std::mutex mutex_;
  // All messagers' container.
  std::shared_ptr<MessagerContainer> msger_container_;
  // Loading scheduler.
  internal::Scheduler* sched_ = nullptr;
  // Init once
  std::once_flag init_once_;
  // Hub options
  std::shared_ptr<const HubOptions> options_;
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
{{ range .Protofiles }}{{ range .Messagers }}
class {{ . }};
template <>
const std::shared_ptr<{{ . }}> Hub::Get<{{ . }}>() const;
{{ end }}{{ end }}
class MessagerContainer {
  friend class Hub;

 public:
  MessagerContainer(std::shared_ptr<MessagerMap> msger_map = nullptr);

 private:
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const;
{{ range .Shards }}  void InitShard{{ . }}();
{{ end }}
 private:
  std::shared_ptr<MessagerMap> msger_map_;
  std::time_t last_loaded_time_;{{ range .Protofiles }}{{ range .Messagers }}
  std::shared_ptr<{{ . }}> {{ toSnake . }}_;{{ end }}{{ end }}
};

using MessagerGenerator = std::function<std::shared_ptr<Messager>()>;
// messager name -> messager generator
using Registrar = std::unordered_map<std::string, MessagerGenerator>;
class Registry {
  friend class Hub;

 public:
  static void Init();

  template <typename T>
  static void Register();
{{ if .Shards }}
 private:
{{ range .Shards }}  static void InitShard{{ . }}();
{{ end }}{{ end }}
 private:
  static std::once_flag once;
  static Registrar registrar;
};

template <typename T>
void Registry::Register() {
  registrar[T::Name()] = []() { return std::make_shared<T>(); };
}
}  // namespace tableau
