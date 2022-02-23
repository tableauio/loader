#include "test_conf.pc.h"

namespace tableau {
const std::string ActivityConf::kProtoName = "ActivityConf";
bool ActivityConf::Load(const std::string& dir, Format fmt) {
  bool ok = LoadMessage(dir, data_, fmt);
  if (!ok) {
    return false;
  }

  for (auto&& item1 : data_.activity_map()) {
    std::cout << "item1: " << item1.first << std::endl;
    ordered_map_[item1.first] =
        Activity_OrderedMapValueType(Activity_Chapter_OrderedMap(), &item1.second.chapter_map());
    auto&& ordered_map1 = ordered_map_[item1.first].first;
    for (auto&& item2 : item1.second.chapter_map()) {
      std::cout << "item2: " << item2.first << std::endl;
      ordered_map1[item2.first] =
          Activity_Chapter_OrderedMapValueType(Section_OrderedMap(), &item2.second.section_map());
      auto&& ordered_map2 = ordered_map1[item2.first].first;
      for (auto&& item3 : item2.second.section_map()) {
        std::cout << "item3: " << item3.first << std::endl;
        ordered_map2[item3.first] = &item3.second;
      }
    }
  }
  return true;
}

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

const ActivityConf::Activity_OrderedMap& ActivityConf::OrderedMap() const { return ordered_map_; }

const ActivityConf::Activity_Chapter_OrderedMap* ActivityConf::GetOrderedMap(uint64_t key1) const {
  auto conf = &OrderedMap();

  auto iter = conf->find(key1);
  if (iter == conf->end()) {
    return nullptr;
  }
  return &iter->second.first;
}

const ActivityConf::Section_OrderedMap* ActivityConf::GetOrderedMap(uint64_t key1, uint32_t key2) const {
  auto conf = GetOrderedMap(key1);
  if (conf == nullptr) {
    return nullptr;
  }

  auto iter = conf->find(key2);
  if (iter == conf->end()) {
    return nullptr;
  }
  return &iter->second.first;
}

}  // namespace tableau
