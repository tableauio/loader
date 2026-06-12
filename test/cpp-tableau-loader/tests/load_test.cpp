// Load block: hub loading, custom conf (ProcessAfterLoadAll), binary-format
// loading and patch loading. Mirrors the same scenarios in:
//   - Go:  test/go-tableau-loader/load_test.go
//   - C#:  test/csharp-tableau-loader/tests/LoadTests.cs
//   - TS:  test/ts-tableau-loader/tests/load.test.ts

#include <google/protobuf/util/message_differencer.h>
#include <gtest/gtest.h>

#include "hub/custom/item/custom_item_conf.h"
#include "protoconf/hero_conf.pc.h"
#include "protoconf/item_conf.pc.h"
#include "protoconf/patch_conf.pc.h"
#include "tests/hub_fixture.h"

namespace {

using ::google::protobuf::util::MessageDifferencer;

// ---- Load ----

TEST_F(HubFixture, Load_AllMessagers_Succeeds) {
  // SetUp() already asserts the hub loads; verify a few messagers are present.
  EXPECT_NE(Hub::Instance().Get<protoconf::ItemConfMgr>(), nullptr);
  EXPECT_NE(Hub::Instance().Get<tableau::ActivityConf>(), nullptr);
  EXPECT_NE(Hub::Instance().Get<tableau::FruitConf>(), nullptr);
}

// ---- CustomConf ----

TEST_F(HubFixture, CustomItemConf_ProcessAfterLoadAll_ResolvesSpecialItem) {
  auto custom = Hub::Instance().Get<CustomItemConf>();
  ASSERT_NE(custom, nullptr);
  EXPECT_FALSE(custom->GetSpecialItemName().empty());
}

// ---- Bin ----

TEST_F(HubFixture, HeroConf_LoadFromBin_Succeeds) {
  tableau::HeroConf hero_conf;
  bool ok = hero_conf.Load(test::TestPaths::Bin().string() + "/", tableau::Format::kBin);
  ASSERT_TRUE(ok) << "failed to load HeroConf.binpb: " << tableau::GetErrMsg();
}

TEST_F(HubFixture, HeroConf_LoadFromMissingDir_Fails) {
  tableau::HeroConf hero_conf;
  auto missing = (test::TestPaths::Testdata() / "notexist").string() + "/";
  EXPECT_FALSE(hero_conf.Load(missing, tableau::Format::kBin));
}

// ---- Patch ----
// PatchTest uses its own fixture: it only initializes the registry (InitOnce)
// and loads a fresh hub per test with patch options, mirroring the C#/Go patch
// tests that construct a standalone hub.

class PatchTest : public ::testing::Test {
 protected:
  void SetUp() override { Hub::Instance().InitOnce(); }

  std::shared_ptr<tableau::load::Options> NewOptions() const {
    auto options = std::make_shared<tableau::load::Options>();
    options->ignore_unknown_fields = true;
    return options;
  }
};

TEST_F(PatchTest, PatchConf_RecursivePatchConf_MatchesExpectedResult) {
  auto options = NewOptions();
  options->patch_dirs = {test::TestPaths::PatchConf().string()};

  bool ok = Hub::Instance().Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, options);
  ASSERT_TRUE(ok) << "load failed: " << tableau::GetErrMsg();

  auto mgr = Hub::Instance().Get<protoconf::RecursivePatchConfMgr>();
  ASSERT_NE(mgr, nullptr);

  // Load expected golden result.
  tableau::RecursivePatchConf expected;
  ok = expected.Load(test::TestPaths::PatchResult().string() + "/", tableau::Format::kJSON);
  ASSERT_TRUE(ok) << "load expected failed: " << tableau::GetErrMsg();

  EXPECT_TRUE(MessageDifferencer::Equals(mgr->Data(), expected.Data()))
      << "actual: " << mgr->Data().ShortDebugString() << "\nexpected: " << expected.Data().ShortDebugString();
}

TEST_F(PatchTest, PatchConf_PatchReplaceConf_ReplacesEntirely) {
  auto options = NewOptions();
  options->patch_dirs = {test::TestPaths::PatchConf().string()};

  ASSERT_TRUE(Hub::Instance().Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, options));
  auto mgr = Hub::Instance().Get<protoconf::PatchReplaceConfMgr>();
  ASSERT_NE(mgr, nullptr);
  EXPECT_EQ("orange", mgr->Data().name());
}

TEST_F(PatchTest, PatchConf2_LoadsSuccessfully) {
  auto options = NewOptions();
  options->patch_dirs = {test::TestPaths::PatchConf2().string()};
  ASSERT_TRUE(Hub::Instance().Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, options));
}

TEST_F(PatchTest, PatchPaths_DifferentFormat_TxtPb) {
  auto options = NewOptions();
  options->patch_dirs = {test::TestPaths::PatchConf2().string()};
  auto mopts = std::make_shared<tableau::load::MessagerOptions>();
  mopts->patch_paths = {(test::TestPaths::PatchConf2() / "PatchMergeConf.txtpb").string()};
  options->messager_options["PatchMergeConf"] = mopts;

  bool ok = Hub::Instance().Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, options);
  ASSERT_TRUE(ok) << "load failed: " << tableau::GetErrMsg();
  EXPECT_NE(Hub::Instance().Get<protoconf::PatchMergeConfMgr>(), nullptr);
}

TEST_F(PatchTest, PatchPaths_MultiplePatchFiles) {
  auto options = NewOptions();
  auto mopts = std::make_shared<tableau::load::MessagerOptions>();
  mopts->patch_paths = {(test::TestPaths::PatchConf() / "PatchMergeConf.json").string(),
                        (test::TestPaths::PatchConf2() / "PatchMergeConf.json").string()};
  options->messager_options["PatchMergeConf"] = mopts;

  ASSERT_TRUE(Hub::Instance().Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, options));
  auto mgr = Hub::Instance().Get<protoconf::PatchMergeConfMgr>();
  ASSERT_NE(mgr, nullptr);
  // patchconf2's patch contributes key 999 into both ItemMap and ReplaceItemMap.
  EXPECT_TRUE(mgr->Data().item_map().contains(999)) << "ItemMap should contain key 999 from patchconf2";
  EXPECT_TRUE(mgr->Data().replace_item_map().contains(999)) << "ReplaceItemMap should contain key 999";
}

TEST_F(PatchTest, ModeOnlyMain_IgnoresPatches) {
  auto options = NewOptions();
  auto mopts = std::make_shared<tableau::load::MessagerOptions>();
  mopts->patch_paths = {(test::TestPaths::PatchConf() / "PatchMergeConf.json").string(),
                        (test::TestPaths::PatchConf2() / "PatchMergeConf.json").string()};
  options->messager_options["PatchMergeConf"] = mopts;
  options->mode = tableau::load::LoadMode::kOnlyMain;

  ASSERT_TRUE(Hub::Instance().Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, options));
  auto mgr = Hub::Instance().Get<protoconf::PatchMergeConfMgr>();
  ASSERT_NE(mgr, nullptr);

  // Should equal a fresh OnlyMain load of the same file.
  tableau::PatchMergeConf direct;
  auto direct_opts = std::make_shared<tableau::load::MessagerOptions>();
  direct_opts->mode = tableau::load::LoadMode::kOnlyMain;
  ASSERT_TRUE(direct.Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, direct_opts));

  EXPECT_TRUE(MessageDifferencer::Equals(mgr->Data(), direct.Data()));
}

TEST_F(PatchTest, ModeOnlyPatch_AppliesPatchesFromEmpty) {
  auto options = NewOptions();
  auto mopts = std::make_shared<tableau::load::MessagerOptions>();
  mopts->patch_paths = {(test::TestPaths::PatchConf() / "PatchMergeConf.json").string(),
                        (test::TestPaths::PatchConf2() / "PatchMergeConf.json").string()};
  options->messager_options["PatchMergeConf"] = mopts;
  options->mode = tableau::load::LoadMode::kOnlyPatch;

  ASSERT_TRUE(Hub::Instance().Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, options));
  auto mgr = Hub::Instance().Get<protoconf::PatchMergeConfMgr>();
  ASSERT_NE(mgr, nullptr);
  // OnlyPatch starts from an empty message; Name must come from a patch file.
  EXPECT_FALSE(mgr->Data().name().empty());
}

}  // namespace
