#pragma once
#include "protoconf/hub.pc.h"
#include "protoconf/item_conf.pc.h"
class CustomItemConf : public tableau::Messager {
 public:
  static const std::string& Name() { return kCustomName; };
  virtual bool Load(const std::string& dir, tableau::Format fmt,
                    const tableau::LoadOptions* options = nullptr) override {
    return true;
  }
  const google::protobuf::Message& Message() const override { return special_item_conf_; }
  virtual bool ProcessAfterLoadAll(const tableau::Hub& hub) override;

 public:
  const std::string& GetSpecialItemName() const;

 private:
  static const std::string kCustomName;
  protoconf::ItemConf::Item special_item_conf_;
};