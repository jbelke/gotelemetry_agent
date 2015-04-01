package functions

import (
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"time"
)

func pushHandler(context *aggregations.Context, input interface{}) (interface{}, error) {
	if err := validatePayload("$push", input); err != nil {
		return nil, err
	}

	data := input.(map[string]interface{})

	seriesName := data["$series"].(string)
	value := data["$value"].(float64)

	var when *time.Time

	if w, ok := data["$when"]; ok {
		*when = time.Unix(w.(int64), 0)
	}

	series, err := aggregations.GetSeries(context, seriesName)

	if err != nil {
		return nil, err
	}

	series.Push(when, value)

	return nil, nil
}
