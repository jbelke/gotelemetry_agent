package aggregations

import (
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
)

type Manager struct {
	path         string
	ttl          int
	errorChannel chan error
}

var manager *Manager = nil

func Init(cfg config.ConfigInterface, errorChannel chan error) error {
	dataConfig := cfg.DataConfig()

	if dataConfig.DataLocation != nil {
		var ttl int

		if dataConfig.DefaultTTL != nil {
			ttl = *dataConfig.DefaultTTL
		} else {
			ttl = -1
		}

		manager = &Manager{
			path:         *dataConfig.DataLocation,
			ttl:          ttl,
			errorChannel: errorChannel,
		}

		c, err := GetContext()

		if err != nil {
			return err
		}

		defer c.Close()

		c.Debugf("Writing data layer database to %s", manager.path)
		c.Debugf("Default data layer TTL is set to %d", manager.ttl)

		return nil
	}

	errorChannel <- gotelemetry.NewLogError("Data Manager -> No `data.path` property provided. The Data Manager will not run.")

	return nil
}
