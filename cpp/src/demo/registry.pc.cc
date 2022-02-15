

#include "demo/registry.pc.h"

#include "demo/item.pc.h"

namespace tableau {
Registrar Registry::registrar = Registrar();
void Registry::Init() {
  Register<Item>();
  // TODO: register more here
}

}  // namespace tableau