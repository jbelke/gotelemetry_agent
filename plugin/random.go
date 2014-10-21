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
	job.RegisterPlugin("random", RandomPluginFactory)
}

// Func RandomPluginFactory generates a blank plugin instance
func RandomPluginFactory() job.PluginInstance {
	return &RandomPlugin{
		job.NewPluginHelper(),
		map[string]interface{}{},
	}
}

// Struct RandomPlugin is our sample plugin. It takes a config that provides the name and prefix of a board and
// attempts to create that board inside the account; it then feeds random data to the board on a schedule.
//
// RandomPlugin makes use of job.PluginHelper, which handles most of the scaffolding necessary to run
// the plugin itself.
type RandomPlugin struct {
	*job.PluginHelper
	config map[string]interface{}
}

// Init initializes the plugin. Perform whatever initialization you need here.
//
// Note that this method should be considered synchronous within the job's context—that is,
// once Init() returns, Run() is called right away. Of course, this doesn't prevent you
// from spawning goroutines inside Init() if you need to.
//
// The functionality present in this method will vary depending on the type of plugin
// you're writing. In general, however, you will want to follow a three-phase approach:
//
// 1. Read, parse, and validate your configuration data.
// 2. Create any boards or flows you need.
// 3. If you are using a PluginHelper as the base for your plugin, you will probably
//    want to register your tasks at this point.
func (r *RandomPlugin) Init(job *job.Job, config map[string]interface{}) error {
	r.config = config

	// Phase 1: Read configuration

	// Look for the board spec in our config
	board := config["board"].(map[interface{}]interface{})

	boardName := board["name"].(string)
	boardPrefix := board["prefix"].(string)

	// Phase 2: Create or find a board that matches our config

	// This template was generated using gotelemetry/dumpboard
	template := "{\"name\":\"RandomTest\",\"theme\":\"dark\",\"aspect_ratio\":\"HDTV\",\"font_family\":\"normal\",\"font_size\":\"normal\",\"widget_background\":\"\",\"widget_margins\":3,\"widget_padding\":8,\"widgets\":[{\"variant\":\"value\",\"tag\":\"value_98\",\"column\":7,\"row\":7,\"width\":8,\"height\":5,\"in_board_index\":0,\"background\":\"default\"},{\"variant\":\"value\",\"tag\":\"value_99\",\"column\":17,\"row\":7,\"width\":8,\"height\":5,\"in_board_index\":1,\"background\":\"default\"}]}"

	// Import the board (or, if it already exists, retrieve it)
	b, err := job.GetOrCreateBoard(boardName, boardPrefix, template)

	if err != nil {
		return err
	}

	// Phase 3: Since we use a plugin helper, ask it to execute our update function on
	// a schedule until the plugin is terminated

	if err = r.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(r.FeedValueFlow, time.Second*5, b, "value_98"); err != nil {
		return err
	}

	if err = r.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(r.FeedValueFlow, time.Second*7, b, "value_99"); err != nil {
		return err
	}

	return nil
}

// FeedValueFlow is a generic function that sends a random number to a value flow.
//
// This is executed on a schedule by the plugin helper. In a “real” plugin, you
// could have a specialized function for each of your flow.
//
// PluginHelper tasks such as these are executed in a goroutine, which means that
// you can safely treat them as synchronous without fear of them blocking each other.
//
// Because your code runs in a thread, however, you must follow two important rules:
//
// 1. Never write two tasks that manipulate the same flow, or they will
//    inevitably overwrite each other
// 2. Avoid writing directly to common structs or global variables that
//    may be used by another task.
func (r *RandomPlugin) FeedValueFlow(job *job.Job, f *gotelemetry.Flow) {
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
