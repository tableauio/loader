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

class Hub : public tableau::Hub, public Singleton<Hub> {
 public:
  Hub() : tableau::Hub(GetOptions()) {}
  void Init();

 private:
  void InitCustomMessager();
  bool Filter(const std::string& name) { return true; }

 private:
  const tableau::HubOptions* GetOptions() {
    static const tableau::HubOptions options{std::bind(&Hub::Filter, this, std::placeholders::_1)};
    return &options;
  }
};
