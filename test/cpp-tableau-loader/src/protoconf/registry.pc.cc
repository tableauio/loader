// Code generated by protoc-gen-cpp-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-cpp-tableau-loader v0.6.0
// - protoc                        v3.19.3

#include "registry.pc.h"

namespace tableau {
Registrar Registry::registrar = Registrar();
void Registry::Init() {
  InitShard0();
  InitShard1();
}
}  // namespace tableau
