

#include "demo/registry.tbx.h"

#include "demo/item.tbx.h"

namespace tableau {
Registrar Registry::registrar = Registrar();
void Registry::Init() {
  Register<Item>();
  // TODO: register more here
}

}  // namespace tableau