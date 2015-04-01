package functions

import (
	"errors"
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"strings"
)

func Parse(context *aggregations.Context, input interface{}) (interface{}, error) {
	switch input.(type) {
	case map[string]interface{}:
		inputMap := input.(map[string]interface{})

		first := true

		for index, value := range inputMap {
			if strings.HasPrefix(index, "$") {
				if first {
					first = false
				} else {
					return nil, errors.New("Function calls must contain a single property.")
				}

				if handler, ok := functionHandlers[index]; ok {
					if value, err := Parse(context, value); err == nil {
						return handler(context, value)
					} else {
						return nil, err
					}
				} else {
					return nil, errors.New("Function " + index + " not found.")
				}
			} else {
				if v, err := Parse(context, value); err == nil {
					inputMap[index] = v
				} else {
					return nil, err
				}
			}
		}

		return inputMap, nil

	case []interface{}:
		inputArray := input.([]interface{})

		for index, value := range inputArray {
			if v, err := Parse(context, value); err == nil {
				inputArray[index] = v
			} else {
				return nil, err
			}
		}

		return inputArray, nil
	}

	return input, nil
}
