// Unit tests for the C++ tableau loader, replacing the legacy print-based
// verification that previously lived in src/main.cpp.

#include <gtest/gtest.h>

#include "hub/custom/item/custom_item_conf.h"
#include "hub/hub.h"
#include "protoconf/hub.pc.h"
#include "protoconf/item_conf.pc.h"
#include "protoconf/test_conf.pc.h"
#include "tests/test_paths.h"

namespace {

// HubFixture loads the hub once per test, so test order doesn't matter
// (Hub::Instance() is a process-wide singleton shared with PatchTest).
class HubFixture : public ::testing::Test {
 protected:
  void SetUp() override {
    Hub::Instance().InitOnce();
    auto options = std::make_shared<tableau::load::Options>();
    options->ignore_unknown_fields = true;
    auto mopts = std::make_shared<tableau::load::MessagerOptions>();
    mopts->path = (test::TestPaths::Conf() / "ItemConf.json").string();
    options->messager_options["ItemConf"] = mopts;

    bool ok = Hub::Instance().Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, options);
    ASSERT_TRUE(ok) << "hub load failed: " << tableau::GetErrMsg();
  }
};

// ---- ItemConf ----

TEST_F(HubFixture, ItemConf_FindFirstAwardItem_Found) {
  auto item_mgr = Hub::Instance().Get<protoconf::ItemConfMgr>();
  ASSERT_NE(item_mgr, nullptr);
  auto item = item_mgr->FindFirstAwardItem(1, "apple");
  ASSERT_NE(item, nullptr) << "ItemConf FindFirstAwardItem(1, apple) failed";
}

TEST_F(HubFixture, ItemConf_FindItemInfoMap_NonEmpty) {
  auto item_mgr = Hub::Instance().Get<protoconf::ItemConfMgr>();
  ASSERT_NE(item_mgr, nullptr);
  const auto& info_map = item_mgr->FindItemInfoMap();
  EXPECT_FALSE(info_map.empty());
}

// ---- ActivityConf ----

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

TEST_F(HubFixture, ActivityConf_FindChapter) {
  auto activity_conf = Hub::Instance().Get<tableau::ActivityConf>();
  ASSERT_NE(activity_conf, nullptr);

  auto index_chapters = activity_conf->FindChapter(1);
  ASSERT_NE(index_chapters, nullptr);
  EXPECT_FALSE(index_chapters->empty());

  auto first_chapter = activity_conf->FindFirstChapter(1);
  ASSERT_NE(first_chapter, nullptr);
}

// ---- CustomItemConf ----

TEST_F(HubFixture, CustomItemConf_SpecialItemNameResolved) {
  auto custom = Hub::Instance().Get<CustomItemConf>();
  ASSERT_NE(custom, nullptr);
  EXPECT_FALSE(custom->GetSpecialItemName().empty());
}

}  // namespace
