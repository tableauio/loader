package udiff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestUnifiedDiff_NoDifference(t *testing.T) {
	original := wrapperspb.String("hello")
	current := wrapperspb.String("hello")

	diff, err := UnifiedDiff(original, current)
	require.NoError(t, err)
	assert.Empty(t, diff)
}

func TestUnifiedDiff_WithDifference(t *testing.T) {
	original := wrapperspb.String("hello")
	current := wrapperspb.String("world")

	diff, err := UnifiedDiff(original, current)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)

	// Assert full unified diff lines
	assert.Contains(t, diff, "--- Original")
	assert.Contains(t, diff, "+++ Current")
	assert.Contains(t, diff, `-value: "hello"`)
	assert.Contains(t, diff, `+value: "world"`)
}

func TestUnifiedDiff_NilMessages(t *testing.T) {
	original := wrapperspb.String("")
	current := wrapperspb.String("")

	diff, err := UnifiedDiff(original, current)
	require.NoError(t, err)
	assert.Empty(t, diff)
}

func TestUnifiedDiff_ComplexMessage(t *testing.T) {
	original, err := structpb.NewStruct(map[string]any{
		"name":    "Alice",
		"age":     30,
		"active":  true,
		"score":   99.5,
		"address": map[string]any{"city": "Shanghai", "zip": "200000"},
		"tags":    []any{"admin", "vip"},
	})
	require.NoError(t, err)

	current, err := structpb.NewStruct(map[string]any{
		"name":    "Bob",
		"age":     25,
		"active":  false,
		"score":   88.0,
		"address": map[string]any{"city": "Beijing", "zip": "100000"},
		"tags":    []any{"user"},
	})
	require.NoError(t, err)

	diff, err := UnifiedDiff(original, current)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)

	// Assert changed fields with full unified diff lines
	assert.Contains(t, diff, "-    bool_value: true")
	assert.Contains(t, diff, "+    bool_value: false")
	assert.Contains(t, diff, `-          string_value: "Shanghai"`)
	assert.Contains(t, diff, `+          string_value: "Beijing"`)
	assert.Contains(t, diff, `-          string_value: "200000"`)
	assert.Contains(t, diff, `+          string_value: "100000"`)
	assert.Contains(t, diff, "-    number_value: 30")
	assert.Contains(t, diff, "+    number_value: 25")
	assert.Contains(t, diff, `-    string_value: "Alice"`)
	assert.Contains(t, diff, `+    string_value: "Bob"`)
	assert.Contains(t, diff, "-    number_value: 99.5")
	assert.Contains(t, diff, "+    number_value: 88")
	assert.Contains(t, diff, `-        string_value: "admin"`)
	assert.Contains(t, diff, `-        string_value: "vip"`)
	assert.Contains(t, diff, `+        string_value: "user"`)

	// Assert unchanged fields (keys) are not in diff lines (no +/- prefix)
	assert.NotContains(t, diff, `-  key: "active"`)
	assert.NotContains(t, diff, `+  key: "active"`)
	assert.NotContains(t, diff, `-  key: "address"`)
	assert.NotContains(t, diff, `+  key: "address"`)
	assert.NotContains(t, diff, `-  key: "age"`)
	assert.NotContains(t, diff, `+  key: "age"`)
	assert.NotContains(t, diff, `-  key: "name"`)
	assert.NotContains(t, diff, `+  key: "name"`)
	assert.NotContains(t, diff, `-  key: "score"`)
	assert.NotContains(t, diff, `+  key: "score"`)
	assert.NotContains(t, diff, `-  key: "tags"`)
	assert.NotContains(t, diff, `+  key: "tags"`)
}
