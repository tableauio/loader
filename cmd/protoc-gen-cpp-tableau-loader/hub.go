package main

import "google.golang.org/protobuf/compiler/protogen"

// generateHub generates related registry files.
func generateHub(gen *protogen.Plugin) {
	hppFilename := "hub." + pcExt + ".h"
	g1 := gen.NewGeneratedFile(hppFilename, "")
	generateCommonHeader(gen, g1)
	g1.P()
	g1.P(hubHpp)

	cppFilename := "hub." + pcExt + ".cc"
	g2 := gen.NewGeneratedFile(cppFilename, "")
	generateCommonHeader(gen, g2)
	g2.P()
	g2.P(hubCpp)
}

const hubHpp = `#pragma once
#include <google/protobuf/util/json_util.h>

#include <string>

namespace tableau {
enum class Format {
  kJSON,
  kText,
  kWire,
};

constexpr const char* kJSONExt = ".json";
constexpr const char* kTextExt = ".text";
constexpr const char* kWireExt = ".wire";

static const std::string kEmpty = "";
const std::string& GetErrMsg();

bool Message2JSON(const google::protobuf::Message& message, std::string& json);
bool JSON2Message(const std::string& json, google::protobuf::Message& message);
bool Text2Message(const std::string& text, google::protobuf::Message& message);
bool Wire2Message(const std::string& wire, google::protobuf::Message& message);

const std::string& GetProtoName(const google::protobuf::Message& message);
bool LoadMessage(const std::string& dir, google::protobuf::Message& message, Format fmt = Format::kJSON);

class Messager {
 public:
  virtual ~Messager() = default;
  static const std::string& Name() { return kEmpty; };
  virtual bool Load(const std::string& dir, Format fmt) = 0;
};
using MessagerMap = std::unordered_map<std::string, std::shared_ptr<Messager>>;
using MessagerMapPtr = std::shared_ptr<MessagerMap>;
using Filter = std::function<bool(const std::string& name)>;

class Hub {
 public:
  bool Load(const std::string& dir, Filter filter = nullptr, Format fmt = Format::kJSON);

  template <typename T>
  const std::shared_ptr<T> Get() const;

  template <typename T, typename U, typename... Args>
  const U* Get(Args... args) const;

  template <typename T, typename U, typename... Args>
  const U* GetOrderedMap(Args... args) const;

 private:
  MessagerMapPtr NewMessagerMap(Filter filter = nullptr);
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const { return (*messager_map_ptr_)[name]; }

 private:
  MessagerMapPtr messager_map_ptr_;
};

template <typename T>
const std::shared_ptr<T> Hub::Get() const {
  auto msg = GetMessager(T::Name());
  return std::dynamic_pointer_cast<T>(msg);
}

template <typename T, typename U, typename... Args>
const U* Hub::Get(Args... args) const {
  auto msg = GetMessager(T::Name());
  auto msger = std::dynamic_pointer_cast<T>(msg);
  return msger ? msger->Get(args...) : nullptr;
}

template <typename T, typename U, typename... Args>
const U* Hub::GetOrderedMap(Args... args) const {
  auto msg = GetMessager(T::Name());
  auto msger = std::dynamic_pointer_cast<T>(msg);
  return msger ? msger->GetOrderedMap(args...) : nullptr;
}

}  // namespace tableau`

const hubCpp = `#include "hub.pc.h"

#include <google/protobuf/text_format.h>

#include <fstream>
#include <sstream>

#include "registry.pc.h"

namespace tableau {
static thread_local std::string g_err_msg;
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
      bool ok = ReadFile(basepath + kJSONExt, content);
      if (!ok) {
        return false;
      }
      return JSON2Message(content, message);
    }
    case Format::kText: {
      bool ok = ReadFile(basepath + kTextExt, content);
      if (!ok) {
        return false;
      }
      return Text2Message(content, message);
    }
    case Format::kWire: {
      bool ok = ReadFile(basepath + kWireExt, content);
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
  auto new_messager_map_ptr = NewMessagerMap(filter);
  for (auto iter : *new_messager_map_ptr) {
    auto&& name = iter.first;
    bool ok = iter.second->Load(dir, fmt);
    if (!ok) {
      g_err_msg = "Load " + name + " failed";
      return false;
    }
  }

  // replace
  messager_map_ptr_ = new_messager_map_ptr;
  return true;
}

MessagerMapPtr Hub::NewMessagerMap(Filter filter) {
  MessagerMapPtr messager_map_ptr = std::make_shared<MessagerMap>();
  for (auto&& it : Registry::registrar) {
    if (filter == nullptr || filter(it.first)) {
      (*messager_map_ptr)[it.first] = it.second();
    }
  }
  return messager_map_ptr;
}

}  // namespace tableau`
