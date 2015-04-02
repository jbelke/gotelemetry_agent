package functions

import (
	"errors"
	"fmt"
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas"
	"time"
)

func init() {
	schemas.LoadSchema("aggregate")
	functionHandlers["$aggregate"] = aggregateHandler
}

func aggregateHandler(context *aggregations.Context, input interface{}) (interface{}, error) {
	if err := validatePayload("$aggregate", input); err != nil {
		return nil, err
	}

	data := input.(map[string]interface{})

	var op aggregations.FunctionType

	switch data["op"].(string) {
	case "sum":
		op = aggregations.Sum

	case "avg":
		op = aggregations.Avg

	case "max":
		op = aggregations.Max

	case "min":
		op = aggregations.Min

	case "count":
		op = aggregations.Count

	default:
		return nil, errors.New(fmt.Sprintf("Unknown operation %s", data["op"].(string)))
	}

	seriesName := data["series"].(string)

	var start, end *time.Time

	if period, ok := data["period"]; ok {
		if times, ok := period.(map[string]interface{}); ok {
			if ts, ok := times["from"].(float64); ok {
				start = &time.Time{}
				*start = time.Unix(int64(ts), 0)
			}

			if ts, ok := times["to"].(float64); ok {
				end = &time.Time{}
				*end = time.Unix(int64(ts), 0)
			}
		} else if duration, ok := data["period"].(float64); ok {
			start = &time.Time{}
			*start = time.Now().Add(-time.Duration(duration) * time.Second)

			end = &time.Time{}
			*end = time.Now()
		}
	}

	series, err := aggregations.GetSeries(context, seriesName)

	if err != nil {
		return nil, err
	}

	return series.Aggregate(op, int(data["interval"].(float64)), int(data["count"].(float64)))
}
