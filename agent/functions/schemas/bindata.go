package schemas

import (
	"fmt"
	"io/ioutil"
)

// bindata_read reads the given file from disk. It returns
// an error on failure.
func bindata_read(path, name string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset %s at %s: %v", name, path, err)
	}
	return buf, err
}


// json_add_json reads file data from disk.
// It panics if something went wrong in the process.
func json_add_json() ([]byte, error) {
	return bindata_read(
		"/Users/marcot/Sites/go/src/github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas/json/add.json",
		"json/add.json",
	)
}

// json_aggregate_json reads file data from disk.
// It panics if something went wrong in the process.
func json_aggregate_json() ([]byte, error) {
	return bindata_read(
		"/Users/marcot/Sites/go/src/github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas/json/aggregate.json",
		"json/aggregate.json",
	)
}

// json_compute_json reads file data from disk.
// It panics if something went wrong in the process.
func json_compute_json() ([]byte, error) {
	return bindata_read(
		"/Users/marcot/Sites/go/src/github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas/json/compute.json",
		"json/compute.json",
	)
}

// json_last_json reads file data from disk.
// It panics if something went wrong in the process.
func json_last_json() ([]byte, error) {
	return bindata_read(
		"/Users/marcot/Sites/go/src/github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas/json/last.json",
		"json/last.json",
	)
}

// json_pick_json reads file data from disk.
// It panics if something went wrong in the process.
func json_pick_json() ([]byte, error) {
	return bindata_read(
		"/Users/marcot/Sites/go/src/github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas/json/pick.json",
		"json/pick.json",
	)
}

// json_pop_json reads file data from disk.
// It panics if something went wrong in the process.
func json_pop_json() ([]byte, error) {
	return bindata_read(
		"/Users/marcot/Sites/go/src/github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas/json/pop.json",
		"json/pop.json",
	)
}

// json_push_json reads file data from disk.
// It panics if something went wrong in the process.
func json_push_json() ([]byte, error) {
	return bindata_read(
		"/Users/marcot/Sites/go/src/github.com/telemetryapp/gotelemetry_agent/agent/functions/schemas/json/push.json",
		"json/push.json",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	if f, ok := _bindata[name]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string] func() ([]byte, error) {
	"json/add.json": json_add_json,
	"json/aggregate.json": json_aggregate_json,
	"json/compute.json": json_compute_json,
	"json/last.json": json_last_json,
	"json/pick.json": json_pick_json,
	"json/pop.json": json_pop_json,
	"json/push.json": json_push_json,

}
