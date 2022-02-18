#pragma once
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
using ConfigMap = std::unordered_map<std::string, std::shared_ptr<Messager>>;
using ConfigMapPtr = std::shared_ptr<ConfigMap>;
using Filter = std::function<bool(const std::string& name)>;

class Hub {
 public:
  bool Load(const std::string& dir, Filter filter = nullptr, Format fmt = Format::kJSON);
  template <typename T>
  const std::shared_ptr<T> Get() const;
  template <typename T, typename U, typename... Args>
  const U* Get(Args... args) const;

 private:
  ConfigMapPtr NewConfigMap(Filter filter = nullptr);
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const { return (*config_map_ptr_)[name]; }

 private:
  ConfigMapPtr config_map_ptr_;
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
  if (!msger) {
    return nullptr;
  }
  return msger->Get(args...);
}

}  // namespace tableau
