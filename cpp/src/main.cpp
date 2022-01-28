#include <fstream>
#include <iostream>
#include <string>

#include "demo/hub.h"
#include "protoconf/item.pb.h"

void WriteFile(const std::string& filename, const std::string& input) {
  std::ofstream out(filename);
  out << input;
  out.close();
}

int main() {
  protoconf::Item item;
  std::string json_string;

  item.set_id(100);
  item.set_name("coin");
  item.mutable_path()->set_dir("/home/protoconf/");
  item.mutable_path()->set_name("icon.png");

  if (!demo::tableau::Message2JSON(item, json_string)) {
    std::cout << "protobuf convert json failed!" << std::endl;
    return 1;
  }
  std::cout << "protobuf convert json done!" << std::endl << json_string << std::endl;

  item.Clear();

  // std::cout << "-----" << std::endl;
  // if (!demo::tableau::Json2Proto(json_string, item)) {
  //     std::cout << "json to protobuf failed!" << std::endl;
  //     return 1;
  // }
  // std::cout << "json to protobuf done!" << std::endl
  //           << "name: " << item.name() << std::endl
  //           << "dir: " << item.mutable_path()->dir()
  //           << std::endl;

  std::cout << "-----" << std::endl;
  bool ok = demo::tableau::Hub::Instance().Load("../testdata/", [](const std::string& name) { return true; });
  if (!ok) {
    std::cout << "protobuf hub load failed!" << std::endl;
    return 1;
  }
  auto item1 = demo::tableau::Get<protoconf::Item>();
  if (!item1) {
    std::cout << "protobuf hub get Item failed!" << std::endl;
    return 1;
  }
  std::cout << "item1: " << item1->DebugString() << std::endl;

  json_string.clear();
  if (!demo::tableau::Message2JSON(*item1, json_string)) {
    std::cout << "protobuf convert json failed!" << std::endl;
    return 1;
  }
  WriteFile("./test_item.json", json_string);
  return 0;
}