package plugin

import (
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
	"math/rand"
	"time"
)

// init() registers this plugin with the Plugin Manager.
// The plugin provides a RandomPluginFactory that the manager calls whenever it needs
// to create a new job
func init() {
	job.RegisterPlugin("random", &RandomPluginFactory{})
}

//Struct RandomPluginFactory implements the factory methods required to
//create new instances of a plugin.
type RandomPluginFactory struct {
}

// NewInstance generates a blank plugin instance
func (r *RandomPluginFactory) NewInstance() job.PluginInstance {
	return &RandomPlugin{
		job.NewPluginHelper(),
	}
}

// Struct RandomPlugin is our sample plugin. It takes a config that provides the name and prefix of a board and
// attempts to create that board inside the account; it then feeds random data to the board on a schedule.
type RandomPlugin struct {
	*job.PluginHelper
}

// Init initializes the plugin. Perform whatever initialization you need here.
//
// Note that this method should be considered synchronous within the job's context—that is,
// once Init() returns, Run() is called right away. Of course, this doesn't prevent you
// from spawning goroutines inside Init() if you need to.
func (r *RandomPlugin) Init(job *job.Job, config map[string]interface{}) error {
	// Look for the board spec in our config
	board := config["board"].(map[interface{}]interface{})

	boardName := board["name"].(string)
	boardPrefix := board["prefix"].(string)

	// This template was generated using gotelemetry/dumpboard
	template := "{\"name\":\"RandomTest\",\"theme\":\"dark\",\"aspect_ratio\":\"HDTV\",\"font_family\":\"normal\",\"font_size\":\"normal\",\"widget_background\":\"\",\"widget_margins\":3,\"widget_padding\":8,\"widgets\":[{\"variant\":\"value\",\"tag\":\"value_98\",\"column\":7,\"row\":7,\"width\":8,\"height\":5,\"in_board_index\":0,\"background\":\"default\"},{\"variant\":\"value\",\"tag\":\"value_99\",\"column\":17,\"row\":7,\"width\":8,\"height\":5,\"in_board_index\":1,\"background\":\"default\"}]}"

	// Import the board (or, if it already exists, retrieve it)
	b, err := job.GetOrCreateBoard(boardName, boardPrefix, template)

	if err != nil {
		return err
	}

	// Map the board's widgets to their respective flows and pre-populate
	// the flows with their current values
	flows, err := b.MapWidgetsToFlows()

	if err != nil {
		return err
	}

	if err = r.PluginHelper.AddTaskWithClosureForFlowWithName(FeedValueFlow, time.Second*5, flows, "value_98", boardPrefix); err != nil {
		return err
	}

	if err = r.PluginHelper.AddTaskWithClosureForFlowWithName(FeedValueFlow, time.Second*7, flows, "value_99", boardPrefix); err != nil {
		return err
	}

	return nil
}

func FeedValueFlow(job *job.Job, f *gotelemetry.Flow) {
	data, err := f.ValueData()

	if !err {
		// Report an error this way to ensure proper logging
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+f.Tag+"`"))
		return
	}

	data.Value = rand.Float64() * 10000

	// Post a flow update this way. Updates are automatically accumulated and submitted to
	// Telemetry according to the schedule illustrated in the config
	job.PostFlowUpdate(f)

	// Log to the global log this way.
	job.Logf("Updated flow %s", f.Tag)
}
