package plugin

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type MockFdPlugin struct {
	*job.PluginHelper
}

var db, dbErr = sql.Open("mysql", "root@/FD_dump")

func init() {
	job.RegisterPlugin("plugin_fd", MockFdPluginFactory)

	if dbErr != nil {
		log.Fatal(gotelemetry.NewError(500, "Plugin 'plugin_fd' could not connect to the database"))
	}
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

	template :=	"{\"name\":\"StandarBoardTest\",\"theme\":\"dark\",\"aspect_ratio\":\"HDTV\",\"font_family\":\"normal\",\"font_size\":\"normal\",\"widget_background\":\"\",\"widget_margins\":3,\"widget_padding\":8,\"widgets\":[{\"flow\":{\"tag\":\"upstatus_5\",\"data\":{\"up\":[\"Server1\",\"Server2\"]}},\"variant\":\"upstatus\",\"column\":0,\"row\":13,\"width\":5,\"height\":7,\"in_board_index\":0,\"background\":\"default\"},{\"flow\":{\"tag\":\"graph_15\",\"data\":{\"renderer\":\"line\",\"series\":[{\"values\":[15,20,88,12,45,33,55,75,50,53]}]}},\"variant\":\"graph\",\"column\":23,\"row\":13,\"width\":9,\"height\":7,\"in_board_index\":1,\"background\":\"default\"},{\"flow\":{\"tag\":\"graph_13\",\"data\":{\"renderer\":\"line\",\"series\":[{\"values\":[9,18,18,18,18,18,19,18,18,20,18,17,17,18,18,19,18,19,20,18,18,18,18,18,18,18,18,18,18,18]}],\"title\":\"30 Day Transaction Volume\"}},\"variant\":\"graph\",\"column\":5,\"row\":13,\"width\":9,\"height\":7,\"in_board_index\":2,\"background\":\"default\"},{\"flow\":{\"tag\":\"graph_14\",\"data\":{\"renderer\":\"line\",\"series\":[{\"values\":[8,7,9,18,18,18,18,18,19,18,18,20,18,17,17,18,18,19,18,19,20,18,18,18,18,18,18,18,18,18,18,18,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}],\"title\":\"90 Day Transaction Volume\"}},\"variant\":\"graph\",\"column\":14,\"row\":13,\"width\":9,\"height\":7,\"in_board_index\":3,\"background\":\"default\"},{\"flow\":{\"tag\":\"text_3\",\"data\":{\"alignment\":\"center\",\"text\":\"Daily transaction volume - Today is: Oct 10, 301012\"}},\"variant\":\"text\",\"column\":0,\"row\":7,\"width\":32,\"height\":3,\"in_board_index\":4,\"background\":\"default\"},{\"flow\":{\"tag\":\"image_13\",\"data\":{\"mode\":\"fit\",\"url\":\"https://www.telemetryapp.com/img/telemetry-logo-header.png\"}},\"variant\":\"image\",\"column\":0,\"row\":0,\"width\":6,\"height\":7,\"in_board_index\":5,\"background\":\"default\"},{\"flow\":{\"tag\":\"gauge_22\",\"data\":{\"max\":100,\"value\":50}},\"variant\":\"gauge\",\"column\":26,\"row\":0,\"width\":6,\"height\":7,\"in_board_index\":6,\"background\":\"default\"},{\"flow\":{\"tag\":\"value_50\",\"data\":{\"label\":\"Approved\",\"value\":23163}},\"variant\":\"value\",\"column\":0,\"row\":9,\"width\":11,\"height\":4,\"in_board_index\":7,\"background\":\"default\"},{\"flow\":{\"tag\":\"value_51\",\"data\":{\"label\":\"Declined\",\"value\":6287}},\"variant\":\"value\",\"column\":11,\"row\":9,\"width\":10,\"height\":4,\"in_board_index\":8,\"background\":\"default\"},{\"flow\":{\"tag\":\"value_52\",\"data\":{\"label\":\"Error\",\"value\":1050}},\"variant\":\"value\",\"column\":21,\"row\":9,\"width\":11,\"height\":4,\"in_board_index\":9,\"background\":\"default\"},{\"flow\":{\"tag\":\"timeseries_9\",\"data\":{\"interval\":\"hours\",\"interval_count\":0,\"renderer\":\"line\",\"series_metadata\":[{\"aggregation\":\"avg\"},{\"aggregation\":\"avg\"},{\"aggregation\":\"avg\"}],\"values\":[100,200,300]}},\"variant\":\"timeseries\",\"column\":6,\"row\":0,\"width\":10,\"height\":7,\"in_board_index\":10,\"background\":\"default\"},{\"flow\":{\"tag\":\"timeseries_10\",\"data\":{\"interval\":\"hours\",\"interval_count\":0,\"renderer\":\"line\",\"series_metadata\":[{\"aggregation\":\"avg\"},{\"aggregation\":\"avg\"},{\"aggregation\":\"avg\"}],\"values\":[100,200,300]}},\"variant\":\"timeseries\",\"column\":16,\"row\":0,\"width\":10,\"height\":7,\"in_board_index\":11,\"background\":\"default\"}]}"
	
	board, err := job.GetOrCreateBoard(boardName, boardPrefix, template)

	if err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedImageFlow, time.Second*5, board, "image_13"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedTimeSeriesFlowAPITransactionsPerMinute, time.Second*5, board, "timeseries_9"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedTimeSeriesFlowHCOTransactionsPerMinute, time.Second*5, board, "timeseries_10"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedGaugeFlow, time.Second*5, board, "gauge_22"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedMessageFlow, time.Second*5, board, "text_3"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedValueFlowForApprovedTransactions, time.Second*5, board, "value_50"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedValueFlowForDeclinedTransactions, time.Second*5, board, "value_51"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedValueFlowForErrorTransactions, time.Second*5, board, "value_52"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedUpStatusFlow, time.Second*5, board, "upstatus_5"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.Feed30DayTransactionVolumeGraphFlow, time.Second*5, board, "graph_13"); err != nil {
		return err
	}

	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.Feed90DayTransactionVolumeGraphFlow, time.Second*5, board, "graph_14"); err != nil {
		return err
	}
	if err = plugin.PluginHelper.AddTaskWithClosureFromBoardForFlowWithTag(plugin.FeedGraphFlow, time.Second*5, board, "graph_15"); err != nil {
		return err
	}

	return nil

}

func (plugin *MockFdPlugin) FeedImageFlow(job *job.Job, flow *gotelemetry.Flow) {
	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) FeedTimeSeriesFlowAPITransactionsPerMinute(job *job.Job, flow *gotelemetry.Flow) {

	data, err := flow.TimeseriesData()
	if !err {
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	data.Title = "API transactions per minute"
	data.Interval = "minutes"
	data.IntervalCount = 60

	data.SeriesMetadata[0].Label = "Approved"
	data.SeriesMetadata[0].Color = "8de2x26"

	data.SeriesMetadata[1].Label = "Declined"
	data.SeriesMetadata[1].Color = "2aa4e5"

	data.SeriesMetadata[2].Label = "Error"
	data.SeriesMetadata[2].Color = "ff543b"

	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) FeedTimeSeriesFlowHCOTransactionsPerMinute(job *job.Job, flow *gotelemetry.Flow) {

	data, err := flow.TimeseriesData()
	if !err {
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	data.Title = "HCO transactions per minute"
	data.Interval = "minutes"
	data.IntervalCount = 60

	data.SeriesMetadata[0].Label = "Approved"
	data.SeriesMetadata[0].Color = "8de226"

	data.SeriesMetadata[1].Label = "Declined"
	data.SeriesMetadata[1].Color = "2aa4e5"

	data.SeriesMetadata[2].Label = "Error"
	data.SeriesMetadata[2].Color = "ff543b"

	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) FeedGaugeFlow(job *job.Job, flow *gotelemetry.Flow) {
	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) FeedMessageFlow(job *job.Job, flow *gotelemetry.Flow) {

	data, err := flow.TextData()

	if !err {
		// Report an error this way to ensure proper logging
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	data.Text = fmt.Sprintf("Daily transaction volume - Today is: %s", time.Now().Format("Jan 1, 2014"))

	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) FeedValueFlowForApprovedTransactions(job *job.Job, flow *gotelemetry.Flow) {
	data, err := flow.ValueData()

	if !err {
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	data.Value = data.Value.(float64) + float64(rand.Intn(100))
	data.Label = "Approved"

	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) FeedValueFlowForDeclinedTransactions(job *job.Job, flow *gotelemetry.Flow) {
	data, err := flow.ValueData()

	if !err {
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	data.Value = data.Value.(float64) + float64(rand.Intn(30))
	data.Label = "Declined"

	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) FeedValueFlowForErrorTransactions(job *job.Job, flow *gotelemetry.Flow) {
	data, err := flow.ValueData()

	if !err {
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	data.Value = data.Value.(float64) + float64(rand.Intn(5))
	data.Label = "Error"

	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) FeedUpStatusFlow(job *job.Job, flow *gotelemetry.Flow) {

	data, err := flow.UpstatusData()

	if !err {
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	if rand.Intn(10)%10 == 0 {
		data.Down = []string{"I'm down for some reason :("}
	} else {
		data.Down = []string{}
	}

	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) Feed30DayTransactionVolumeGraphFlow(job *job.Job, flow *gotelemetry.Flow) {

	data, err := flow.GraphData()

	if !err {
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	data.Title = "30 Day Transaction Volume"
	data.Series[0].Values = query30DayTransactionVolume()

	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) Feed90DayTransactionVolumeGraphFlow(job *job.Job, flow *gotelemetry.Flow) {

	data, err := flow.GraphData()

	if !err {
		job.ReportError(gotelemetry.NewError(500, "Cannot extract value data from flow with tag `"+flow.Tag+"`"))
		return
	}

	data.Title = "90 Day Transaction Volume"
	data.Series[0].Values = query90DayTransactionVolume()

	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

func (plugin *MockFdPlugin) FeedGraphFlow(job *job.Job, flow *gotelemetry.Flow) {
	job.PostFlowUpdate(flow)
	job.Logf("Updated flow %s", flow.Tag)
}

// Util methods to query
func query30DayTransactionVolume() []float64 {
	return queryNDayTransactionVolume(30)
}

func query90DayTransactionVolume() []float64 {
	return queryNDayTransactionVolume(90)
}

func queryNDayTransactionVolume(interval int) []float64 {

	transactionCountArray := make([]float64, interval)

	query := "select date, sum(approved_count) from transaction_summaries where terminal_id in (select t.terminal_id from exactterminals t, exactmerchants m where  t.merchant_id = m.merchant_id and m.merchant_id = 1) and date between DATE('2012-07-20') - INTERVAL " + strconv.Itoa(interval-1) + " DAY and DATE('2012-07-20') group by date"
	rows, err := db.Query(query)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		var date string
		var count float64
		rows.Scan(&date, &count)
		transactionCountArray[i] = count
	}

	return transactionCountArray
}
