#include <fstream>
#include <iostream>
#include <string>

#include "demo/hub.tbx.h"
#include "demo/item.tbx.h"
#include "demo/registry.tbx.h"
#include "protoconf/item.pb.h"

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
  protoconf::Item item;
  std::string jsonstr;

  item.set_id(100);
  item.set_name("coin");
  item.mutable_path()->set_dir("/home/protoconf/");
  item.mutable_path()->set_name("icon.png");

  if (!tableau::Message2JSON(item, jsonstr)) {
    std::cout << "protobuf convert json failed!" << std::endl;
    return 1;
  }
  std::cout << "protobuf convert json done!" << std::endl << jsonstr << std::endl;

  item.Clear();

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
  bool ok = MyHub::Instance().Load("../testdata/", [](const std::string& name) { return true; });
  if (!ok) {
    std::cout << "protobuf hub load failed!" << std::endl;
    return 1;
  }
  auto item1 = MyHub::Instance().Get<tableau::Item>();
  if (!item1) {
    std::cout << "protobuf hub get Item failed!" << std::endl;
    return 1;
  }
  std::cout << "item1: " << item1->Get().DebugString() << std::endl;

  jsonstr.clear();
  if (!tableau::Message2JSON(item1->Get(), jsonstr)) {
    std::cout << "protobuf convert json failed!" << std::endl;
    return 1;
  }
  WriteFile("./test_item.json", jsonstr);
  return 0;
}