
#pragma once
#include <google/protobuf/util/json_util.h>

#include <iostream>
#include <string>

#include "tableau/common.h"
namespace protoconf {
typedef std::unordered_map<std::string, std::shared_ptr<google::protobuf::Message>> MessageMap;
class Hub {
  SINGLETON(Hub);

  /*
    //  public:
    // #define GETTER(ClassName)                                   \
  //   const std::shared_ptr<ClassName> Get #ClassName() {       \
  //     auto conf = Hub::Instance().GetConf(ClassName);         \
  //     return std::dynamic_pointer_cast<ClassName>(ClassName); \
  //   }
    //   GETTER(Item);
    // #undef GETTER
  */

 public:
  bool Load(const std::string& dir, tableau::Filter filter, tableau::Format fmt = tableau::Format::kJSON);
  template <typename T>
  const std::shared_ptr<T> Get() const;

 private:
  static std::shared_ptr<MessageMap> NewConfMap();
  const std::shared_ptr<google::protobuf::Message> GetConf(const std::string& proto_name) const {
    return (*conf_map_)[proto_name];
  }

 private:
  std::string errmsg_;
  std::shared_ptr<MessageMap> conf_map_;
};

template <typename T>
const std::shared_ptr<T> Hub::Get() const {
  auto t = T();
  std::string name = tableau::GetProtoName(t);
  std::cout << "Get proto name: " << name << std::endl;
  auto conf = GetConf(name);
  if (conf) {
    std::cout << "conf debug string: " << conf->DebugString() << std::endl;
  }
  return std::dynamic_pointer_cast<T>(conf);
}

// syntactic sugar
template <typename T>
const std::shared_ptr<T> Get() {
  return Hub::Instance().Get<T>();
}

}  // namespace protoconf
