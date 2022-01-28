
#pragma once
#include <google/protobuf/util/json_util.h>

#include <string>

namespace demo {
namespace tableau {
#define SINGLETON(ClassName)          \
 private:                             \
  ClassName() {}                      \
  ClassName(const ClassName&) {}      \
  void operator=(const ClassName&) {} \
                                      \
 public:                              \
  static ClassName& Instance() {      \
    static ClassName instance;        \
    return instance;                  \
  }

enum class Format {
  kJSON,
  kText,
  kWire,
};

const std::string& GetErrMsg();

bool Message2JSON(const google::protobuf::Message& message, std::string& json);
bool JSON2Message(const std::string& json, google::protobuf::Message& message);
bool Text2Message(const std::string& text, google::protobuf::Message& message);
bool Wire2Message(const std::string& wire, google::protobuf::Message& message);

const std::string& GetProtoName(const google::protobuf::Message& message);
bool LoadMessage(const std::string& dir, google::protobuf::Message& message, Format fmt = Format::kJSON);

typedef std::function<bool(const std::string& proto_name)> Filter;

typedef std::unordered_map<std::string, std::shared_ptr<google::protobuf::Message>> ConfigMap;
typedef std::shared_ptr<ConfigMap> ConfigMapPtr;

class Hub {
  SINGLETON(Hub);

 public:
  bool Load(const std::string& dir, Filter filter, Format fmt = Format::kJSON);
  template <typename T>
  const std::shared_ptr<T> Get() const;

 private:
  static ConfigMapPtr NewConfigMap();
  const std::shared_ptr<google::protobuf::Message> GetConf(const std::string& proto_name) const {
    return (*config_map_ptr_)[proto_name];
  }

 private:
  ConfigMapPtr config_map_ptr_;
};

template <typename T>
const std::shared_ptr<T> Hub::Get() const {
  auto t = T();
  auto conf = GetConf(GetProtoName(t));
  return std::dynamic_pointer_cast<T>(conf);
}

// syntactic sugar
template <typename T>
const std::shared_ptr<T> Get() {
  return Hub::Instance().Get<T>();
}

}  // namespace tableau
}  // namespace demo
