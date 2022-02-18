#pragma once
#include <string>

#include "hub.pc.h"
#include "protoconf/item.pb.h"

namespace tableau {
class Item : public Messager {
 public:
  static const std::string& Name() { return kProtoName; };
  virtual bool Load(const std::string& dir, Format fmt) override;
  const protoconf::Item& Data() const { return data_; };

 private:
  static const std::string kProtoName;
  protoconf::Item data_;
};

}  // namespace tableau
