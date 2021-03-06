package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"github.com/telemetryapp/gotelemetry_agent/agent/functions"
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
	expiration time.Duration
	flowTag    string
	path       string
	args       []string
	template   map[string]interface{}
	flow       *gotelemetry.Flow
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
// - refresh                      The number of seconds between subsequent executions of the
//                                plugin. Default: never
//
// - expiration										The number of seconds after which flow data is set to expire.
//                                Default: refresh * 3; 0 = never.
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

	if expiration, ok := c["expiration"].(int); ok {
		if expiration < 0 {
			return errors.New("Invalid expiration time")
		}

		p.expiration = time.Duration(expiration) * time.Second
	}

	if refresh, ok := c["refresh"].(int); ok {
		if p.expiration == 0 {
			p.expiration = time.Duration(refresh*3) * time.Second
		}

		p.PluginHelper.AddTaskWithClosure(p.performAllTasks, time.Duration(refresh)*time.Second)
	} else {
		p.PluginHelper.AddTaskWithClosure(p.performAllTasks, 0)
	}

	if p.expiration > 0 {
		job.Debugf("Expiration is set to %dµs", p.expiration)
	} else {
		job.Debugf("Expiration is off.")
	}

	return nil
}

func (p *ProcessPlugin) analyzeAndSubmitProcessResponse(j *job.Job, response string) error {
	isJSONPatch := false
	isReplace := false

	if strings.HasPrefix(response, "REPLACE\n") {
		isReplace = true
		response = strings.TrimPrefix(response, "REPLACE\n")
	} else if strings.HasPrefix(response, "PATCH\n") {
		isJSONPatch = true
		response = strings.TrimPrefix(response, "PATCH\n")
	}

	context, err := aggregations.GetContext()

	if err != nil {
		return err
	}

	defer context.Close()

	hasData := false
	data := map[string]interface{}{}

	for _, command := range strings.Split(response, "\n") {
		commandData := map[string]interface{}{}

		command = strings.TrimSpace(command)

		if command == "" {
			continue
		}

		err := json.Unmarshal([]byte(command), &commandData)

		if err != nil {
			context.SetError()
			return err
		}

		if d, err := functions.Parse(context, commandData); err == nil {
			switch d.(type) {
			case map[string]interface{}:
				if hasData {
					return errors.New("Multiple data-bearing commands detected.")
				}

				data = d.(map[string]interface{})
				hasData = true

			default:
				// Do nothing
			}

		} else {
			context.SetError()
			return err
		}
	}

	if !hasData {
		j.Debugf("No data-bearing command found. Skipping API operations")
		return nil
	}

	if isJSONPatch {
		if p.expiration > 0 {
			j.Logf("Warning: Forced expiration is not supported for JSON-Patch operations")
		}

		j.QueueDataUpdate(p.flowTag, data, gotelemetry.BatchTypeJSONPATCH)
	} else if isReplace {
		if p.expiration > 0 {
			newExpiration := time.Now().Add(p.expiration)
			newUnixExpiration := newExpiration.Unix()

			j.Debugf("Forcing expiration to %d (%s)", newUnixExpiration, newExpiration)

			data["expires_at"] = newUnixExpiration
		}

		j.QueueDataUpdate(p.flowTag, data, gotelemetry.BatchTypePOST)
	} else {
		if p.expiration > 0 {
			newExpiration := time.Now().Add(p.expiration)
			newUnixExpiration := newExpiration.Unix()

			j.Debugf("Forcing expiration to %d (%s)", newUnixExpiration, newExpiration)

			data["expires_at"] = newUnixExpiration
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
		j.SetFlowError(p.flowTag, map[string]interface{}{"error": err.Error(), "output": string(out)})
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
