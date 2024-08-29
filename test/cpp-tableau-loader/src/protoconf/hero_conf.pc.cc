// Code generated by protoc-gen-cpp-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-cpp-tableau-loader v0.6.0
// - protoc                        v3.19.3
// source: hero_conf.proto

#include "hero_conf.pc.h"

namespace tableau {
const std::string HeroConf::kProtoName = "HeroConf";

bool HeroConf::Load(const std::string& dir, Format fmt, const LoadOptions* options /* = nullptr */) {
  bool ok = LoadMessage(data_, dir, fmt, options);
  return ok ? ProcessAfterLoad() : false;
}

bool HeroConf::ProcessAfterLoad() {
  // OrderedMap init.
  for (auto&& item1 : data_.hero_map()) {
    ordered_map_[item1.first] = Hero_OrderedMapValue(Hero_Attr_OrderedMap(), &item1.second);
    auto&& ordered_map1 = ordered_map_[item1.first].first;
    for (auto&& item2 : item1.second.attr_map()) {
      ordered_map1[item2.first] = &item2.second;
    }
  }

  // Index init.
  // Index: Title
  for (auto&& item1 : data_.hero_map()) {
    for (auto&& item2 : item1.second.attr_map()) {
      index_attr_map_[item2.second.title()].push_back(&item2.second);
    }
  }

  return true;
}

const protoconf::HeroConf::Hero* HeroConf::Get(const std::string& name) const {
  auto iter = data_.hero_map().find(name);
  if (iter == data_.hero_map().end()) {
    return nullptr;
  }
  return &iter->second;
}

const protoconf::HeroConf::Hero::Attr* HeroConf::Get(const std::string& name, const std::string& title) const {
  const auto* conf = Get(name);
  if (conf == nullptr) {
    return nullptr;
  }
  auto iter = conf->attr_map().find(title);
  if (iter == conf->attr_map().end()) {
    return nullptr;
  }
  return &iter->second;
}

const HeroConf::Hero_OrderedMap* HeroConf::GetOrderedMap() const {
  return &ordered_map_; 
}

const HeroConf::Hero_Attr_OrderedMap* HeroConf::GetOrderedMap(const std::string& name) const {
  const auto* conf = GetOrderedMap();
  if (conf == nullptr) {
    return nullptr;
  }

  auto iter = conf->find(name);
  if (iter == conf->end()) {
    return nullptr;
  }
  return &iter->second.first;
}

// Index: Title
const HeroConf::Index_AttrMap& HeroConf::FindAttr() const { return index_attr_map_ ;}

const HeroConf::Index_AttrVector* HeroConf::FindAttr(const std::string& title) const {
  auto iter = index_attr_map_.find(title);
  if (iter == index_attr_map_.end()) {
    return nullptr;
  }
  return &iter->second;
}

const protoconf::HeroConf::Hero::Attr* HeroConf::FindFirstAttr(const std::string& title) const {
  auto conf = FindAttr(title);
  if (conf == nullptr || conf->size() == 0) {
    return nullptr;
  }
  return (*conf)[0];
}


}  // namespace tableau
