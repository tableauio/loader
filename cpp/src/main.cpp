#include <iostream>
#include "item.tableau.h"
#include <fstream>
#include <string>
#include "common/loader.h"

void WriteFile(const std::string& filename, const std::string& input){
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

    if (!tableau::Proto2Json(item, json_string)) {
        std::cout << "protobuf convert json failed!" << std::endl;
        return 1;
    }
    std::cout << "protobuf convert json done!" << std::endl
              << json_string << std::endl;

    item.Clear();

    // std::cout << "-----" << std::endl;
    // if (!tableau::Json2Proto(json_string, item)) {
    //     std::cout << "json to protobuf failed!" << std::endl;
    //     return 1;
    // }
    // std::cout << "json to protobuf done!" << std::endl
    //           << "name: " << item.name() << std::endl
    //           << "dir: " << item.mutable_path()->dir()
    //           << std::endl;
    
    std::cout << "-----" << std::endl;
    tableau::ItemLoader item_loader;
    if (!item_loader.Load("../testdata/", tableau::Format::kJSON)) {
        std::cout << "protobuf load json failed!" << std::endl;
        return 1;
    }
    item = item_loader.GetConf();
    std::cout << "item: " << item.DebugString() << std::endl;

    json_string.clear();
     if (!tableau::Proto2Json(item, json_string)) {
        std::cout << "protobuf convert json failed!" << std::endl;
        return 1;
    }
    WriteFile("./test_item.json", json_string);
    return 0;
}