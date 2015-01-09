package plugin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/evanphx/json-patch"
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
	"strconv"
	"strings"
	"time"
)

// Function init() registers this plugin with the Plugin Manager.
// The plugin provides a SQLPluginFactory that the manager calls whenever it needs
// to create a new job
func init() {
	job.RegisterPlugin("com.telemetryapp.sql", SQLPluginFactory)
}

// Func IntercomPluginFactory generates a blank plugin instance
func SQLPluginFactory() job.PluginInstance {
	return &SQLPlugin{
		PluginHelper: job.NewPluginHelper(),
	}
}

// Struct SQLPlugin is allows populating flows based on the content of a SQL
// database
//
type SQLPlugin struct {
	*job.PluginHelper
	driverName     string
	datasourceName string
	query          string
	patch          string
	flowTag        string
	variant        string
	template       map[string]interface{}
	flow           *gotelemetry.Flow
}

// Function Init initializes the plugin.
//
// Required configuration parameters are:
func (p *SQLPlugin) Init(job *job.Job) error {
	var err error

	c := job.Config()

	p.driverName = c["driver"].(string)
	p.datasourceName = c["datasource"].(string)
	p.query = c["query"].(string)
	p.flowTag = c["flow_tag"].(string)
	p.variant = c["variant"].(string)
	p.template = config.MapFromYaml(c["template"]).(map[string]interface{})

	patch, err := json.Marshal(config.MapFromYaml(c["patch"]))

	if err != nil {
		job.ReportError(err)
		return err
	}

	p.patch = string(patch)

	p.flow, err = job.GetFlowTagLayout(p.flowTag)

	if err == nil {
		if p.flow.Variant != p.variant {
			return errors.New("Flow " + p.flow.Id + " is of type " + p.flow.Variant + " instead of the expected " + p.variant)
		}

		if err != nil {
			return err
		}
	} else {
		p.flow, err = job.CreateFlow(p.flowTag, p.variant, "gotelemetry_agent", "", "")

		if err != nil {
			return err
		}

		err = job.ReadFlow(p.flow)

		if err != nil {
			return err
		}

		err = p.flow.Populate(p.variant, p.template)

		if err != nil {
			return err
		}

		err = job.PostImmediateFlowUpdate(p.flow)

		if err != nil {
			return err
		}
	}

	if refresh, ok := c["refresh"]; ok {
		p.PluginHelper.AddTaskWithClosure(p.performAllTasks, time.Duration(refresh.(int))*time.Second)
	} else {
		p.PluginHelper.AddTaskWithClosure(p.performAllTasks, 0)
	}

	return nil
}

func (p *SQLPlugin) performAllTasks(j *job.Job) {
	j.Log("Starting SQL plugin...")

	db, err := sql.Open(p.driverName, p.datasourceName)

	if err != nil {
		j.ReportError(err)
		return
	}

	rs, err := db.Query(p.query)

	if err != nil {
		j.ReportError(err)
		return
	}

	defer rs.Close()

	j.ReadFlow(p.flow)

	doc, err := json.Marshal(p.flow.Data)

	if err != nil {
		j.ReportError(err)
		return
	}

	rowIndex := 0

	for rs.Next() {
		row := []interface{}{}

		columns, err := rs.Columns()

		if err != nil {
			j.ReportError(err)
			return
		}

		for index := 0; index < len(columns); index++ {
			var s interface{}
			row = append(row, &s)
		}

		err = rs.Scan(row...)

		if err != nil {
			j.ReportError(err)
			return
		}

		patchSource := strings.Replace(p.patch, "$$row", strconv.Itoa(rowIndex), -1)

		rowIndex += 1

		for index, col := range row {
			c := *(col.(*interface{}))

			j.Logf("%#T", c)

			switch c.(type) {
			case []uint8:
				col = string(c.([]uint8))
			}

			v, err := json.Marshal(col)

			if err != nil {
				j.ReportError(err)
				return
			}

			j.Logf("%s", v)

			patchSource = strings.Replace(patchSource, fmt.Sprintf(`"$$%d"`, index), string(v), -1)

			if err != nil {
				j.ReportError(err)
				return
			}
		}

		j.Logf("P: %s", patchSource)

		patch, err := jsonpatch.DecodePatch([]byte(patchSource))

		doc, err = patch.Apply(doc)

		if err != nil {
			j.ReportError(err)
			return
		}
	}

	err = json.Unmarshal(doc, &p.flow.Data)

	if err != nil {
		j.ReportError(err)
	}

	j.Logf("Posting flow %s", p.flow.Id)

	j.PostFlowUpdate(p.flow)

	j.Log("SQL plugin complete.")
}
