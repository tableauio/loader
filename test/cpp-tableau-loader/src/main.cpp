#include <fstream>
#include <iostream>
#include <string>

#include "protoconf/hub.pc.h"
#include "protoconf/item_conf.pc.h"
#include "protoconf/logger.pc.h"
#include "protoconf/registry.pc.h"
#include "protoconf/test_conf.pc.h"

void WriteFile(const std::string& filename, const std::string& input) {
  std::ofstream out(filename);
  out << input;
  out.close();
}

template <class T, const bool threaded = false>
class Singleton {
 private:
  Singleton(const T&) = delete;
  Singleton(T&&) = delete;
  void operator=(const T&) = delete;
  static inline T* GetInstancePtr() {
    T* ptr = nullptr;
    if (threaded) {
      static thread_local T* new_ptr = new T();
      ptr = new_ptr;
    } else {
      static T* new_ptr = new T();
      ptr = new_ptr;
    }
    assert(ptr != nullptr);
    return ptr;
  }

 public:
  static inline T& Instance() { return *GetInstancePtr(); }

 protected:
  Singleton() = default;
};

class MyHub : public Singleton<tableau::Hub, true> {};

// syntactic sugar
template <typename T>
const std::shared_ptr<T> Get() {
  return MyHub::Instance().Get<T>();
}

void LogWrite(std::ostream* os, const tableau::log::SourceLocation& loc, const tableau::log::LevelInfo& lvl,
              const std::string& content) {
  // clang-format off
  *os << tableau::log::NowStr() << " "
    // << std::this_thread::get_id() << "|"
    // << gettid() << " "
    << lvl.name << " [" 
    << loc.filename << ":" << loc.line << "][" 
    << loc.funcname << "]" 
    << content
    << std::endl << std::flush;
  // clang-format on
}

int main() {
  // custom log
  tableau::log::DefaultLogger()->SetWriter(LogWrite);

  tableau::Registry::Init();
  tableau::LoadOptions options;
  options.ignore_unknown_fields = true;
  bool ok = MyHub::Instance().Load(
      "../../testdata/", [](const std::string& name) { return true; }, tableau::Format::kJSON, &options);
  if (!ok) {
    std::cout << "protobuf hub load failed: " << tableau::GetErrMsg() << std::endl;
    return 1;
  }
  auto item_mgr = MyHub::Instance().Get<protoconf::ItemConfMgr>();
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

  //   auto activity_conf = MyHub::Instance().Get<tableau::ActivityConf>();
  //   if (!activity_conf) {
  //     std::cout << "protobuf hub get ActivityConf failed!" << std::endl;
  //     return 1;
  //   }

  //   const auto* section_conf = activity_conf->Get(100001, 1, 2);
  //   if (!section_conf) {
  //     std::cout << "ActivityConf get section failed!" << std::endl;
  //     return 1;
  //   }

  //   const auto* section_conf = MyHub::Instance().Get<protoconf::ActivityConfMgr, protoconf::Section>(100001, 1, 2);
  //   if (!section_conf) {
  //     std::cout << "ActivityConf get section failed!" << std::endl;
  //     return 1;
  //   }

  //   std::cout << "-----section_conf" << std::endl;
  //   std::cout << section_conf->DebugString() << std::endl;

  const auto* chapter_ordered_map =
      MyHub::Instance().GetOrderedMap<protoconf::ActivityConfMgr, tableau::ActivityConf::Activity_Chapter_OrderedMap>(
          100001);
  if (!chapter_ordered_map) {
    std::cout << "ActivityConf GetOrderedMap chapter failed!" << std::endl;
    return 1;
  }

  for (auto&& it : *chapter_ordered_map) {
    std::cout << "---" << it.first << "-----section_ordered_map" << std::endl;
    for (auto&& item : it.second.first) {
      std::cout << item.first << std::endl;
    }

    std::cout << "---" << it.first << " -----section_map" << std::endl;
    for (auto&& item : *it.second.second) {
      std::cout << item.first << std::endl;
    }
  }

  const auto* rank_ordered_map =
      MyHub::Instance().GetOrderedMap<protoconf::ActivityConfMgr, tableau::ActivityConf::int32_OrderedMap>(100001, 1,
                                                                                                           2);
  if (!rank_ordered_map) {
    std::cout << "ActivityConf GetOrderedMap rank failed!" << std::endl;
    return 1;
  }
  std::cout << "-----rank_ordered_map" << std::endl;
  for (auto&& it : *rank_ordered_map) {
    std::cout << it.first << std::endl;
  }

  auto activity_conf = MyHub::Instance().Get<tableau::ActivityConf>();
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
  return 0;
}