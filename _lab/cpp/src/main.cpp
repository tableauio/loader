#include <fstream>
#include <iostream>
#include <string>

#include "demo/hub.pc.h"
#include "demo/item_conf.pc.h"
#include "demo/logger.pc.h"
#include "demo/registry.pc.h"
#include "demo/test_conf.pc.h"
#include "protoconf/item_conf.pb.h"

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

int main() {
  std::string jsonstr;
  //   protoconf::ItemConf item;

  //   item.set_id(100);
  //   item.set_name("coin");
  //   item.mutable_path()->set_dir("/home/protoconf/");
  //   item.mutable_path()->set_name("icon.png");

  //   if (!tableau::Message2JSON(item, jsonstr)) {
  //     std::cout << "protobuf convert json failed!" << std::endl;
  //     return 1;
  //   }
  //   std::cout << "protobuf convert json done!" << std::endl << jsonstr << std::endl;

  //   item.Clear();

  // std::cout << "-----" << std::endl;
  // if (!tableau::Json2Proto(jsonstr, item)) {
  //     std::cout << "json to protobuf failed!" << std::endl;
  //     return 1;
  // }
  // std::cout << "json to protobuf done!" << std::endl
  //           << "name: " << item.name() << std::endl
  //           << "dir: " << item.mutable_path()->dir()
  //           << std::endl;

  std::cout << "-----" << std::endl;
  tableau::Registry::Init();
  protoconf::ActivityConf act_conf;
  tableau::ProtobufLogHandler(google::protobuf::LOGLEVEL_INFO, "info.cc", 10, "info msg");
  tableau::ProtobufLogHandler(google::protobuf::LOGLEVEL_WARNING, "warn.cc", 10, "warn msg");
  tableau::ProtobufLogHandler(google::protobuf::LOGLEVEL_ERROR, "error.cc", 10, "error msg");
  tableau::ProtobufLogHandler(google::protobuf::LOGLEVEL_FATAL, "fatal.cc", 10, "fatal msg");
  // google::protobuf::SetLogHandler(tableau::ProtobufLogHandler);
  //   if (!act_conf.ParseFromString("0101010110")) {
  //     std::cout << "failed to parse" << std::endl;
  //     return 1;
  //   }

  bool ok = MyHub::Instance().Load("../../../test/testdata/", [](const std::string& name) { return true; });
  if (!ok) {
    std::cout << "protobuf hub load failed: " << tableau::GetErrMsg() << std::endl;
    return 1;
  }
  auto item1 = MyHub::Instance().Get<tableau::ItemConf>();
  if (!item1) {
    std::cout << "protobuf hub get Item failed!" << std::endl;
    return 1;
  }
  std::cout << "item1: " << item1->Data().DebugString() << std::endl;

  jsonstr.clear();
  if (!tableau::Message2JSON(item1->Data(), jsonstr)) {
    std::cout << "protobuf convert json failed!" << std::endl;
    return 1;
  }
  WriteFile("./test_item.json", jsonstr);

  auto activity_conf = MyHub::Instance().Get<tableau::ActivityConf>();
  if (!activity_conf) {
    std::cout << "protobuf hub get ActivityConf failed!" << std::endl;
    return 1;
  }

  // std::cout << "acitivity: " << activity_conf->Data().DebugString() << std::endl;

  auto section_conf = activity_conf->Get(100001, 1, 2);
  if (!section_conf) {
    std::cout << "ActivityConf get section failed!" << std::endl;
    return 1;
  }

  section_conf = MyHub::Instance().Get<protoconf::ActivityConfMgr, protoconf::Section>(100001, 1, 2);
  if (!section_conf) {
    std::cout << "ActivityConf Get section failed: " << tableau::GetErrMsg() << std::endl;
    return 1;
  }

  std::cout << "-----section_conf" << std::endl;
  std::cout << section_conf->DebugString() << std::endl;

  auto chapter_ordered_map =
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

  auto section_ordered_map =
      MyHub::Instance().GetOrderedMap<protoconf::ActivityConfMgr, tableau::ActivityConf::protoconf_Section_OrderedMap>(
          100001, 1);
  if (!section_ordered_map) {
    std::cout << "ActivityConf GetOrderedMap section failed!" << std::endl;
    return 1;
  }

  std::cout << "-----section_ordered_map" << std::endl;
  for (auto&& item : *section_ordered_map) {
    std::cout << item.first << std::endl;
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
