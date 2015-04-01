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
	obj := data["from"].(map[string]interface{})

	if val, ok := obj[prop]; ok {
		return val, nil
	}

	return nil, errors.New(fmt.Sprintf("Property %s not found in %#v", prop, obj))
}
