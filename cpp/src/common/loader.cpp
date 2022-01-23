#include "common/loader.h"

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
}  // namespace tableau