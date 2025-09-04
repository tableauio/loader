#include "hub.pc.h"

// clang-format off
{{ range .Protofiles }}#include "{{ toSnake .Name }}.pc.h"
{{ end }}// clang-format on

namespace tableau {{ "{" }}{{ range .Protofiles }}{{ range .Messagers }}
template <>
const std::shared_ptr<{{ . }}> Hub::Get<{{ . }}>() const {
  return GetMessagerContainerWithProvider()->{{ toSnake . }}_;
}
{{ end }}{{ end }}
void MessagerContainer::InitShard{{ .Shard }}() {{ "{" }}{{ range .Protofiles }}{{ range .Messagers }}
  {{ toSnake . }}_ = std::dynamic_pointer_cast<{{ . }}>(GetMessager({{ . }}::Name()));{{ end }}{{ end }}
}

void Registry::InitShard{{ .Shard }}() {{ "{" }}{{ range .Protofiles }}{{ range .Messagers }}
  Register<{{ . }}>();{{ end }}{{ end }}
}
}  // namespace tableau
