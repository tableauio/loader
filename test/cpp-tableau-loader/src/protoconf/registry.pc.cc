// Code generated by protoc-gen-cpp-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-cpp-tableau-loader v0.4.7
// - protoc                        v3.19.3

#include "registry.pc.h"

#include "hero_conf.pc.h"
#include "item_conf.pc.h"
#include "test_conf.pc.h"

namespace tableau {
Registrar Registry::registrar = Registrar();
void Registry::Init() {
  Register<HeroConf>();
  Register<ItemConf>();
  Register<ActivityConf>();
  Register<ChapterConf>();
  Register<ThemeConf>();
}
}  // namespace tableau
