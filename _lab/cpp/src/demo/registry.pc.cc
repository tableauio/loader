

#include "registry.pc.h"

#include "item_conf.pc.h"
#include "test_conf.pc.h"

namespace tableau {
Registrar Registry::registrar = Registrar();
void Registry::Init() {
  Register<ItemConf>();
  Register<ActivityConf>();
  // TODO: register more here
}

}  // namespace tableau