
#pragma once
#include <google/protobuf/util/json_util.h>

#include <string>
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
typedef std::unordered_map<std::string, std::shared_ptr<Messager>> ConfigMap;
typedef std::shared_ptr<ConfigMap> ConfigMapPtr;
typedef std::function<bool(const std::string& name)> Filter;
typedef std::function<std::shared_ptr<Messager>()> MessagerGenerator;

class Hub {
  SINGLETON(Hub);

 public:
  void Init();
  template <typename T>
  void Register();
  bool Load(const std::string& dir, Filter filter, Format fmt = Format::kJSON);
  template <typename T>
  const std::shared_ptr<T> Get() const;

 private:
  ConfigMapPtr NewConfigMap();
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const { return (*config_map_ptr_)[name]; }

 private:
  ConfigMapPtr config_map_ptr_;
  // messager name -> messager generator
  std::unordered_map<std::string, MessagerGenerator> messager_map_;
};

template <typename T>
void Hub::Register() {
  messager_map_[T::Name()] = []() { return std::make_shared<T>(); };
}

template <typename T>
const std::shared_ptr<T> Hub::Get() const {
  auto msg = GetMessager(T::Name());
  return std::dynamic_pointer_cast<T>(msg);
}

// syntactic sugar
template <typename T>
const std::shared_ptr<T> Get() {
  return Hub::Instance().Get<T>();
}

}  // namespace tableau
