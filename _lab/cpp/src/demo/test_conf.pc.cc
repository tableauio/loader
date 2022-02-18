#include "test_conf.pc.h"

namespace tableau {
const std::string ActivityConf::kProtoName = "ActivityConf";
bool ActivityConf::Load(const std::string& dir, Format fmt) { return LoadMessage(dir, data_, fmt); }
const protoconf::ActivityConf::Activity* ActivityConf::Get(uint64_t key1) const {
  auto iter = data_.activity_map().find(key1);
  if (iter == data_.activity_map().end()) {
    return nullptr;
  }
  return &iter->second;
}

const protoconf::ActivityConf::Activity::Chapter* ActivityConf::Get(uint64_t key1, uint32_t key2) const {
  const auto* conf = Get(key1);
  if (conf == nullptr) {
    return nullptr;
  }
  auto iter = conf->chapter_map().find(key2);
  if (iter == conf->chapter_map().end()) {
    return nullptr;
  }
  return &iter->second;
}

const protoconf::Section* ActivityConf::Get(uint64_t key1, uint32_t key2, uint32_t key3) const {
  const auto* conf = Get(key1, key2);
  if (conf == nullptr) {
    return nullptr;
  }
  auto iter = conf->section_map().find(key3);
  if (iter == conf->section_map().end()) {
    return nullptr;
  }
  return &iter->second;
}

}  // namespace tableau
