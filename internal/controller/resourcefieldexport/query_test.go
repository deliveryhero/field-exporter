package resourcefieldexport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
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
		{
			name:      "no results error",
			input:     map[string]any{"status": map[string]any{"ip": "127.0.0.1"}},
			query:     ".spec",
			expectErr: "no results returned for query .spec",
		},
		{
			name: "multiple results error",
			input: map[string]any{
				"status": map[string]any{
					"pods": []any{
						map[string]any{
							"ip": "first",
						},
						map[string]any{
							"ip": "second",
						},
					},
				},
			},
			query:     ".status.pods[].ip",
			expectErr: `query .status.pods[].ip returned more than one result: [first second]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fieldValues(context.Background(), tc.input, tc.query)
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

func TestFieldValue(t *testing.T) {
	for _, tc := range []struct {
		name         string
		input        map[string]any
		query        string
		expectResult string
		expectErr    string
	}{
		{
			name:         "string value",
			input:        map[string]any{"status": map[string]any{"ip": "localhost"}},
			query:        ".status.ip",
			expectResult: "localhost",
		},
		{
			name:         "int value",
			input:        map[string]any{"status": map[string]any{"replicas": 8}},
			query:        ".status.replicas",
			expectResult: "8",
		},
		{
			name:         "int value",
			input:        map[string]any{"status": map[string]any{"available": true}},
			query:        ".status.available",
			expectResult: "true",
		},
		{
			name:      "error case",
			input:     map[string]any{"status": map[string]any{"available": true}},
			query:     ".status",
			expectErr: `unsupported data type map[string]interface {} for query .status`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fieldStringValue(context.Background(), tc.input, tc.query)
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
