import "time"

type MessagerContainer interface {
	GetMessagerMap() MessagerMap
	GetMessager(name string) Messager
	GetLastLoadedTime() time.Time

	// Auto-generated getters below
{{ range . }} Get{{ . }}() *{{ . }}
{{ end }}}

type messagerContainer struct {
	messagerMap MessagerMap
	loadedTime  time.Time
	// all messagers as fields for fast access
{{ range . }}	{{ toLowerCamel . }} *{{ . }}
{{ end }}}

func newMessagerContainer(messagerMap MessagerMap) *messagerContainer {
	messagerContainer := &messagerContainer{
		messagerMap: messagerMap,
		loadedTime:  time.Now(),
	}
{{ range . }}	messagerContainer.{{ toLowerCamel . }}, _ = messagerMap[(&{{ . }}{}).Name()].(*{{ . }})
{{ end }}	return messagerContainer
}

func (mc *messagerContainer) GetMessagerMap() MessagerMap {
	return mc.messagerMap
}

func (mc *messagerContainer) GetMessager(name string) Messager {
	return mc.messagerMap[name]
}

func (mc *messagerContainer) GetLastLoadedTime() time.Time {
	return mc.loadedTime
}
{{ range . }}
func (mc *messagerContainer) Get{{ . }}() *{{ . }} {
	return mc.{{ toLowerCamel . }}
}
{{ end }}
