// Index block: index finders (leveled, ordered and multi-column). Mirrors:
//   - Go:  test/go-tableau-loader/index_test.go
//   - C#:  test/csharp-tableau-loader/tests/IndexTests.cs
//   - TS:  test/ts-tableau-loader/tests/index.test.ts (cases 13-16)

#include <gtest/gtest.h>

#include <limits>

#include "protoconf/index_conf.pc.h"
#include "protoconf/item_conf.pc.h"
#include "tests/hub_fixture.h"

namespace {

// ---- FruitConf: leveled index (Price<ID>) finders ----

TEST_F(HubFixture, FruitConf_IndexFinders) {
  auto fruit = Hub::Instance().Get<tableau::FruitConf>();
  ASSERT_NE(fruit, nullptr);

  // global index: price -> items
  auto items = fruit->FindItem(10);
  ASSERT_NE(items, nullptr);
  ASSERT_EQ(items->size(), 1u);
  EXPECT_EQ((*items)[0]->id(), 1001);
  ASSERT_NE(fruit->FindFirstItem(20), nullptr);
  EXPECT_EQ(fruit->FindFirstItem(20)->id(), 1002);
  EXPECT_EQ(fruit->FindItem(999), nullptr);
  // 6 items, all unique prices -> 6 entries
  EXPECT_EQ(fruit->FindItemMap().size(), 6u);

  // 1st-level scoped index: (fruitType) -> price -> items
  auto orange = fruit->FindItem(kFruitTypeOrange, 15);
  ASSERT_NE(orange, nullptr);
  ASSERT_EQ(orange->size(), 1u);
  EXPECT_EQ((*orange)[0]->id(), 2001);
  ASSERT_NE(fruit->FindFirstItem(kFruitTypeBanana, 8), nullptr);
  EXPECT_EQ(fruit->FindFirstItem(kFruitTypeBanana, 8)->id(), 3001);
  ASSERT_NE(fruit->FindItemMap(kFruitTypeApple), nullptr);
  EXPECT_EQ(fruit->FindItemMap(kFruitTypeApple)->size(), 2u);
  EXPECT_EQ(fruit->FindItemMap(999), nullptr);
  EXPECT_EQ(fruit->FindItem(999, 15), nullptr);
}

// ---- FruitConf: ordered index (Price<ID>@OrderedFruit) finders ----

TEST_F(HubFixture, FruitConf_OrderedIndexFinders) {
  auto fruit = Hub::Instance().Get<tableau::FruitConf>();
  ASSERT_NE(fruit, nullptr);

  auto items = fruit->FindOrderedFruit(10);
  ASSERT_NE(items, nullptr);
  ASSERT_EQ(items->size(), 1u);
  EXPECT_EQ((*items)[0]->id(), 1001);
  ASSERT_NE(fruit->FindFirstOrderedFruit(25), nullptr);
  EXPECT_EQ(fruit->FindFirstOrderedFruit(25)->id(), 2002);
  EXPECT_EQ(fruit->FindOrderedFruit(999), nullptr);

  // ordered map is keyed in ascending price order (std::map iterates ascending).
  const auto& m = fruit->FindOrderedFruitMap();
  EXPECT_EQ(m.size(), 6u);
  int prev = std::numeric_limits<int>::min();
  for (const auto& kv : m) {
    EXPECT_GE(kv.first, prev) << "ordered map keys not ascending";
    prev = kv.first;
  }

  // 1st-level scoped ordered index
  auto orange = fruit->FindOrderedFruit(kFruitTypeOrange, 25);
  ASSERT_NE(orange, nullptr);
  ASSERT_EQ(orange->size(), 1u);
  EXPECT_EQ((*orange)[0]->id(), 2002);
  ASSERT_NE(fruit->FindFirstOrderedFruit(kFruitTypeBanana, 12), nullptr);
  EXPECT_EQ(fruit->FindFirstOrderedFruit(kFruitTypeBanana, 12)->id(), 3002);
  EXPECT_EQ(fruit->FindOrderedFruit(999, 25), nullptr);
}

// ---- ItemConf: single-column index ----

TEST_F(HubFixture, ItemConf_FindItemInfoMap_NonEmpty) {
  auto item = Hub::Instance().Get<protoconf::ItemConfMgr>();
  ASSERT_NE(item, nullptr);
  const auto& info_map = item->FindItemInfoMap();
  EXPECT_FALSE(info_map.empty());
}

// ---- ItemConf: multi-column (composite-key) indexes ----

TEST_F(HubFixture, ItemConf_ParamExtType_OrderedMultiColumnIndex) {
  auto item = Hub::Instance().Get<protoconf::ItemConfMgr>();
  ASSERT_NE(item, nullptr);

  // apple has param_list [1,2,3] x extTypeList [APPLE, ORANGE] -> 6 buckets,
  // and it is the only item carrying both columns, so the map has 6 entries.
  const auto& m = item->FindParamExtTypeMap();
  EXPECT_EQ(m.size(), 6u);

  // std::map iterates in ascending key order; each key round-trips through the finder.
  int prev_param = std::numeric_limits<int>::min();
  for (const auto& kv : m) {
    EXPECT_GE(kv.first.param, prev_param) << "ParamExtType keys not ascending by param";
    prev_param = kv.first.param;
    auto got = item->FindParamExtType(kv.first.param, kv.first.ext_type);
    ASSERT_NE(got, nullptr);
    EXPECT_EQ(got->size(), kv.second.size());
  }

  // a concrete known bucket: (param=1, extType=APPLE) -> apple
  auto apple = item->FindParamExtType(1, protoconf::FRUIT_TYPE_APPLE);
  ASSERT_NE(apple, nullptr);
  ASSERT_EQ(apple->size(), 1u);
  EXPECT_EQ((*apple)[0]->name(), "apple");
  EXPECT_EQ(item->FindParamExtType(999, protoconf::FRUIT_TYPE_APPLE), nullptr);
}

TEST_F(HubFixture, ItemConf_AwardItem_MultiColumnIndex) {
  auto item = Hub::Instance().Get<protoconf::ItemConfMgr>();
  ASSERT_NE(item, nullptr);

  const auto& m = item->FindAwardItemMap();
  for (const auto& kv : m) {
    auto got = item->FindAwardItem(kv.first.id, kv.first.name);
    ASSERT_NE(got, nullptr);
    EXPECT_EQ(got->size(), kv.second.size());
  }

  // apple is keyed by (id=1, name="apple")
  auto apple = item->FindAwardItem(1, "apple");
  ASSERT_NE(apple, nullptr);
  ASSERT_EQ(apple->size(), 1u);
  EXPECT_EQ((*apple)[0]->id(), 1u);
  ASSERT_NE(item->FindFirstAwardItem(1, "apple"), nullptr);
}

// ---- ActivityConf: leveled index finder (C++-specific extra) ----

TEST_F(HubFixture, ActivityConf_FindChapter) {
  auto activity_conf = Hub::Instance().Get<tableau::ActivityConf>();
  ASSERT_NE(activity_conf, nullptr);

  auto index_chapters = activity_conf->FindChapter(1);
  ASSERT_NE(index_chapters, nullptr);
  EXPECT_FALSE(index_chapters->empty());

  auto first_chapter = activity_conf->FindFirstChapter(1);
  ASSERT_NE(first_chapter, nullptr);
}

}  // namespace
