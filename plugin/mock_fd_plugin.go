package plugin

import (
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
	"math/rand"
	"time"
)

type MockFdPlugin struct {
	*job.PluginHelper
}

func init() {
	job.RegisterPlugin("plugin_fd", MockFdPluginFactory)
}

func MockFdPluginFactory() job.PluginInstance {
	return &MockFdPlugin{
		job.NewPluginHelper(),
	}
}

func (plugin *MockFdPlugin) Init(job *job.Job) error {
	config := job.Config()

	configBoard := config["board"].(map[interface{}]interface{})
	boardName := configBoard["name"].(string)
	boardPrefix := configBoard["prefix"].(string)

	template := "{\"name\":\"StandardBoardTest\",\"theme\":\"dark\",\"aspect_ratio\":\"HDTV\",\"font_family\":\"normal\",\"font_size\":\"normal\",\"widget_background\":\"\",\"widget_margins\":3,\"widget_padding\":8,\"widgets\":[{\"flow\":{\"tag\":\"upstatus_5\",\"data\":{\"up\":[\"Server1\",\"Server2\"]}},\"variant\":\"upstatus\",\"column\":0,\"row\":13,\"width\":5,\"height\":7,\"in_board_index\":0,\"background\":\"default\"},{\"flow\":{\"tag\":\"graph_15\",\"data\":{\"renderer\":\"line\",\"series\":[{\"values\":[15,20,88,12,45,33,55,75,50,53]}]}},\"variant\":\"graph\",\"column\":23,\"row\":13,\"width\":9,\"height\":7,\"in_board_index\":1,\"background\":\"default\"},{\"flow\":{\"tag\":\"graph_13\",\"data\":{\"renderer\":\"line\",\"series\":[{\"values\":[79,44,9,60,21,69,28,24,73,30]}]}},\"variant\":\"graph\",\"column\":5,\"row\":13,\"width\":9,\"height\":7,\"in_board_index\":2,\"background\":\"default\"},{\"flow\":{\"tag\":\"graph_14\",\"data\":{\"renderer\":\"line\",\"series\":[{\"values\":[15,20,88,12,45,33,55,75,50,53]}]}},\"variant\":\"graph\",\"column\":14,\"row\":13,\"width\":9,\"height\":7,\"in_board_index\":3,\"background\":\"default\"},{\"flow\":{\"tag\":\"text_3\",\"data\":{\"alignment\":\"center\",\"text\":\"Message 1\"}},\"variant\":\"text\",\"column\":0,\"row\":7,\"width\":32,\"height\":3,\"in_board_index\":4,\"background\":\"default\"},{\"flow\":{\"tag\":\"image_13\",\"data\":{\"mode\":\"fit\",\"url\":\"https://www.telemetryapp.com/img/telemetry-logo-header.png\"}},\"variant\":\"image\",\"column\":0,\"row\":0,\"width\":6,\"height\":7,\"in_board_index\":5,\"background\":\"default\"},{\"flow\":{\"tag\":\"gauge_22\",\"data\":{\"max\":100,\"value\":50}},\"variant\":\"gauge\",\"column\":26,\"row\":0,\"width\":6,\"height\":7,\"in_board_index\":6,\"background\":\"default\"},{\"flow\":{\"tag\":\"timeseries_10\",\"data\":{\"interval\":\"hours\",\"interval_count\":0,\"renderer\":\"line\",\"series_metadata\":[{\"aggregation\":\"avg\"}],\"values\":[100,200,300]}},\"variant\":\"timeseries\",\"column\":16,\"row\":0,\"width\":10,\"height\":7,\"in_board_index\":7,\"background\":\"default\"},{\"flow\":{\"tag\":\"timeseries_9\",\"data\":{\"interval\":\"hours\",\"interval_count\":0,\"renderer\":\"line\",\"series_metadata\":[{\"aggregation\":\"avg\"}],\"values\":[100,200,300]}},\"variant\":\"timeseries\",\"column\":6,\"row\":0,\"width\":10,\"height\":7,\"in_board_index\":8,\"background\":\"default\"},{\"flow\":{\"tag\":\"value_50\",\"data\":{\"value\":100}},\"variant\":\"value\",\"column\":0,\"row\":9,\"width\":11,\"height\":4,\"in_board_index\":9,\"background\":\"default\"},{\"flow\":{\"tag\":\"value_51\",\"data\":{\"value\":100}},\"variant\":\"value\",\"column\":11,\"row\":9,\"width\":10,\"height\":4,\"in_board_index\":10,\"background\":\"default\"},{\"flow\":{\"tag\":\"value_52\",\"data\":{\"value\":100}},\"variant\":\"value\",\"column\":21,\"row\":9,\"width\":11,\"height\":4,\"in_board_index\":11,\"background\":\"default\"}]}"

	board, err := job.GetOrCreateBoard(boardName, boardPrefix, template)

	if err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedValueFlow, time.Second*5, board, "value_50"); err != nil {
		return err
	}
	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedValueFlow, time.Second*5, board, "value_51"); err != nil {
		return err
	}
	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedValueFlow, time.Second*5, board, "value_52"); err != nil {
		return err
	}

	return nil

}

// FIXME: Using Random from randomPlugin as a place holder. Expect to change this for s SQL Query to fetch real value from PRD
func (plugin *MockFdPlugin) FeedValueFlow(job *job.Job, flow *gotelemetry.Flow) {
	data, err := flow.ValueData()

	if !err {
		// Report an error this way to ensure proper logging
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	data.Value = rand.Float64() * 10000

	// Post a flow update this way. Updates are automatically accumulated and submitted to
	// Telemetry according to the schedule illustrated in the config
	job.PostFlowUpdate(flow)

	// Log to the global log this way.
	job.Logf("Updated flow %s", flow.Tag)
}
