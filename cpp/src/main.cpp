#include <google/protobuf/util/json_util.h>
#include <iostream>
#include "protoconf/item.pb.h"
using google::protobuf::util::JsonStringToMessage;

bool proto_to_json(const google::protobuf::Message& message, std::string& json) {
    google::protobuf::util::JsonPrintOptions options;
    options.add_whitespace = true;
    options.always_print_primitive_fields = true;
    options.preserve_proto_field_names = true;
    return MessageToJsonString(message, &json, options).ok();
}

bool json_to_proto(const std::string& json, google::protobuf::Message& message) {
    return JsonStringToMessage(json, &message).ok();
}

int main() {
    protoconf::Item item;
    std::string json_string;

    item.set_id(100);
    item.set_name("coin");
    item.mutable_path()->set_dir("/home/protoconf/");
    item.mutable_path()->set_name("icon.png");

    /* protobuf 转 json。 */
    if (!proto_to_json(item, json_string)) {
        std::cout << "protobuf convert json failed!" << std::endl;
        return 1;
    }
    std::cout << "protobuf convert json done!" << std::endl
              << json_string << std::endl;

    item.Clear();
    std::cout << "-----" << std::endl;

    /* json 转 protobuf。 */
    if (!json_to_proto(json_string, item)) {
        std::cout << "json to protobuf failed!" << std::endl;
        return 1;
    }
    std::cout << "json to protobuf done!" << std::endl
              << "name: " << item.name() << std::endl
              << "dir: " << item.mutable_path()->dir()
              << std::endl;
    return 0;
}