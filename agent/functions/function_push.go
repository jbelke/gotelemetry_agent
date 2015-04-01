package functions

import (
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas"
	"time"
)

func init() {
	schemas.LoadSchema("push")
	functionHandlers["$push"] = pushHandler
}

func pushHandler(context *aggregations.Context, input interface{}) (interface{}, error) {
	if err := validatePayload("$push", input); err != nil {
		return nil, err
	}

	data := input.(map[string]interface{})

	seriesName := data["$series"].(string)
	value := data["$value"].(float64)

	var ts *time.Time

	if w, ok := data["$ts"]; ok {
		*ts = time.Unix(w.(int64), 0)
	}

	series, err := aggregations.GetSeries(context, seriesName)

	if err != nil {
		return nil, err
	}

	err = series.Push(ts, value)

	return nil, err
}
