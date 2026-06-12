// Get block: point-lookup getters (map getters). Mirrors the same scenarios in:
//   - Go:  test/go-tableau-loader/get_test.go
//   - C#:  test/csharp-tableau-loader/tests/GetTests.cs
//   - TS:  test/ts-tableau-loader/tests/get.test.ts

#include <gtest/gtest.h>

#include "protoconf/hero_conf.pc.h"
#include "protoconf/index_conf.pc.h"
#include "protoconf/item_conf.pc.h"
#include "tests/hub_fixture.h"

namespace {

// ---- ItemConf: 1st-level int-keyed map getter ----

TEST_F(HubFixture, ItemConf_Get) {
  auto item = Hub::Instance().Get<protoconf::ItemConfMgr>();
  ASSERT_NE(item, nullptr);

  // found
  auto apple = item->Get(1);
  ASSERT_NE(apple, nullptr);
  EXPECT_EQ(apple->name(), "apple");
  auto coin = item->Get(0);
  ASSERT_NE(coin, nullptr);
  EXPECT_EQ(coin->name(), "coin1");

  // not found
  EXPECT_EQ(item->Get(12345), nullptr);
}

// ---- ActivityConf: nested map getters with a 64-bit outer key ----

TEST_F(HubFixture, ActivityConf_Get) {
  auto conf = Hub::Instance().Get<tableau::ActivityConf>();
  ASSERT_NE(conf, nullptr);

  auto activity = conf->Get(100001);
  ASSERT_NE(activity, nullptr);
  EXPECT_EQ(activity->activity_name(), "活动1");

  auto chapter = conf->Get(100001, 1);
  ASSERT_NE(chapter, nullptr);
  EXPECT_EQ(chapter->chapter_name(), "签到活动章1");

  auto section = conf->Get(100001, 1, 2);
  ASSERT_NE(section, nullptr);
  EXPECT_EQ(section->section_id(), 2u);

  // sectionRankMap["2001"] === 2
  auto rank = conf->Get(100001, 1, 2, 2001);
  ASSERT_NE(rank, nullptr);
  EXPECT_EQ(*rank, 2);
}

TEST_F(HubFixture, ActivityConf_Get3_NotFound) {
  auto conf = Hub::Instance().Get<tableau::ActivityConf>();
  ASSERT_NE(conf, nullptr);
  EXPECT_EQ(conf->Get(100001, 1, 999), nullptr);
}

// ---- ThemeConf: string-keyed map getter ----

TEST_F(HubFixture, ThemeConf_Get) {
  auto theme = Hub::Instance().Get<tableau::ThemeConf>();
  ASSERT_NE(theme, nullptr);
  EXPECT_GT(theme->Data().theme_map_size(), 0);
}

// ---- HeroBaseConf ----

TEST_F(HubFixture, HeroBaseConf_Get) {
  auto conf = Hub::Instance().Get<tableau::HeroBaseConf>();
  ASSERT_NE(conf, nullptr);
  EXPECT_GT(conf->Data().hero_map_size(), 0);
}

// ---- FruitConf: leveled point getters ----

TEST_F(HubFixture, FruitConf_Get) {
  auto fruit = Hub::Instance().Get<tableau::FruitConf>();
  ASSERT_NE(fruit, nullptr);

  // Get1: found / not found
  ASSERT_NE(fruit->Get(kFruitTypeApple), nullptr);
  EXPECT_EQ(fruit->Get(999), nullptr);

  // Get2: APPLE -> item 1001, price 10
  auto apple_item = fruit->Get(kFruitTypeApple, 1001);
  ASSERT_NE(apple_item, nullptr);
  EXPECT_EQ(apple_item->id(), 1001);
  EXPECT_EQ(apple_item->price(), 10);

  // not found: wrong item id / wrong fruitType
  EXPECT_EQ(fruit->Get(kFruitTypeApple, 9999), nullptr);
  EXPECT_EQ(fruit->Get(999, 1001), nullptr);
}

}  // namespace
