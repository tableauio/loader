// OrderedMap block: GetOrderedMap traversal. Mirrors the same scenarios in:
//   - Go:  test/go-tableau-loader/ordered_map_test.go
//   - C#:  test/csharp-tableau-loader/tests/OrderedMapTests.cs
//   - TS:  test/ts-tableau-loader/tests/ordered_map.test.ts

#include <gtest/gtest.h>

#include <limits>

#include "protoconf/hero_conf.pc.h"
#include "protoconf/item_conf.pc.h"
#include "tests/hub_fixture.h"

namespace {

// ---- ActivityConf: nested ordered map ----
// getOrderedMap exposes the full sorted tree; each node carries its value
// (.second) and the next-level sub-map (.first). std::map iterates ascending.

TEST_F(HubFixture, ActivityConf_GetOrderedMap_Traverses) {
  auto conf = Hub::Instance().Get<tableau::ActivityConf>();
  ASSERT_NE(conf, nullptr);

  const auto* ordered_map = conf->GetOrderedMap();
  ASSERT_NE(ordered_map, nullptr);
  EXPECT_FALSE(ordered_map->empty());

  // 1st level: keys ascending; node.second is the Activity value.
  uint64_t prev = 0;
  for (const auto& kv : *ordered_map) {
    EXPECT_GE(kv.first, prev) << "ordered map keys not ascending";
    prev = kv.first;

    // 2nd level sub-map is reachable and non-empty (every activity has chapters).
    const auto& chapter_ordered_map = kv.second.first;
    EXPECT_FALSE(chapter_ordered_map.empty()) << "activity " << kv.first << " has no chapters";
    for (const auto& kv2 : chapter_ordered_map) {
      (void)kv2.first;
      (void)kv2.second.second;
    }
  }
}

// Hub::GetOrderedMap<Mgr, T>(leveled-keys): the typed, key-scoped overloads that
// return a specific sub-tree level directly (1-key -> chapter map, 3-key -> rank
// map). These complement the no-arg conf->GetOrderedMap() traversal above.

TEST_F(HubFixture, ActivityConf_GetOrderedMap_Chapter) {
  const auto* chapter_ordered_map =
      Hub::Instance().GetOrderedMap<protoconf::ActivityConfMgr,
                                    tableau::ActivityConf::OrderedMap_Activity_ChapterMap>(100001);
  ASSERT_NE(chapter_ordered_map, nullptr);
  EXPECT_FALSE(chapter_ordered_map->empty());
}

TEST_F(HubFixture, ActivityConf_GetOrderedMap_Rank) {
  const auto* rank_ordered_map =
      Hub::Instance().GetOrderedMap<protoconf::ActivityConfMgr, tableau::ActivityConf::OrderedMap_int32Map>(100001, 1,
                                                                                                            2);
  ASSERT_NE(rank_ordered_map, nullptr);
}

// ---- ItemConf: 1st-level ordered map ----

TEST_F(HubFixture, ItemConf_GetOrderedMap_Ascending) {
  auto item = Hub::Instance().Get<protoconf::ItemConfMgr>();
  ASSERT_NE(item, nullptr);

  const auto* ordered_map = item->GetOrderedMap();
  ASSERT_NE(ordered_map, nullptr);
  EXPECT_FALSE(ordered_map->empty());

  // keys ascending by id; prev starts at 0 so iteration 1 (kv.first >= 0) holds.
  uint32_t prev = 0;
  for (const auto& kv : *ordered_map) {
    EXPECT_GE(kv.first, prev) << "ItemConf ordered map keys not ascending";
    prev = kv.first;
  }
}

// ---- HeroConf: ordered map accessible ----

TEST_F(HubFixture, HeroConf_GetOrderedMap_Accessible) {
  auto hero_conf = Hub::Instance().Get<tableau::HeroConf>();
  ASSERT_NE(hero_conf, nullptr);

  const auto* hero_ordered_map = hero_conf->GetOrderedMap();
  ASSERT_NE(hero_ordered_map, nullptr);
}

}  // namespace
