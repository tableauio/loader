package udiff

import (
	"github.com/aymanbagabas/go-udiff"
	"github.com/tableauio/tableau/store"
	"google.golang.org/protobuf/proto"
)

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
	return udiff.Unified("Original", "Current", string(originalText), string(currentText)), nil
}
