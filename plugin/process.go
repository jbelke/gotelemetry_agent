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
func init() {
	job.RegisterPlugin("com.telemetryapp.process", ProcessPluginFactory)
}

// Func ProcessPluginFactory generates a blank instance of the
// `com.telemetryapp.process` plugin
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

// Function Init initializes the plugin.
//
// The required configuration parameters are:
//
// - path                         The executable's path
//
// - flow_tag                     The tag of the flow to populate
//
// - variant                      The varient of the flow
//
// - template                     A template that will be used to populate the flow when it is created
//
// When the plugin is executed, it loads up the current state of the flow
// and sends it to the external process as its only parameter.
//
// In output, the process has two options:
//
// - Output a JSON payload, which is used to replace the payload of the flow, which is then automatically submitted to the Telemetry API
//
// - Output the text PATCH, followed by a newline, followed by a JSON-Patch payload that is applied to the flow.
//
// For example:
//
//  jobs:
//    - id: Telemetry External
//      plugin: com.telemetryapp.process
//      config:
//        refresh: 86400
//        path: ./test.php
//        flow_tag: php_test
//        variant: value
//        template:
//          color: white
//          label: PHP Test
//          value: 100
//
// test.php:
//
//   #!/usr/bin/php
//   <?php
//   echo "PATCH\n";
//   echo '[{"op":"replace", "path":"/value", "value":123}]';
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
