package expr_test

import (
	"testing"

	"github.com/antonmedv/expr"
	"github.com/stretchr/testify/require"
)

func TestNamespacedLookup(t *testing.T) {
	data := map[string]interface{}{
		"data:age":    18,
		"system:user": map[string]interface{}{"ID": 1},
		"var:items":   []interface{}{map[string]interface{}{"name": "A"}},
		"data:index":  0,
	}
	for _, tc := range []struct {
		code string
		want interface{}
	}{
		{"data:age >= 18", true},
		{"system:user.ID == 1", true},
		{"var:items[(data:index)].name", "A"},
		{"data:missing", nil},
	} {
		t.Run(tc.code, func(t *testing.T) {
			program, err := expr.Compile(tc.code)
			require.NoError(t, err)
			require.Equal(t, tc.code, program.Source.Content())
			got, err := expr.Run(program, data)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
