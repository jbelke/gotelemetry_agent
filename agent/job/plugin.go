package job

import (
	"github.com/telemetryapp/gotelemetry"
)

// Interfaces

type PluginInstance interface {
	Init(job *Job, config map[string]interface{}) error
	Run(job *Job)
	Reconfigure(job *Job, config map[string]interface{}) error
	Terminate(job *Job)
}

type PluginFactory interface {
	NewInstance() PluginInstance
}

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
