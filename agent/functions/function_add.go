package functions

import (
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
)

func addHandler(context *aggregations.Context, input interface{}) (interface{}, error) {
	if err := validatePayload("$add", input); err != nil {
		return nil, err
	}

	data := input.(map[string]interface{})

	return data["$left"].(float64) + data["$right"].(float64), nil
}
