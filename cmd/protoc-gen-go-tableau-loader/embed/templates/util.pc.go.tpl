import (
	"errors"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/tableauio/tableau/store"
	"google.golang.org/protobuf/proto"
)

var ErrNotFound = errors.New("not found")

func boolToInt(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

// GetMessager gets a messager from provided [MessagerMap]. It will return nil
// if not found by messager name.
func GetMessager[T Messager](messagerMap MessagerMap) T {
	var t T
	messager, _ := messagerMap[t.Name()].(T)
	return messager
}

// UnifiedDiff generates the proto message delta as a unified diff.
func UnifiedDiff(original, current proto.Message) (string, error) {
	originalText, err := store.MarshalToText(original, true)
	if err != nil {
		return "", err
	}
	currentText, err := store.MarshalToText(current, true)
	if err != nil {
		return "", err
	}
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(originalText)),
		B:        difflib.SplitLines(string(currentText)),
		FromFile: "Original",
		ToFile:   "Current",
		Context:  3,
	}
	return difflib.GetUnifiedDiffString(diff)
}