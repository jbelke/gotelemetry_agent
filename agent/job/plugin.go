package job

import (
	"github.com/telemetryapp/gotelemetry"
)

// Interfaces

// Interface PluginInstance represents an instance of a plugin that is called upon to perform a job.
//
// In other words, this is what you are supposed to write.
type PluginInstance interface {
	Init(job *Job) error                                       // Initializes the instance
	Run(job *Job)                                              // Runs the instance synchronously until Terminate() is called
	Reconfigure(job *Job, config map[string]interface{}) error // Dynamically reconfigures the instance, or returns an error if it can't
	Terminate(job *Job)                                        // Terminates the instance, returning only when its execution is complete
}

type PluginFactory func() PluginInstance

// Manager

type pluginManager struct {
	factories map[string]PluginFactory
}

var pluginErrors = []error{}

var manager = &pluginManager{map[string]PluginFactory{}}

func RegisterPlugin(name string, factory PluginFactory) {
	if _, exists := manager.factories[name]; exists {
		pluginErrors = append(pluginErrors, gotelemetry.NewError(500, "Duplicate plugin name `"+name+"`"))
		return
	}

	manager.factories[name] = factory
}

func AllPlugins() map[string]PluginFactory {
	return manager.factories
}

func GetPlugin(name string) (PluginFactory, error) {
	result, success := manager.factories[name]

	if !success {
		return nil, gotelemetry.NewError(500, "Plugin `"+name+"` not found")
	}

	return result, nil
}

func GetPluginRegistrationErrors() []error {
	return pluginErrors
}
