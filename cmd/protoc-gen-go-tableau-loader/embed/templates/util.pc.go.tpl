import (
	"errors"
)

var ErrNotFound = errors.New("not found")

func boolToInt(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

{{ if .Go127 -}}
// Deprecated: Use [Hub.Get] or [MessagerContainer.Get] instead.
func GetMessager[T Messager](messagerMap MessagerMap) T {
	return getMessager[T](messagerMap)
}

func getMessager[T Messager](messagerMap MessagerMap) T {
	var t T
	messager, _ := messagerMap[t.Name()].(T)
	return messager
}
{{- else -}}
// GetMessager gets a messager from provided [MessagerMap]. It will return nil
// if not found by messager name.
func GetMessager[T Messager](messagerMap MessagerMap) T {
	var t T
	messager, _ := messagerMap[t.Name()].(T)
	return messager
}
{{- end }}
