#pragma once

#include "common/loader.h"
#include "protoconf/item.pb.h"

namespace tableau {
class ItemLoader : public Loader {
 public:
  virtual bool Load(const std::string& dirpath, Format fmt = Format::kJSON) override;
  virtual const std::string& GetName() const override;
  const protoconf::Item& GetConf() const { return conf_; }

 private:
  protoconf::Item conf_;
};
}  // namespace tableau