#include "test_conf.pc.h"

namespace tableau {
const std::string ActivityConf::kProtoName = "ActivityConf";
bool ActivityConf::Load(const std::string& dir, Format fmt) {
  bool ok = LoadMessage(dir, data_, fmt);
  return ok ? ProcessAfterLoad() : false;
}

bool ActivityConf::ProcessAfterLoad() {
  // init ordered map.
  for (auto&& item1 : data_.activity_map()) {
    ordered_map_[item1.first] = Activity_OrderedMapValue(Activity_Chapter_OrderedMap(), &item1.second.chapter_map());
    auto&& ordered_map1 = ordered_map_[item1.first].first;
    for (auto&& item2 : item1.second.chapter_map()) {
      ordered_map1[item2.first] =
          Activity_Chapter_OrderedMapValue(protoconf_Section_OrderedMap(), &item2.second.section_map());
      auto&& ordered_map2 = ordered_map1[item2.first].first;
      for (auto&& item3 : item2.second.section_map()) {
        ordered_map2[item3.first] = &item3.second;
      }
    }
  }

  // init index.
  for (auto&& item1 : data_.activity_map()) {
    for (auto&& item2 : item1.second.chapter_map()) {
      index_chapter_map_[item2.second.chapter_id()].push_back(&item2.second);
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

const ActivityConf::Activity_OrderedMap* ActivityConf::GetOrderedMap() const { return &ordered_map_; }

const ActivityConf::Activity_Chapter_OrderedMap* ActivityConf::GetOrderedMap(uint64_t key1) const {
  auto conf = GetOrderedMap();
  if (conf == nullptr) {
    return nullptr;
  }

  auto iter = conf->find(key1);
  if (iter == conf->end()) {
    return nullptr;
  }
  return &iter->second.first;
}

const ActivityConf::protoconf_Section_OrderedMap* ActivityConf::GetOrderedMap(uint64_t key1, uint32_t key2) const {
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

const ActivityConf::Index_ChapterMap& ActivityConf::FindChapter() const { return index_chapter_map_; }

const ActivityConf::Index_ChapterVector* ActivityConf::FindChapter(uint32_t chapter_id) const {
  auto iter = index_chapter_map_.find(chapter_id);
  if (iter == index_chapter_map_.end()) {
    return nullptr;
  }
  return &iter->second;
}

const protoconf::ActivityConf::Activity::Chapter* ActivityConf::FindFirstChapter(uint32_t chapter_id) const {
  auto conf = FindChapter(chapter_id);
  if (conf == nullptr || conf->size() == 0) {
    return nullptr;
  }
  return (*conf)[0];
}

}  // namespace tableau
