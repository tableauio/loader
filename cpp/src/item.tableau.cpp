#include "item.tableau.h"
#include "common/loader.h"
#include "protoconf/item.pb.h"
#include <sstream>
#include <fstream>
#include <iostream>

namespace tableau {
bool ItemLoader::Load(const std::string& dirpath, Format fmt) {
  conf_.Clear();
  std::string filepath = dirpath + GetName() + ".json";
  std::ifstream fs(filepath);
  std::stringstream ss;
  ss << fs.rdbuf();
  std::string json(ss.str());
  return Json2Proto(json, conf_);
}

const std::string& ItemLoader::GetName() const {
  auto* d = protoconf::Item::GetDescriptor();
  return d != nullptr ? d->name() : empty;
}
}  // namespace tableau