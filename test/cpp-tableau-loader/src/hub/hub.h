#pragma once
#include "protoconf/hub.pc.h"

template <class T, const bool threaded = false>
class Singleton {
 private:
  Singleton(const T&) = delete;
  Singleton(T&&) = delete;
  void operator=(const T&) = delete;
  static inline T* GetInstancePtr() {
    T* ptr = nullptr;
    if (threaded) {
      static thread_local T* new_ptr = new T();
      ptr = new_ptr;
    } else {
      static T* new_ptr = new T();
      ptr = new_ptr;
    }
    assert(ptr != nullptr);
    return ptr;
  }

 public:
  static inline T& Instance() { return *GetInstancePtr(); }

 protected:
  Singleton() = default;
};

template <class T>
class HubBase : public tableau::Hub, public Singleton<HubBase<T>> {
 public:
  HubBase() : tableau::Hub(T::GetOptions()) { T::Init(); }
};

class DefaultHubOptions {
 public:
  static const std::shared_ptr<tableau::HubOptions> GetOptions();
  static void Init();

 private:
  static bool Filter(const std::string& name);
  static std::shared_ptr<tableau::MessagerContainer> MessagerContainerProvider();
  static void InitCustomMessager();
};

using Hub = HubBase<DefaultHubOptions>;
