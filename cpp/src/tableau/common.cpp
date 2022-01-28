#include "tableau/common.h"

#include <google/protobuf/util/json_util.h>

#include <fstream>
#include <sstream>

namespace tableau {
bool Proto2Json(const google::protobuf::Message& message, std::string& json) {
  google::protobuf::util::JsonPrintOptions options;
  options.add_whitespace = true;
  options.always_print_primitive_fields = true;
  options.preserve_proto_field_names = true;
  return google::protobuf::util::MessageToJsonString(message, &json, options).ok();
}

bool Json2Proto(const std::string& json, google::protobuf::Message& message) {
  return google::protobuf::util::JsonStringToMessage(json, &message).ok();
}

const std::string& GetProtoName(const google::protobuf::Message& message) {
  static const std::string empty = "";
  const auto* md = message.GetDescriptor();
  return md != nullptr ? md->name() : empty;
}

bool Load(const std::string& dir, google::protobuf::Message& message, Format fmt) {
  message.Clear();
  std::string filepath = dir + GetProtoName(message) + ".json";
  std::ifstream fs(filepath);
  std::stringstream ss;
  ss << fs.rdbuf();
  // TODO: support 3 formats: json, prototext, and protowire.
  std::string json(ss.str());
  return Json2Proto(json, message);
}
}  // namespace tableau