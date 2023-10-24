package resourcefieldexport

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/util/json"

	gdpv1alpha1 "github.com/deliveryhero/field-exporter/api/v1alpha1"
)

func verifyStatusConditions(ctx context.Context, objectMap map[string]any, requiredStatusConditions []gdpv1alpha1.StatusCondition) error {
	if len(requiredStatusConditions) == 0 {
		return nil
	}
	conditions, err := statusConditions(ctx, objectMap)
	if err != nil {
		return err
	}
	conditionByType := make(map[string]string)
	for _, c := range conditions {
		conditionByType[c.Type] = c.Status
	}
	var conditionErrors []error
	for _, condition := range requiredStatusConditions {
		var (
			value string
			ok    bool
		)
		if value, ok = conditionByType[condition.Type]; !ok {
			conditionErrors = append(conditionErrors, fmt.Errorf("status condition %s is not present", condition.Type))
			continue
		}
		if condition.Status != value {
			conditionErrors = append(conditionErrors, fmt.Errorf("status condition %s has value %s, expected %s", condition.Type, value, condition.Status))
		}
	}
	return errors.Join(conditionErrors...)
}

func statusConditions(ctx context.Context, input map[string]any) ([]knownCondition, error) {
	conditions, err := fieldValues(ctx, input, ".status.conditions")
	if err != nil {
		return nil, fmt.Errorf("failed to get status conditions: %w", err)
	}
	serializedConditions, err := json.Marshal(conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize .status.conditions")
	}

	var knownConditions []knownCondition
	err = json.Unmarshal(serializedConditions, &knownConditions)
	if err != nil {
		return nil, fmt.Errorf("status conditions have an unexpected format: %w", err)
	}
	return knownConditions, nil
}

type knownCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}
