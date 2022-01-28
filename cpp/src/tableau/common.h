#pragma once
#include <google/protobuf/message.h>

#include <string>

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

namespace tableau {
enum class Format {
  kJSON,
  kProtowire,
  kPrototext,
};

struct Result {
  int code;
  std::string msg;
};
typedef std::function<bool(const std::string& proto_name)> Filter;

const std::string& GetProtoName(const google::protobuf::Message& message);
bool Proto2Json(const google::protobuf::Message& message, std::string& json);
bool Json2Proto(const std::string& json, google::protobuf::Message& message);
bool Load(const std::string& dir, google::protobuf::Message& message, Format fmt = Format::kJSON);

}  // namespace tableau