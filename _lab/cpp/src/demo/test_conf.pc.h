#pragma once
#include <google/protobuf/map.h>

#include <map>
#include <string>
#include <unordered_map>
#include <vector>

#include "hub.pc.h"
#include "protoconf/test_conf.pb.h"

namespace tableau {

class ActivityConf : public Messager {
 public:
  static const std::string& Name() { return kProtoName; };
  virtual bool Load(const std::string& dir, Format fmt) override;
  const protoconf::ActivityConf& Data() const { return data_; };
  const protoconf::ActivityConf::Activity* Get(uint64_t key1) const;
  const protoconf::ActivityConf::Activity::Chapter* Get(uint64_t key1, uint32_t key2) const;
  const protoconf::Section* Get(uint64_t key1, uint32_t key2, uint32_t key3) const;

 private:
  virtual bool ProcessAfterLoad() override final;

 private:
  static const std::string kProtoName;
  protoconf::ActivityConf data_;

  // OrederedMap related accessers.
 public:
  using protoconf_Section_OrderedMap = std::map<uint32_t, const protoconf::Section*>;
  const protoconf_Section_OrderedMap* GetOrderedMap(uint64_t key1, uint32_t key2) const;

  using Section_Map = ::google::protobuf::Map<uint32_t, protoconf::Section>;
  using Activity_Chapter_OrderedMapValue = std::pair<protoconf_Section_OrderedMap, const Section_Map*>;
  using Activity_Chapter_OrderedMap = std::map<uint32_t, Activity_Chapter_OrderedMapValue>;
  const Activity_Chapter_OrderedMap* GetOrderedMap(uint64_t key1) const;

  using Activity_Chapter_Map = ::google::protobuf::Map<uint32_t, protoconf::ActivityConf::Activity::Chapter>;
  using Activity_OrderedMapValue = std::pair<Activity_Chapter_OrderedMap, const Activity_Chapter_Map*>;
  using Activity_OrderedMap = std::map<uint64_t, Activity_OrderedMapValue>;
  const Activity_OrderedMap* GetOrderedMap() const;

 private:
  Activity_OrderedMap ordered_map_;

  // Index related accessers.
 public:
  using Index_ChapterVector = std::vector<const protoconf::ActivityConf::Activity::Chapter*>;
  using Index_ChapterMap = std::unordered_map<uint32_t, Index_ChapterVector>;
  const Index_ChapterMap& FindChapter() const;
  const Index_ChapterVector* FindChapter(uint32_t chapter_id) const;
  const protoconf::ActivityConf::Activity::Chapter* FindFirstChapter(uint32_t chapter_id) const;

 private:
  Index_ChapterMap index_chapter_map_;
};

}  // namespace tableau

namespace protoconf {
using ActivityConfMgr = tableau::ActivityConf;
}  // namespace protoconf
