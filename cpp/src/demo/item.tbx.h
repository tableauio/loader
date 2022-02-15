#pragma once
#include "demo/hub.tbx.h"
#include "protoconf/item.pb.h"

namespace tableau {
class Item : public Messager {
 public:
  static const std::string& Name() { return kProtoName; };
  const protoconf::Item& Get() const { return data_; };
  virtual bool Load(const std::string& dir, Format fmt) override;

 private:
  static const std::string kProtoName;
  protoconf::Item data_;
};
}  // namespace tableau