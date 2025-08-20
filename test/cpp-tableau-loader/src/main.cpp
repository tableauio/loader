#include <google/protobuf/util/message_differencer.h>

#include <fstream>
#include <iostream>
#include <string>
#include <unordered_map>

#include "hub/custom/item/custom_item_conf.h"
#include "hub/hub.h"
#include "protoconf/hub.pc.h"
#include "protoconf/item_conf.pc.h"
#include "protoconf/logger.pc.h"
#include "protoconf/patch_conf.pc.h"
#include "protoconf/test_conf.pc.h"

bool LoadWithPatch(std::shared_ptr<const tableau::load::Options> options) {
  return Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
}

bool CustomReadFile(const std::filesystem::path& filename, std::string& content) {
  std::ifstream file(filename);
  if (!file.is_open()) {
    return false;
  }
  content.assign(std::istreambuf_iterator<char>(file), {});
  ATOM_DEBUG("custom read %s success", filename.c_str());
  return true;
}

bool TestPatch() {
  auto options = std::make_shared<tableau::load::Options>();
  options->read_func = CustomReadFile;

  // patchconf
  ATOM_DEBUG("-----TestPatch patchconf");
  options->patch_dirs = {"../../testdata/patchconf/"};
  bool ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    ATOM_ERROR("failed to load with patchconf");
    return false;
  }

  // print recursive patch conf
  auto mgr = Hub::Instance().Get<protoconf::RecursivePatchConfMgr>();
  if (!mgr) {
    ATOM_ERROR("protobuf hub get RecursivePatchConf failed!");
    return false;
  }
  ATOM_DEBUG("RecursivePatchConf: %s", mgr->Data().ShortDebugString().c_str());
  tableau::RecursivePatchConf result;
  ok = result.Load("../../testdata/patchresult/", tableau::Format::kJSON);
  if (!ok) {
    ATOM_ERROR("failed to load with patch result");
    return false;
  }
  ATOM_DEBUG("Expected patch result: %s", result.Data().ShortDebugString().c_str());
  if (!google::protobuf::util::MessageDifferencer::Equals(mgr->Data(), result.Data())) {
    ATOM_ERROR("patch result not correct");
    return false;
  }

  // patchconf2
  ATOM_DEBUG("-----TestPatch patchconf2");
  options->patch_dirs = {"../../testdata/patchconf2/"};
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    ATOM_ERROR("failed to load with patchconf2");
    return false;
  }

  // patchconf2 different format
  ATOM_DEBUG("-----TestPatch patchconf2 different format");
  options->patch_dirs = {"../../testdata/patchconf2/"};
  auto mopts = std::make_shared<tableau::load::MessagerOptions>();
  mopts->patch_paths = {"../../testdata/patchconf2/PatchMergeConf.txt"};
  options->messager_options["PatchMergeConf"] = mopts;
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    ATOM_ERROR("failed to load with patchconf2");
    return false;
  }

  // multiple patch files
  ATOM_DEBUG("-----TestPatch multiple patch files");
  mopts = std::make_shared<tableau::load::MessagerOptions>();
  mopts->patch_paths = {"../../testdata/patchconf/PatchMergeConf.json",
                        "../../testdata/patchconf2/PatchMergeConf.json"};
  options->messager_options["PatchMergeConf"] = mopts;
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    ATOM_ERROR("failed to load with multiple patch files");
    return false;
  }

  // mode only main
  ATOM_DEBUG("-----TestPatch ModeOnlyMain");
  mopts = std::make_shared<tableau::load::MessagerOptions>();
  mopts->patch_paths = {"../../testdata/patchconf/PatchMergeConf.json",
                        "../../testdata/patchconf2/PatchMergeConf.json"};
  options->messager_options["PatchMergeConf"] = mopts;
  options->mode = tableau::load::LoadMode::kOnlyMain;
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    ATOM_ERROR("failed to load with mode only main");
    return false;
  }
  auto patch_mgr = Hub::Instance().Get<protoconf::PatchMergeConfMgr>();
  if (!patch_mgr) {
    ATOM_ERROR("protobuf hub get PatchMergeConf failed!");
    return false;
  }
  ATOM_DEBUG("PatchMergeConf: %s", patch_mgr->Data().ShortDebugString().c_str());

  // mode only patch
  ATOM_DEBUG("-----TestPatch ModeOnlyPatch");
  mopts = std::make_shared<tableau::load::MessagerOptions>();
  mopts->patch_paths = {"../../testdata/patchconf/PatchMergeConf.json",
                        "../../testdata/patchconf2/PatchMergeConf.json"};
  options->messager_options["PatchMergeConf"] = mopts;
  options->mode = tableau::load::LoadMode::kOnlyPatch;
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    ATOM_ERROR("failed to load with mode only patch");
    return false;
  }
  return true;
}

int main() {
  Hub::Instance().InitOnce();
  auto options = std::make_shared<tableau::load::Options>();
  options->ignore_unknown_fields = true;
  options->patch_dirs = {"../../testdata/patchconf/"};
  auto mopts = std::make_shared<tableau::load::MessagerOptions>();
  mopts->path = "../../testdata/conf/ItemConf.json";
  options->messager_options["ItemConf"] = mopts;

  bool ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    ATOM_ERROR("protobuf hub load failed: %s", tableau::GetErrMsg().c_str());
    return 1;
  }
  auto msger_map = Hub::Instance().GetMessagerMap();
  for (auto&& item : *msger_map) {
    auto&& stats = item.second->GetStats();
    ATOM_DEBUG("%s: duration: %dus", item.first.c_str(), stats.duration.count());
  }

  auto item_mgr = Hub::Instance().Get<protoconf::ItemConfMgr>();
  if (!item_mgr) {
    ATOM_ERROR("protobuf hub get Item failed!");
    return 1;
  }
  // std::cout << "item1: " << item_mgr->Data().DebugString() << std::endl;

  ATOM_DEBUG("-----Index: multi-column index test");
  tableau::ItemConf::Index_AwardItemKey key{1, "apple"};
  auto item = item_mgr->FindFirstAwardItem(key);
  if (!item) {
    ATOM_ERROR("ItemConf FindFirstAwardItem failed!");
    return 1;
  }
  ATOM_DEBUG("item: %s", item->ShortDebugString().c_str());

  //   auto activity_conf = Hub::Instance().Get<tableau::ActivityConf>();
  //   if (!activity_conf) {
  //     std::cout << "protobuf hub get ActivityConf failed!" << std::endl;
  //     return 1;
  //   }

  //   const auto* section_conf = activity_conf->Get(100001, 1, 2);
  //   if (!section_conf) {
  //     std::cout << "ActivityConf get section failed!" << std::endl;
  //     return 1;
  //   }

  //   const auto* section_conf = Hub::Instance().Get<protoconf::ActivityConfMgr, protoconf::Section>(100001, 1, 2);
  //   if (!section_conf) {
  //     std::cout << "ActivityConf get section failed!" << std::endl;
  //     return 1;
  //   }

  //   std::cout << "-----section_conf" << std::endl;
  //   std::cout << section_conf->DebugString() << std::endl;

  const auto* chapter_ordered_map =
      Hub::Instance().GetOrderedMap<protoconf::ActivityConfMgr, tableau::ActivityConf::Activity_Chapter_OrderedMap>(
          100001);
  if (!chapter_ordered_map) {
    ATOM_ERROR("ActivityConf GetOrderedMap chapter failed!");
    return 1;
  }

  for (auto&& it : *chapter_ordered_map) {
    ATOM_DEBUG("---%d-----section_ordered_map", it.first);
    for (auto&& kv : it.second.first) {
      ATOM_DEBUG("%d", kv.first);
    }

    ATOM_DEBUG("---%d-----section_map", it.first);
    for (auto&& kv : it.second.second->section_map()) {
      ATOM_DEBUG("%d", kv.first);
    }

    ATOM_DEBUG("chapter_id: %d", it.second.second->chapter_id());
    ATOM_DEBUG("chapter_name: %s", it.second.second->chapter_name().c_str());
    ATOM_DEBUG("award_id: %d", it.second.second->award_id());
  }

  const auto* rank_ordered_map =
      Hub::Instance().GetOrderedMap<protoconf::ActivityConfMgr, tableau::ActivityConf::int32_OrderedMap>(100001, 1, 2);
  if (!rank_ordered_map) {
    ATOM_ERROR("ActivityConf GetOrderedMap rank failed!");
    return 1;
  }
  ATOM_DEBUG("-----rank_ordered_map");
  for (auto&& it : *rank_ordered_map) {
    ATOM_DEBUG("%d", it.first);
  }

  auto activity_conf = Hub::Instance().Get<tableau::ActivityConf>();
  if (!activity_conf) {
    ATOM_ERROR("protobuf hub get ActivityConf failed!");
    return 1;
  }

  ATOM_DEBUG("-----Index accessers test");
  auto index_chapters = activity_conf->FindChapter(1);
  if (!index_chapters) {
    ATOM_ERROR("ActivityConf FindChapter failed!");
    return 1;
  }
  ATOM_DEBUG("-----FindChapter");
  for (auto&& chapter : *index_chapters) {
    ATOM_DEBUG("%s", chapter->ShortDebugString().c_str());
  }

  auto index_first_chapter = activity_conf->FindFirstChapter(1);
  if (!index_first_chapter) {
    ATOM_ERROR("ActivityConf FindFirstChapter failed!");
    return 1;
  }

  ATOM_DEBUG("-----FindFirstChapter");
  ATOM_DEBUG("%s", index_first_chapter->ShortDebugString().c_str());

  ATOM_DEBUG("specialItemName: %s", Hub::Instance().Get<CustomItemConf>()->GetSpecialItemName().c_str());

  if (!TestPatch()) {
    ATOM_ERROR("TestPatch failed!");
    return 1;
  }
  return 0;
}