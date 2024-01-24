#include "hub/custom/item/custom_item_conf.h"

const std::string CustomItemConf::kCustomName = "CustomItemConf";

bool CustomItemConf::ProcessAfterLoadAll(const tableau::Hub& hub) {
  auto conf = hub.Get<protoconf::ItemConfMgr, protoconf::ItemConf::Item>(1);
  if (!conf) {
    std::cout << "hub get item 1 failed!" << std::endl;
    return false;
  }
  special_item_conf_ = *conf;  // value copy
  std::cout << "custom item conf processed" << std::endl;
  return true;
}

const std::string& CustomItemConf::GetSpecialItemName() const { return special_item_conf_.name(); }