#include "item.pc.h"

namespace tableau {
const std::string Item::kProtoName = "Item";
bool Item::Load(const std::string& dir, Format fmt) { return LoadMessage(dir, data_, fmt); }
}  // namespace tableau
