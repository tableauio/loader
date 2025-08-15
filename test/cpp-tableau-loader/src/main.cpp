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

bool LoadWithPatch(std::shared_ptr<const tableau::LoadOptions> options) {
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
  auto options = std::make_shared<tableau::LoadOptions>();
  options->read_func = CustomReadFile;

  // patchconf
  std::cout << "-----TestPatch patchconf" << std::endl;
  options->patch_dirs = {"../../testdata/patchconf/"};
  bool ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    std::cout << "failed to load with patchconf" << std::endl;
    return false;
  }

  // print recursive patch conf
  auto mgr = Hub::Instance().Get<protoconf::RecursivePatchConfMgr>();
  if (!mgr) {
    std::cout << "protobuf hub get RecursivePatchConf failed!" << std::endl;
    return false;
  }
  std::cout << "RecursivePatchConf: " << std::endl << mgr->Data().ShortDebugString() << std::endl;
  tableau::RecursivePatchConf result;
  ok = result.Load("../../testdata/patchresult/", tableau::Format::kJSON);
  if (!ok) {
    std::cout << "failed to load with patch result" << std::endl;
    return false;
  }
  std::cout << "Expected patch result: " << std::endl << result.Data().ShortDebugString() << std::endl;
  if (!google::protobuf::util::MessageDifferencer::Equals(mgr->Data(), result.Data())) {
    std::cout << "patch result not correct" << std::endl;
    return false;
  }

  // patchconf2
  std::cout << "-----TestPatch patchconf2" << std::endl;
  options->patch_dirs = {"../../testdata/patchconf2/"};
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    std::cout << "failed to load with patchconf2" << std::endl;
    return false;
  }

  // patchconf2 different format
  std::cout << "-----TestPatch patchconf2 different format" << std::endl;
  options->patch_dirs = {"../../testdata/patchconf2/"};
  auto mopts = std::make_shared<tableau::MessagerOptions>();
  mopts->patch_paths = {"../../testdata/patchconf2/PatchMergeConf.txt"};
  options->messager_options["PatchMergeConf"] = mopts;
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    std::cout << "failed to load with patchconf2" << std::endl;
    return false;
  }

  // multiple patch files
  std::cout << "-----TestPatch multiple patch files" << std::endl;
  mopts = std::make_shared<tableau::MessagerOptions>();
  mopts->patch_paths = {"../../testdata/patchconf/PatchMergeConf.json",
                        "../../testdata/patchconf2/PatchMergeConf.json"};
  options->messager_options["PatchMergeConf"] = mopts;
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    std::cout << "failed to load with multiple patch files" << std::endl;
    return false;
  }

  // mode only main
  std::cout << "-----TestPatch ModeOnlyMain" << std::endl;
  mopts = std::make_shared<tableau::MessagerOptions>();
  mopts->patch_paths = {"../../testdata/patchconf/PatchMergeConf.json",
                        "../../testdata/patchconf2/PatchMergeConf.json"};
  options->messager_options["PatchMergeConf"] = mopts;
  options->mode = tableau::LoadMode::kOnlyMain;
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    std::cout << "failed to load with mode only main" << std::endl;
    return false;
  }
  auto patch_mgr = Hub::Instance().Get<protoconf::PatchMergeConfMgr>();
  if (!patch_mgr) {
    std::cout << "protobuf hub get PatchMergeConf failed!" << std::endl;
    return 1;
  }
  std::cout << "PatchMergeConf: " << patch_mgr->Data().ShortDebugString() << std::endl;

  // mode only patch
  std::cout << "-----TestPatch ModeOnlyPatch" << std::endl;
  mopts = std::make_shared<tableau::MessagerOptions>();
  mopts->patch_paths = {"../../testdata/patchconf/PatchMergeConf.json",
                        "../../testdata/patchconf2/PatchMergeConf.json"};
  options->messager_options["PatchMergeConf"] = mopts;
  options->mode = tableau::LoadMode::kOnlyPatch;
  ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    std::cout << "failed to load with mode only patch" << std::endl;
    return false;
  }
  return true;
}

int main() {
  Hub::Instance().InitOnce();
  auto options = std::make_shared<tableau::LoadOptions>();
  options->ignore_unknown_fields = true;
  options->patch_dirs = {"../../testdata/patchconf/"};
  auto mopts = std::make_shared<tableau::MessagerOptions>();
  mopts->path = "../../testdata/conf/ItemConf.json";
  options->messager_options["ItemConf"] = mopts;

  bool ok = Hub::Instance().Load("../../testdata/conf/", tableau::Format::kJSON, options);
  if (!ok) {
    std::cout << "protobuf hub load failed: " << tableau::GetErrMsg() << std::endl;
    return 1;
  }
  auto msger_map = Hub::Instance().GetMessagerMap();
  for (auto&& item : *msger_map) {
    auto&& stats = item.second->GetStats();
    ATOM_DEBUG("%s: duration: %dus", item.first.c_str(), stats.duration.count());
  }

  auto item_mgr = Hub::Instance().Get<protoconf::ItemConfMgr>();
  if (!item_mgr) {
    std::cout << "protobuf hub get Item failed!" << std::endl;
    return 1;
  }
  // std::cout << "item1: " << item_mgr->Data().DebugString() << std::endl;

  std::cout << "-----Index: multi-column index test" << std::endl;
  tableau::ItemConf::Index_AwardItemKey key{1, "apple"};
  auto item = item_mgr->FindFirstAwardItem(key);
  if (!item) {
    std::cout << "ItemConf FindFirstAwardItem failed!" << std::endl;
    return 1;
  }
  std::cout << "item: " << item->ShortDebugString() << std::endl;

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
    std::cout << "ActivityConf GetOrderedMap chapter failed!" << std::endl;
    return 1;
  }

  for (auto&& it : *chapter_ordered_map) {
    std::cout << "---" << it.first << "-----section_ordered_map" << std::endl;
    for (auto&& kv : it.second.first) {
      std::cout << kv.first << std::endl;
    }

    std::cout << "---" << it.first << " -----section_map" << std::endl;
    for (auto&& kv : it.second.second->section_map()) {
      std::cout << kv.first << std::endl;
    }

    std::cout << "chapter_id: " << it.second.second->chapter_id() << std::endl;
    std::cout << "chapter_name: " << it.second.second->chapter_name() << std::endl;
    std::cout << "award_id:" << it.second.second->award_id() << std::endl;
  }

  const auto* rank_ordered_map =
      Hub::Instance().GetOrderedMap<protoconf::ActivityConfMgr, tableau::ActivityConf::int32_OrderedMap>(100001, 1, 2);
  if (!rank_ordered_map) {
    std::cout << "ActivityConf GetOrderedMap rank failed!" << std::endl;
    return 1;
  }
  std::cout << "-----rank_ordered_map" << std::endl;
  for (auto&& it : *rank_ordered_map) {
    std::cout << it.first << std::endl;
  }

  auto activity_conf = Hub::Instance().Get<tableau::ActivityConf>();
  if (!activity_conf) {
    std::cout << "protobuf hub get ActivityConf failed!" << std::endl;
    return 1;
  }

  std::cout << "-----Index accessers test" << std::endl;
  auto index_chapters = activity_conf->FindChapter(1);
  if (!index_chapters) {
    std::cout << "ActivityConf FindChapter failed!" << std::endl;
    return 1;
  }
  std::cout << "-----FindChapter" << std::endl;
  for (auto&& chapter : *index_chapters) {
    std::cout << chapter->ShortDebugString() << std::endl;
  }

  auto index_first_chapter = activity_conf->FindFirstChapter(1);
  if (!index_first_chapter) {
    std::cout << "ActivityConf FindFirstChapter failed!" << std::endl;
    return 1;
  }

  std::cout << "-----FindFirstChapter" << std::endl;
  std::cout << index_first_chapter->ShortDebugString() << std::endl;

  std::cout << "specialItemName: " << Hub::Instance().Get<CustomItemConf>()->GetSpecialItemName() << std::endl;

  if (!TestPatch()) {
    std::cerr << "TestPatch failed!" << std::endl;
    return 1;
  }
  return 0;
}