package functions

import (
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas"
	"io"
)

func init() {
	schemas.LoadSchema("pop")
	functionHandlers["$pop"] = popHandler
}

func popHandler(context *aggregations.Context, input interface{}) (interface{}, error) {
	if err := validatePayload("$pop", input); err != nil {
		return nil, err
	}

	data := input.(map[string]interface{})

	seriesName := data["$series"].(string)

	condition, ok := data["$condition"].(bool)

	if !ok {
		condition = true
	}

	series, err := aggregations.GetSeries(context, seriesName)

	if err != nil {
		return nil, err
	}

	result, err := series.Pop(condition)

	if err != nil {
		if err == io.EOF {
			return map[string]interface{}{}, nil
		}

		return nil, err
	}

	return result, nil
}
