import (
	"time"
)

type MessagerContainer struct {
	messagerMap MessagerMap
	loadedTime  time.Time
	// all messagers as fields for fast access
{{ range .Messagers }}	{{ toLowerCamel . }} *{{ . }}
{{ end }}}

func newMessagerContainer(messagerMap MessagerMap) *MessagerContainer {
	return &MessagerContainer{
		messagerMap: messagerMap,
		loadedTime:  time.Now(),
{{ if .Go127 -}}
{{ range .Messagers }}		{{ toLowerCamel . }}: getMessager[*{{ . }}](messagerMap),
{{ end -}}
{{ else -}}
{{ range .Messagers }}		{{ toLowerCamel . }}: GetMessager[*{{ . }}](messagerMap),
{{ end -}}
{{ end -}}
	}
}

func (mc *MessagerContainer) GetMessagerMap() MessagerMap {
	return mc.messagerMap
}

func (mc *MessagerContainer) GetMessager(name string) Messager {
	return mc.messagerMap[name]
}

{{ if .Go127 }}
// Get returns the messager of type T, or the zero value if not found.
func (mc *MessagerContainer) Get[T Messager]() T {
	return getMessager[T](mc.messagerMap)
}
{{ end }}
func (mc *MessagerContainer) GetLastLoadedTime() time.Time {
	return mc.loadedTime
}
// Auto-generated getters below
{{ range .Messagers }}
func (mc *MessagerContainer) Get{{ . }}() *{{ . }} {
	return mc.{{ toLowerCamel . }}
}
{{ end }}
