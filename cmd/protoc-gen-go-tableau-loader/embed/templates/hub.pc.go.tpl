import "time"

type Hub interface {
	GetMessagerMap() MessagerMap
	GetLastLoadedTime() time.Time

	// Auto-generated getters below
{{ range . }}	Get{{ . }}() *{{ . }}
{{ end }}}