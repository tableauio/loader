// Code generated by protoc-gen-cpp-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-cpp-tableau-loader v0.7.0
// - protoc                        v3.19.3
// source: item_conf.proto

#pragma once
#include <string>

#include "hub.pc.h"
#include "item_conf.pb.h"

namespace tableau {
class ItemConf : public Messager {
 public:
  static const std::string& Name() { return kProtoName; }
  virtual bool Load(const std::string& dir, Format fmt, const LoadOptions* options = nullptr) override;
  const protoconf::ItemConf& Data() const { return data_; }
  const google::protobuf::Message* Message() const override { return &data_; }

 private:
  virtual bool ProcessAfterLoad() override final;

 public:
  const protoconf::ItemConf::Item* Get(uint32_t id) const;

 private:
  static const std::string kProtoName;
  protoconf::ItemConf data_;

  // OrderedMap accessers.
 public:
  using Item_OrderedMap = std::map<uint32_t, const protoconf::ItemConf::Item*>;
  const Item_OrderedMap* GetOrderedMap() const;

 private:
  Item_OrderedMap ordered_map_;

  // Index accessers.
  // Index: Type
 public:
  using Index_ItemVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_ItemMap = std::unordered_map<protoconf::FruitType, Index_ItemVector>;
  const Index_ItemMap& FindItem() const;
  const Index_ItemVector* FindItem(protoconf::FruitType type) const;
  const protoconf::ItemConf::Item* FindFirstItem(protoconf::FruitType type) const;

 private:
  Index_ItemMap index_item_map_;

  // Index: Param@ItemInfo
 public:
  using Index_ItemInfoVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_ItemInfoMap = std::unordered_map<int32_t, Index_ItemInfoVector>;
  const Index_ItemInfoMap& FindItemInfo() const;
  const Index_ItemInfoVector* FindItemInfo(int32_t param) const;
  const protoconf::ItemConf::Item* FindFirstItemInfo(int32_t param) const;

 private:
  Index_ItemInfoMap index_item_info_map_;

  // Index: Default@ItemDefaultInfo
 public:
  using Index_ItemDefaultInfoVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_ItemDefaultInfoMap = std::unordered_map<std::string, Index_ItemDefaultInfoVector>;
  const Index_ItemDefaultInfoMap& FindItemDefaultInfo() const;
  const Index_ItemDefaultInfoVector* FindItemDefaultInfo(const std::string& default_) const;
  const protoconf::ItemConf::Item* FindFirstItemDefaultInfo(const std::string& default_) const;

 private:
  Index_ItemDefaultInfoMap index_item_default_info_map_;

  // Index: ExtType@ItemExtInfo
 public:
  using Index_ItemExtInfoVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_ItemExtInfoMap = std::unordered_map<protoconf::FruitType, Index_ItemExtInfoVector>;
  const Index_ItemExtInfoMap& FindItemExtInfo() const;
  const Index_ItemExtInfoVector* FindItemExtInfo(protoconf::FruitType ext_type) const;
  const protoconf::ItemConf::Item* FindFirstItemExtInfo(protoconf::FruitType ext_type) const;

 private:
  Index_ItemExtInfoMap index_item_ext_info_map_;

  // Index: (ID,Name)@AwardItem
 public:
  struct Index_AwardItemKey {
    uint32_t id;
    std::string name;
    bool operator==(const Index_AwardItemKey& other) const {
      return id == other.id && name == other.name;
    }
  };
  struct Index_AwardItemKeyHasher {
    std::size_t operator()(const Index_AwardItemKey& key) const {
      return util::SugaredHashCombine(key.id, key.name);
    }
  };
  using Index_AwardItemVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_AwardItemMap = std::unordered_map<Index_AwardItemKey, Index_AwardItemVector, Index_AwardItemKeyHasher>;
  const Index_AwardItemMap& FindAwardItem() const;
  const Index_AwardItemVector* FindAwardItem(const Index_AwardItemKey& key) const;
  const protoconf::ItemConf::Item* FindFirstAwardItem(const Index_AwardItemKey& key) const;

 private:
  Index_AwardItemMap index_award_item_map_;

  // Index: (ID,Type,Param,ExtType)@SpecialItem
 public:
  struct Index_SpecialItemKey {
    uint32_t id;
    protoconf::FruitType type;
    int32_t param;
    protoconf::FruitType ext_type;
    bool operator==(const Index_SpecialItemKey& other) const {
      return id == other.id && type == other.type && param == other.param && ext_type == other.ext_type;
    }
  };
  struct Index_SpecialItemKeyHasher {
    std::size_t operator()(const Index_SpecialItemKey& key) const {
      return util::SugaredHashCombine(key.id, key.type, key.param, key.ext_type);
    }
  };
  using Index_SpecialItemVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_SpecialItemMap = std::unordered_map<Index_SpecialItemKey, Index_SpecialItemVector, Index_SpecialItemKeyHasher>;
  const Index_SpecialItemMap& FindSpecialItem() const;
  const Index_SpecialItemVector* FindSpecialItem(const Index_SpecialItemKey& key) const;
  const protoconf::ItemConf::Item* FindFirstSpecialItem(const Index_SpecialItemKey& key) const;

 private:
  Index_SpecialItemMap index_special_item_map_;

  // Index: PathDir@ItemPathDir
 public:
  using Index_ItemPathDirVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_ItemPathDirMap = std::unordered_map<std::string, Index_ItemPathDirVector>;
  const Index_ItemPathDirMap& FindItemPathDir() const;
  const Index_ItemPathDirVector* FindItemPathDir(const std::string& dir) const;
  const protoconf::ItemConf::Item* FindFirstItemPathDir(const std::string& dir) const;

 private:
  Index_ItemPathDirMap index_item_path_dir_map_;

  // Index: PathName@ItemPathName
 public:
  using Index_ItemPathNameVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_ItemPathNameMap = std::unordered_map<std::string, Index_ItemPathNameVector>;
  const Index_ItemPathNameMap& FindItemPathName() const;
  const Index_ItemPathNameVector* FindItemPathName(const std::string& name) const;
  const protoconf::ItemConf::Item* FindFirstItemPathName(const std::string& name) const;

 private:
  Index_ItemPathNameMap index_item_path_name_map_;

  // Index: PathFriendID@ItemPathFriendID
 public:
  using Index_ItemPathFriendIDVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_ItemPathFriendIDMap = std::unordered_map<uint32_t, Index_ItemPathFriendIDVector>;
  const Index_ItemPathFriendIDMap& FindItemPathFriendID() const;
  const Index_ItemPathFriendIDVector* FindItemPathFriendID(uint32_t id) const;
  const protoconf::ItemConf::Item* FindFirstItemPathFriendID(uint32_t id) const;

 private:
  Index_ItemPathFriendIDMap index_item_path_friend_id_map_;

  // Index: UseEffectType@UseEffectType
 public:
  using Index_UseEffectTypeVector = std::vector<const protoconf::ItemConf::Item*>;
  using Index_UseEffectTypeMap = std::unordered_map<protoconf::UseEffect::Type, Index_UseEffectTypeVector>;
  const Index_UseEffectTypeMap& FindUseEffectType() const;
  const Index_UseEffectTypeVector* FindUseEffectType(protoconf::UseEffect::Type type) const;
  const protoconf::ItemConf::Item* FindFirstUseEffectType(protoconf::UseEffect::Type type) const;

 private:
  Index_UseEffectTypeMap index_use_effect_type_map_;

};

}  // namespace tableau

namespace protoconf {
// Here are some type aliases for easy use.
using ItemConfMgr = tableau::ItemConf;
}  // namespace protoconf
