package functions

import (
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas"
	"io"
)

func init() {
	schemas.LoadSchema("last")
	functionHandlers["$last"] = lastHandler
}

func lastHandler(context *aggregations.Context, input interface{}) (interface{}, error) {
	if err := validatePayload("$last", input); err != nil {
		return nil, err
	}

	data := input.(map[string]interface{})

	seriesName := data["series"].(string)

	series, err := aggregations.GetSeries(context, seriesName)

	if err != nil {
		return nil, err
	}

	result, err := series.Last()

	if err != nil {
		if err == io.EOF {
			if defaultValue, ok := data["default"]; ok {
				return defaultValue, nil
			} else {
				return map[string]interface{}{}, nil
			}
		}

		return nil, err
	}

	return result, nil
}
