#include "hub/custom/item/custom_item_conf.h"

#include "logger.pc.h"

const std::string CustomItemConf::kCustomName = "CustomItemConf";

bool CustomItemConf::ProcessAfterLoadAll(const tableau::Hub& hub) {
  auto conf = hub.Get<protoconf::ItemConfMgr, protoconf::ItemConf::Item>(1);
  if (!conf) {
    ATOM_ERROR("hub get item 1 failed!");
    return false;
  }
  special_item_conf_ = *conf;  // value copy
  ATOM_DEBUG("custom item conf processed");
  return true;
}

const std::string& CustomItemConf::GetSpecialItemName() const { return special_item_conf_.name(); }