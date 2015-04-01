package functions

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas"
	"strings"
)

type functionHandler func(context *aggregations.Context, input interface{}) (interface{}, error)

var functionHandlers = map[string]functionHandler{
	"$add": addHandler,

	"$push": pushHandler,
}

func validatePayload(name string, payload interface{}) error {
	if schema, ok := schemas.Schemas[name]; ok {
		result := schema.Validate(payload)

		if result.Valid() {
			return nil
		} else {
			errorStrings := []string{}

			for _, err := range result.Errors() {
				errorStrings = append(errorStrings, err.String())
			}

			js, _ := json.Marshal(payload)

			return errors.New(fmt.Sprintf("In expression {%s: %s}: %s", name, string(js), strings.Join(errorStrings, " - ")))
		}
	} else {
		return errors.New("Unable to find a validator for the function `" + name + "`")
	}
}
