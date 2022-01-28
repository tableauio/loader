#include "demo/hub.h"

#include "protoconf/item.pb.h"
#include "tableau/common.h"

namespace protoconf {
std::shared_ptr<MessageMap> Hub::NewConfMap() {
  std::shared_ptr<MessageMap> conf_map = std::make_shared<MessageMap>();

#define REGISTER(ClassName) (*conf_map)[#ClassName] = std::make_shared<ClassName>();

  REGISTER(Item);

#undef REGISTER

  return conf_map;
}
bool Hub::Load(const std::string& dir, tableau::Filter filter, tableau::Format fmt) {
  auto new_conf_map = NewConfMap();

  for (auto iter : *new_conf_map) {
    std::string name = iter.first;
    std::cout << "conf map item: " << name << std::endl;
    auto conf = iter.second;
    bool yes = filter(name);
    if (yes) continue;
    bool ok = tableau::Load(dir, *conf, fmt);
    if (!ok) {
      errmsg_ = "Load " + name + " failed";
      return false;
    }
  }

  // replace
  conf_map_ = new_conf_map;
  return true;
}

}  // namespace protoconf