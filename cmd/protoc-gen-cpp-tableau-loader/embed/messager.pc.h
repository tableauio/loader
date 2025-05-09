#pragma once
#include <google/protobuf/message.h>

#include <chrono>
#include <functional>
#include <string>

#include "util.pc.h"

namespace tableau {
class Hub;
struct LoadOptions;

class Messager {
 public:
  struct Stats {
    std::chrono::microseconds duration;  // total load time consuming.
  };

 public:
  virtual ~Messager() = default;
  static const std::string& Name() { return kEmpty; }
  const Stats& GetStats() { return stats_; }
  // Load fills message from file in the specified directory and format.
  virtual bool Load(const std::string& dir, Format fmt, const LoadOptions* options = nullptr) = 0;
  // Message returns the inner message data.
  virtual const google::protobuf::Message* Message() const { return nullptr; }
  // callback after all messagers loaded.
  virtual bool ProcessAfterLoadAll(const Hub& hub) { return true; }

 protected:
  // callback after this messager loaded.
  virtual bool ProcessAfterLoad() { return true; };
  Stats stats_;
};
}  // namespace tableau