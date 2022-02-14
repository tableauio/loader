#include "demo/hub.tbx.h"

#include <google/protobuf/text_format.h>

#include <fstream>
#include <sstream>

#include "demo/registry.tbx.h"

namespace tableau {
static std::string g_err_msg;
const std::string& GetErrMsg() { return g_err_msg; }
bool Message2JSON(const google::protobuf::Message& message, std::string& json) {
  google::protobuf::util::JsonPrintOptions options;
  options.add_whitespace = true;
  options.always_print_primitive_fields = true;
  options.preserve_proto_field_names = true;
  return google::protobuf::util::MessageToJsonString(message, &json, options).ok();
}

bool JSON2Message(const std::string& json, google::protobuf::Message& message) {
  return google::protobuf::util::JsonStringToMessage(json, &message).ok();
}

bool Text2Message(const std::string& text, google::protobuf::Message& message) {
  return google::protobuf::TextFormat::ParseFromString(text, &message);
}
bool Wire2Message(const std::string& wire, google::protobuf::Message& message) { return message.ParseFromString(wire); }

const std::string& GetProtoName(const google::protobuf::Message& message) {
  const auto* md = message.GetDescriptor();
  return md != nullptr ? md->name() : kEmpty;
}

bool ReadFile(const std::string& filename, std::string& content) {
  std::ifstream file(filename);
  if (!file.is_open()) {
    g_err_msg = "Failed to open file: " + filename;
    return false;
  }
  std::stringstream ss;
  ss << file.rdbuf();
  content = std::move(ss.str());
  return true;
}

bool LoadMessage(const std::string& dir, google::protobuf::Message& message, Format fmt) {
  message.Clear();
  std::string basepath = dir + GetProtoName(message);
  // TODO: support 3 formats: json, text, and wire.
  std::string content;
  switch (fmt) {
    case Format::kJSON: {
      bool ok = ReadFile(basepath + ".json", content);
      if (!ok) {
        return false;
      }
      return JSON2Message(content, message);
    }
    case Format::kText: {
      bool ok = ReadFile(basepath + ".text", content);
      if (!ok) {
        return false;
      }
      return Text2Message(content, message);
    }
    case Format::kWire: {
      bool ok = ReadFile(basepath + ".wire", content);
      if (!ok) {
        return false;
      }
      return Wire2Message(content, message);
    }
    default: {
      g_err_msg = "Unsupported format: %d" + static_cast<int>(fmt);
      return false;
    }
  }
}

bool StoreMessage(const std::string& dir, google::protobuf::Message& message, Format fmt) {
  // TODO: write protobuf message to file, support 3 formats: json, text, and wire.
  return false;
}

bool Hub::Load(const std::string& dir, Filter filter, Format fmt) {
  auto new_config_map_ptr = NewConfigMap();
  for (auto iter : *new_config_map_ptr) {
    auto&& name = iter.first;
    bool yes = filter(name);
    if (!yes) continue;
    bool ok = iter.second->Load(dir, fmt);
    if (!ok) {
      g_err_msg = "Load " + name + " failed";
      return false;
    }
  }

  // replace
  config_map_ptr_ = new_config_map_ptr;
  return true;
}

ConfigMapPtr Hub::NewConfigMap() {
  ConfigMapPtr config_map_ptr = std::make_shared<ConfigMap>();
  for (auto&& it : Registry::registrar) {
    (*config_map_ptr)[it.first] = it.second();
  }
  return config_map_ptr;
}

}  // namespace tableau