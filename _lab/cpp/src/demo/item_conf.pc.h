#pragma once
#include <string>

#include "hub.pc.h"
#include "protoconf/item_conf.pb.h"

namespace tableau {
class ItemConf : public Messager {
 public:
  static const std::string& Name() { return kProtoName; };
  virtual bool Load(const std::string& dir, Format fmt) override;
  const protoconf::ItemConf& Data() const { return data_; };

 private:
  static const std::string kProtoName;
  protoconf::ItemConf data_;
};

}  // namespace tableau

namespace protoconf {
using ItemConfMgr = tableau::ItemConf;
}  // namespace protoconf
