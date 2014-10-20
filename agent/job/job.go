// Package job provides an interface to a Telemetry Agent job. A job represents the smallest
// unit of work that the agent recognizes—the specific of what a job does are left to the
// individual plugin
package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/telemetryapp/gotelemetry"
	"log"
)

type Job struct {
	ID           string                   // The ID of the job
	credentials  gotelemetry.Credentials  // The credentials used by the job. These are not exposed to the plugin
	stream       *gotelemetry.BatchStream // The batch stream used by the job. This is likewide not exposed to the plugin
	instance     PluginInstance           // The plugin instance
	errorChannel *chan error              // A channel to which all errors are funneled
}

// newJob creates and starts a new Job
func newJob(credentials gotelemetry.Credentials, stream *gotelemetry.BatchStream, id string, config map[string]interface{}, instance PluginInstance, errorChannel *chan error) (*Job, error) {
	result := &Job{
		ID:           id,
		credentials:  credentials,
		stream:       stream,
		instance:     instance,
		errorChannel: errorChannel,
	}

	go result.start(config)

	return result, nil
}

// start starts a job. It must be executed asychronously in its own goroutine
func (j *Job) start(config map[string]interface{}) {
	err := j.instance.Init(j, config)

	if err != nil {
		println("Error initializing " + j.ID)
		j.ReportError(err)
		return

		//TODO Signal failure to manager
	}

	j.instance.Run(j)
}

// GetOrCreateBoard either creates a board based on an exported template, or retrieves it
// if a board with the same name already exists.
//
// The template must be passed in JSON format as a string (you can use gotelemetry/boarddump to
// generate a template based on an existing board).
func (j *Job) GetOrCreateBoard(name, prefix string, templateSource string) (*gotelemetry.Board, error) {
	template := &gotelemetry.ExportedBoard{}

	err := json.Unmarshal([]byte(templateSource), &template)

	if err != nil {
		return nil, err
	}

	return gotelemetry.ImportBoard(j.credentials, name, prefix, template)
}

// CreateFlow creates a new flow.
func (j *Job) CreateFlow(tag string, variant, sourceProvider, filter, params string) (*gotelemetry.Flow, error) {
	return gotelemetry.NewFlowWithLayout(j.credentials, tag, variant, sourceProvider, filter, params)
}

// GetFlowLayout returns the layout of a given flow
func (j *Job) GetFlowLayout(id string) (*gotelemetry.Flow, error) {
	return gotelemetry.GetFlowLayout(j.credentials, id)
}

// ReadFlow populates a flow struct with the data that is currently on the server
// Note that it is not necessary to populate f.Data, as the method will automatically
// initialize a nil value with the appropriate data structure for the flow's variant.
func (j *Job) ReadFlow(f *gotelemetry.Flow) error {
	return f.Read(j.credentials)
}

// PostFlowUpdate queues a flow update. The method returns immediately, but the update
// will most likely be sent to the Telemetry API at a later point based on the configuration
// of the underlying stream
func (j *Job) PostFlowUpdate(flow *gotelemetry.Flow) {
	j.stream.C <- flow
}

// ReportError sends a formatted error to the agent's global error log. This should be
// a plugin's preferred error reporting method when running.
func (j *Job) ReportError(err error) {
	actualError := errors.New(j.ID + ": -> " + err.Error())

	if j.errorChannel != nil {
		*j.errorChannel <- actualError
	}
}

// Log sends data to the agent's global log. It works like log.Log
func (j *Job) Log(v ...interface{}) {
	log.Printf("%s -> %s", j.ID, fmt.Sprint(v))
}

// Log sends a formatted string to the agent's global log. It works like log.Logf
func (j *Job) Logf(format string, v interface{}) {
	log.Printf("%s -> %s", j.ID, fmt.Sprintf(format, v))
}
