#pragma once
#include <string>

#include "hub.pc.h"
#include "protoconf/test_conf.pb.h"

namespace tableau {
class ActivityConf : public Messager {
 public:
  static const std::string& Name() { return kProtoName; };
  virtual bool Load(const std::string& dir, Format fmt) override;
  const protoconf::ActivityConf& Data() const { return data_; };
  const protoconf::ActivityConf::Activity* Get(uint64_t key1) const;
  const protoconf::ActivityConf::Activity::Chapter* Get(uint64_t key1, uint32_t key2) const;
  const protoconf::Section* Get(uint64_t key1, uint32_t key2, uint32_t key3) const;

 private:
  static const std::string kProtoName;
  protoconf::ActivityConf data_;
};

}  // namespace tableau
