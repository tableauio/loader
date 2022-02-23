#pragma once
#include <string>
#include <map>

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

  using Protoconf_Section_Map = std::map<uint32_t, const protoconf::Section*>;

  using Protoconf_Activity_Chapter_Map_ValueType = std::pair<const protoconf::ActivityConf::Activity::Chapter*, Protoconf_Section_Map>;
  using Protoconf_Activity_Chapter_Map = std::map<uint32_t, Protoconf_Activity_Chapter_Map_ValueType>;

  using Protoconf_Activity_Map_ValueType = std::pair<const protoconf::ActivityConf::Activity*, Protoconf_Activity_Chapter_Map>;
  using Protoconf_Activity_Map = std::map<uint64_t, Protoconf_Activity_Map_ValueType>;

  const Protoconf_Activity_Map& OrderedMap() const;
  const Protoconf_Activity_Chapter_Map* GetOrderedMap(uint64_t key1) const;
  const Protoconf_Section_Map* GetOrderedMap(uint64_t key1, uint32_t key2) const;

 private:
  static const std::string kProtoName;
  protoconf::ActivityConf data_;
  Protoconf_Activity_Map ordered_map_;
};

}  // namespace tableau

namespace protoconf {
using ActivityConfMgr = tableau::ActivityConf;
}  // namespace protoconf
