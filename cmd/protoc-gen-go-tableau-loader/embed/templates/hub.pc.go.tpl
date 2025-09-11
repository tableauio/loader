import "time"

type Hub interface {
	GetMessagerMap() MessagerMap
	GetMessager(name string) Messager
	GetLastLoadedTime() time.Time

	// Auto-generated getters below
{{ range . }}	Get{{ . }}() *{{ . }}
{{ end }}}