
#pragma once
#include <google/protobuf/util/json_util.h>
#include <iostream>
#include "common/common.h"
namespace tableau {
static const std::string empty = "";
bool Proto2Json(const google::protobuf::Message& message, std::string& json);
bool Json2Proto(const std::string& json, google::protobuf::Message& message);
class Loader {
 public:
  virtual ~Loader() = default;
  virtual bool Load(const std::string& dirpath, Format fmt = Format::kJSON) = 0;
  virtual const std::string& GetName() const = 0;
};
}  // namespace tableau