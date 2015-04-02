package functions

import (
	"errors"
	"fmt"
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas"
)

func init() {
	schemas.LoadSchema("pick")
	functionHandlers["$pick"] = pickHandler
}

func pickHandler(context *aggregations.Context, input interface{}) (interface{}, error) {
	if err := validatePayload("$pick", input); err != nil {
		return nil, err
	}

	data := input.(map[string]interface{})
	prop := data["prop"].(string)

	def, hasDefault := data["default"]

	if obj, ok := data["from"].(map[string]interface{}); ok {
		if val, ok := obj[prop]; ok {
			return val, nil
		}

		if hasDefault {
			return def, nil
		}

		return nil, errors.New(fmt.Sprintf("Property %s not found in %#v, and no default value defined", prop, obj))
	}

	if obj, ok := data["from"].([]interface{}); ok {
		result := []interface{}{}

		for index, rec := range obj {
			if record, ok := rec.(map[string]interface{}); ok {
				if val, ok := record[prop]; ok {
					result = append(result, val)
				} else if hasDefault {
					result = append(result, def)
				} else {
					return nil, errors.New(fmt.Sprintf("Property %s not found in %#v at index %d, and no default value defined", prop, obj, index))
				}
			} else {
				return nil, errors.New(fmt.Sprintf("$pick cannot handle the value %#v at index %d of %#v, and no default value defined", rec, index, obj))
			}
		}

		return result, nil
	}

	return nil, errors.New(fmt.Sprintf("$pick doesn't know how to handle %#v", data["from"]))
}
