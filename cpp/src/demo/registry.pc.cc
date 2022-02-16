

#include "registry.pc.h"

#include "item.pc.h"
#include "test.pc.h"

namespace tableau {
Registrar Registry::registrar = Registrar();
void Registry::Init() {
  Register<Item>();
  Register<ActivityConf>();
  // TODO: register more here
}

}  // namespace tableau