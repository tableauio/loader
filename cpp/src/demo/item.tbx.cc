#include "demo/item.tbx.h"
namespace tableau {
// static Item __item_register;
const std::string Item::kProtoName = "Item";

Item::Item() { Hub::Instance().Register<Item>(); }
bool Item::Load(const std::string& dir, Format fmt) { return LoadMessage(dir, item_, fmt); }

}  // namespace tableau