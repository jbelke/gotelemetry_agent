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
func init() {
	job.RegisterPlugin("com.telemetryapp.sql", SQLPluginFactory)
}

// Func IntercomPluginFactory generates a blank plugin instance of the
// `com.telemetryapp.sql` plugin
func SQLPluginFactory() job.PluginInstance {
	return &SQLPlugin{
		PluginHelper: job.NewPluginHelper(),
	}
}

// Struct SQLPlugin is allows populating flows based on the content of a SQL
// database. For configuration parameters, see Init()
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
// The required configuration parameters are:
//
// - driver                       The SQL driver to use
//
// - datasource                   The datasource on which to operate
//
// - query                        The query to be executed
//
// - flow_tag                     The tag of the flow to populate
//
// - variant                      The varient of the flow
//
// - template                     A template that will be used to populate the flow when it is created
//
// - patch                        A JSON Patch payload that describes how the data extracted from the database must be applied to the flow
//
// The patch is executed once for each row; you can use $$row as a placeholder for
// the number of the current row, and $$n as a placeholder for the value of column
// n in the current row.
//
// For example:
//
//   - id: Users with five or more sessions
//     plugin: com.telemetryapp.sql
//     config:
//       driver: sqlite3
//       datasource: /tmp/telemetry_intercom.sqlite
//       query: "select count(*) / cast(t.total as real) * 100 from intercom_users cross join (select count(*) as total from intercom_users) as t where session_count >= 5;"
//       patch:
//         - { "op": "replace" , "path": "/value", "value": $$0 }
//       flow_tag: users_5_sessions
//       variant: value
//       template:
//         color: white
//         label: Frequent Users
//         value_type: percent
//         value: 100
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

	if err := j.ReadFlow(p.flow); err != nil {
		j.ReportError(err)
		return
	}

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

			patchSource = strings.Replace(patchSource, fmt.Sprintf(`"$$%d"`, index), string(v), -1)

			if err != nil {
				j.ReportError(err)
				return
			}
		}

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
