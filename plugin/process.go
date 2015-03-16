package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/telemetryapp/gotelemetry"
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
	flowTag  string
	path     string
	args     []string
	template map[string]interface{}
	flow     *gotelemetry.Flow
}

// Function Init initializes the plugin.
//
// The required configuration parameters are:
//
// - path                         The executable's path
//
// - args													An array of arguments that are sent to the executable
//
// - flow_tag                     The tag of the flow to populate
//
// - variant                      The variant of the flow
//
// - template                     A template that will be used to populate the flow when it is created
//
// If `variant` and `template` are both specified, the plugin will verify that the flow exists and is of the
// correct variant on startup. In that case, if the flow is found but is of the wrong variant, an error is
// output to log and the plugin is not allowed to run. If the flow does not exist, it is created using
// the contents of `template`. If the creation fails, the plugin is not allowed to run.
//
// In output, the process has two options:
//
// - Output a JSON payload, which is used to PATCH the payload of the flow using a simple top-level property replacement operation
//
// - Output the text PATCH, followed by a newline, followed by a JSON-Patch payload that is applied to the flow.
//
// - Output the text REPLACE, followed by a newline, followed by a payload that is used to replace the contents of the flow.
//
// For example:
//
//  jobs:
//    - id: Telemetry External
//      plugin: com.telemetryapp.process
//      config:
//        refresh: 86400
//        path: ./test.php
//        args:
//        	- value
//        	- 1
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
//   echo '[{"op":"replace", "path":"/value", "value":' + $argv[2] + '}]';

func (p *ProcessPlugin) Init(job *job.Job) error {
	var ok bool

	c := job.Config()

	job.Debugf("The configuration is %#v", c)

	p.flowTag, ok = c["flow_tag"].(string)

	if !ok {
		return errors.New("The required `flow_tag` property (`string`) is either missing or of the wrong type.")
	}

	p.path, ok = c["path"].(string)

	if !ok {
		return errors.New("The required `path` property (`string`) is either missing or of the wrong type.")
	}

	p.args = []string{}

	if args, ok := c["args"].([]interface{}); ok {
		for _, arg := range args {
			if a, ok := arg.(string); ok {
				p.args = append(p.args, a)
			} else {
				p.args = append(p.args, fmt.Sprintf("%#v", arg))
			}
		}
	}

	if _, err := os.Stat(p.path); os.IsNotExist(err) {
		return errors.New("File " + p.path + " does not exist.")
	}

	template, templateOK := c["template"]
	variant, variantOK := c["variant"].(string)

	if variantOK && templateOK {
		if f, err := job.GetOrCreateFlow(p.flowTag, variant, template); err != nil {
			return err
		} else {
			p.flow = f
		}
	}

	if refresh, ok := c["refresh"].(int); ok {
		p.PluginHelper.AddTaskWithClosure(p.performAllTasks, time.Duration(refresh)*time.Second)
	} else {
		p.PluginHelper.AddTaskWithClosure(p.performAllTasks, 0)
	}

	return nil
}

func (p *ProcessPlugin) analyzeAndSubmitProcessResponse(j *job.Job, response string) error {
	var data interface{}

	if strings.HasPrefix(response, "PATCH\n") {
		err := json.Unmarshal([]byte(strings.TrimPrefix(response, "PATCH\n")), &data)

		if err != nil {
			return err
		}

		j.QueueDataUpdate(p.flowTag, data, gotelemetry.BatchTypeJSONPATCH)
	} else if strings.HasPrefix(response, "REPLACE\n") {
		err := json.Unmarshal([]byte(strings.TrimPrefix(response, "REPLACE\n")), &data)

		if err != nil {
			return err
		}

		j.QueueDataUpdate(p.flowTag, data, gotelemetry.BatchTypePOST)
	} else {
		err := json.Unmarshal([]byte(response), &data)

		if err != nil {
			return err
		}

		j.QueueDataUpdate(p.flowTag, data, gotelemetry.BatchTypePATCH)
	}

	return nil
}

func (p *ProcessPlugin) performAllTasks(j *job.Job) {
	j.Debugf("Starting process plugin...")

	defer p.PluginHelper.TrackTime(j, time.Now(), "Process plugin completed in %s.")

	if len(p.args) > 0 {
		j.Debugf("Executing `%s` with arguments %#v", p.path, p.args)
	} else {
		j.Debugf("Executing `%s` with no arguments", p.path)
	}

	out, err := exec.Command(p.path, p.args...).Output()

	if err != nil {
		j.ReportError(err)
		return
	}

	response := string(out)

	j.Debugf("Process output: %s", strings.Replace(response, "\n", "\\n", -1))
	j.Debugf("Posting flow %s", p.flowTag)

	if err := p.analyzeAndSubmitProcessResponse(j, response); err != nil {
		j.ReportError(errors.New("Unable to analyze process output: " + err.Error()))
	}
}
