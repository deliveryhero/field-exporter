package resourcefieldexport

import (
	"context"
	"fmt"
	"github.com/itchyny/gojq"
)

func fieldValues(ctx context.Context, input map[string]interface{}, queryString string) (any, error) {
	query, err := gojq.Parse(queryString)
	if err != nil {
		return "", fmt.Errorf("invalid query %q: %w", queryString, err)
	}

	resultIter := query.RunWithContext(ctx, input)
	var results []any
	for {
		value, ok := resultIter.Next()
		if !ok {
			break
		}
		if err, ok := value.(error); ok {
			return "", err
		}
		if value == nil {
			continue
		}
		results = append(results, value)
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no results returned for query %s", queryString)
	}

	if len(results) != 1 {
		return "", fmt.Errorf("query %s returned more than one result: %v", queryString, results)
	}

	return results[0], nil
}

func fieldStringValue(ctx context.Context, input map[string]interface{}, query string) (string, error) {
	result, err := fieldValues(ctx, input, query)
	if err != nil {
		return "", err
	}

	switch x := result.(type) {
	case string:
		return x, nil
	case int:
		return fmt.Sprintf("%d", x), nil
	case bool:
		return fmt.Sprintf("%t", x), nil
	default:
		return "", fmt.Errorf("unsupported data type %T for query %s", result, query)
	}
}
