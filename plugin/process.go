package plugin

import (
	"encoding/json"
	"errors"
	"github.com/evanphx/json-patch"
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
	"os"
	"os/exec"
	"strings"
	"time"
)

// init() registers this plugin with the Plugin Manager.
// The plugin provides a ProcessPluginFactory that the manager calls whenever it needs
// to create a new job
func init() {
	job.RegisterPlugin("com.telemetryapp.process", ProcessPluginFactory)
}

// Func ProcessPluginFactory generates a blank plugin instance
func ProcessPluginFactory() job.PluginInstance {
	return &ProcessPlugin{
		PluginHelper: job.NewPluginHelper(),
	}
}

// Struct ProcessPlugin allows the agent to execute an external process and use its
// output as data that can be fed to the Telemetry API.
//
// For configuration parameters, see the Init() function
type ProcessPlugin struct {
	*job.PluginHelper
	tag      string
	path     string
	variant  string
	template map[string]interface{}
	flow     *gotelemetry.Flow
}

// Init initializes the plugin.
func (p *ProcessPlugin) Init(job *job.Job) error {
	c := job.Config()

	p.tag = c["flow_tag"].(string)
	p.path = c["path"].(string)
	p.variant = c["variant"].(string)
	p.template = config.MapFromYaml(c["template"]).(map[string]interface{})

	if _, err := os.Stat(p.path); os.IsNotExist(err) {
		return errors.New("File " + p.path + " does not exist.")
	}

	if f, err := job.GetOrCreateFlow(p.tag, p.variant, "gotelemetry_agent", p.template); err != nil {
		return err
	} else {
		p.flow = f
	}

	if refresh, ok := c["refresh"]; ok {
		p.PluginHelper.AddTaskWithClosure(p.performAllTasks, time.Duration(refresh.(int))*time.Second)
	} else {
		p.PluginHelper.AddTaskWithClosure(p.performAllTasks, 0)
	}

	return nil
}

func (p *ProcessPlugin) performAllTasks(j *job.Job) {
	if err := j.ReadFlow(p.flow); err != nil {
		j.ReportError(err)
		return
	}

	if flowData, err := json.Marshal(p.flow.Data); err != nil {
		j.ReportError(err)
		return
	} else {
		out, err := exec.Command(p.path, string(flowData)).Output()

		if err != nil {
			j.ReportError(err)
			return
		}

		if strings.HasPrefix(string(out), "PATCH") {
			out = []byte(strings.TrimPrefix(string(out), "PATCH\n"))

			patch, err := jsonpatch.DecodePatch(out)

			if err != nil {
				j.ReportError(err)
				return
			}

			flowData, err = patch.Apply(flowData)

			if err != nil {
				j.ReportError(err)
				return
			}

			if err := json.Unmarshal(flowData, p.flow.Data); err != nil {
				j.ReportError(err)
				return
			}
		} else {
			if err := json.Unmarshal(out, p.flow.Data); err != nil {
				j.ReportError(err)
				return
			}
		}

		j.Logf("Posting flow %s", p.flow.Id)

		j.Logf("%#v", p.flow.Data)

		j.PostFlowUpdate(p.flow)

		j.Log("SQL plugin complete.")
	}

}
