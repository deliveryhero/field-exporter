package resourcefieldexport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/util/json"

	gdpv1alpha1 "github.com/deliveryhero/field-exporter/api/v1alpha1"
)

func TestStatusConditions(t *testing.T) {
	for _, tc := range []struct {
		name      string
		input     string
		expected  []knownCondition
		expectErr string
	}{
		{
			name:      "status not present",
			input:     `{}`,
			expectErr: `failed to get status conditions: no results returned for query .status.conditions`,
		},
		{
			name:      "conditions not present",
			input:     `{"status": {}}`,
			expectErr: `failed to get status conditions: no results returned for query .status.conditions`,
		},
		{
			name:     "conditions empty",
			input:    `{"status": {"conditions": [] }}`,
			expected: []knownCondition{},
		},
		{
			name:  "one condition",
			input: `{"status": {"conditions": [{"type": "Ready", "status": "True", "message": "Resource is ready"}] }}`,
			expected: []knownCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
		{
			name: "multiple conditions",
			input: `{
				"status": {
					"conditions": [
						{"type": "Ready", "status": "True"},
						{"type": "Available", "status": "False"},
						{"type": "CanScale", "status": "True"}
					]
				}
			}`,
			expected: []knownCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
				{
					Type:   "Available",
					Status: "False",
				},
				{
					Type:   "CanScale",
					Status: "True",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := make(map[string]any)
			err := json.Unmarshal([]byte(tc.input), &input)
			require.NoError(t, err)
			conditions, err := statusConditions(context.Background(), input)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, conditions)
		})
	}
}

func TestVerifyStatusConditions(t *testing.T) {
	for _, tc := range []struct {
		name               string
		inputConditions    string
		requiredConditions []gdpv1alpha1.StatusCondition
		expectErr          string
	}{
		{
			name:            "no required conditions",
			inputConditions: `{}`,
		},
		{
			name:            "conditions empty",
			inputConditions: `{"status":{"conditions":[]}}`,
			requiredConditions: []gdpv1alpha1.StatusCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
			expectErr: "status condition Ready is not present",
		},
		{
			name:            "conditions match",
			inputConditions: `{"status":{"conditions":[{"type":"Ready", "status":"True"}]}}`,
			requiredConditions: []gdpv1alpha1.StatusCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
		{
			name:            "conditions mismatch",
			inputConditions: `{"status":{"conditions":[{"type":"Ready", "status":"True"}]}}`,
			requiredConditions: []gdpv1alpha1.StatusCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := make(map[string]any)
			require.NoError(t, json.Unmarshal([]byte(tc.inputConditions), &input))
			err := verifyStatusConditions(context.Background(), input, tc.requiredConditions)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
