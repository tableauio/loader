import (
	"time"
)

type MessagerContainer struct {
	messagerMap MessagerMap
	loadedTime  time.Time
	// all messagers as fields for fast access
{{ range . }}	{{ toLowerCamel . }} *{{ . }}
{{ end }}}

func newMessagerContainer(messagerMap MessagerMap) *MessagerContainer {
	return &MessagerContainer{
		messagerMap: messagerMap,
		loadedTime:  time.Now(),
{{ range . }}		{{ toLowerCamel . }}: GetMessager[*{{ . }}](messagerMap),
{{ end }}	}
}

func (mc *MessagerContainer) GetMessagerMap() MessagerMap {
	return mc.messagerMap
}

func (mc *MessagerContainer) GetMessager(name string) Messager {
	return mc.messagerMap[name]
}

func (mc *MessagerContainer) GetLastLoadedTime() time.Time {
	return mc.loadedTime
}
// Auto-generated getters below
{{ range . }}
func (mc *MessagerContainer) Get{{ . }}() *{{ . }} {
	return mc.{{ toLowerCamel . }}
}
{{ end }}
