package resourcefieldexport

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFieldValues(t *testing.T) {
	for _, tc := range []struct {
		name         string
		input        map[string]any
		query        string
		expectResult any
		expectErr    string
	}{
		{
			name:         "valid query",
			input:        map[string]any{"status": map[string]any{"ip": "127.0.0.1"}},
			query:        ".status.ip",
			expectResult: "127.0.0.1",
		},
		{
			name:      "invalid query",
			input:     map[string]any{"status": map[string]any{"ip": "127.0.0.1"}},
			query:     "-!x",
			expectErr: `invalid query "-!x": unexpected token "!"`,
		},
		{
			name:         "multi-tier result",
			input:        map[string]any{"status": map[string]any{"ip": "127.0.0.1"}},
			query:        ".status",
			expectResult: map[string]any{"ip": "127.0.0.1"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fieldValues(tc.input, tc.query)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectResult, result)
		})
	}
}
