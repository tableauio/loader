
#pragma once
#include "hub.pc.h"
namespace tableau {
using MessagerGenerator = std::function<std::shared_ptr<Messager>()>;
// messager name -> messager generator
using Registrar = std::unordered_map<std::string, MessagerGenerator>;
class Registry {
 public:
  static void Init();
  template <typename T>
  static void Register();

  static Registrar registrar;
};

template <typename T>
void Registry::Register() {
  registrar[T::Name()] = []() { return std::make_shared<T>(); };
}

}  // namespace tableau
