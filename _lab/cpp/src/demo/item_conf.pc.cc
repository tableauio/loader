#include "item_conf.pc.h"

namespace tableau {
const std::string ItemConf::kProtoName = "Item";
bool ItemConf::Load(const std::string& dir, Format fmt) {
  bool ok = LoadMessage(dir, data_, fmt);
  return ok ? ProcessAfterLoad() : false;
}
}  // namespace tableau
