package aggregations

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"fmt"
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
)

type AggregationManager struct {
	path         string
	ttl          int
	errorChannel chan error
	seriesCache  map[string]*Series
}

var manager *AggregationManager = nil

func Init(cfg config.ConfigInterface, errorChannel chan error) error {
	dataConfig := cfg.DataConfig()

	if dataConfig.DataLocation != nil {
		var ttl int

		if dataConfig.DefaultTTL != nil {
			ttl = *dataConfig.DefaultTTL
		} else {
			ttl = -1
		}

		manager = &AggregationManager{
			path:         *dataConfig.DataLocation,
			ttl:          ttl,
			errorChannel: errorChannel,
			seriesCache:  map[string]*Series{},
		}

		manager.Debugf("Writing data layer database to %s", manager.path)
		manager.Debugf("Default data layer TTL is set to %d", manager.ttl)

		return manager.createMetadataTable()
	}

	errorChannel <- gotelemetry.NewLogError("Data Manager -> No `data.path` property provided. The Data Manager will not run.")

	return nil
}

// Logf sends a formatted string to the agent's global log. It works like log.Logf
func (a *AggregationManager) Logf(format string, v ...interface{}) {
	if a.errorChannel != nil {
		a.errorChannel <- gotelemetry.NewLogError("Data Manager -> %#s", fmt.Sprintf(format, v...))
	}
}

// Debugf sends a formatted string to the agent's debug log, if it exists. It works like log.Logf
func (a *AggregationManager) Debugf(format string, v ...interface{}) {
	if a.errorChannel != nil {
		a.errorChannel <- gotelemetry.NewDebugError("Data Manager -> %#s", fmt.Sprintf(format, v...))
	}
}

func (a *AggregationManager) withConnection(closure func(c *sqlite3.Conn) error) error {
	c, err := sqlite3.Open(a.path)

	if c != nil {
		defer c.Close()
	}

	if err != nil {
		return err
	}

	return closure(c)
}

func (a *AggregationManager) createMetadataTable() error {
	return a.withConnection(func(c *sqlite3.Conn) error {
		return c.Exec("CREATE TABLE IF NOT EXISTS telemetry_metadata (name VARCHAR PRIMARY KEY, ttl INT)")
	})
}

func (a *AggregationManager) saveSeriesMetadata(s *Series) error {
	if err := a.createMetadataTable(); err != nil {
		return err
	}

	return a.withConnection(func(c *sqlite3.Conn) error {
		return c.Exec("INSERT OR REPLACE INTO telemetry_metadata (name, ttl) VALUES (?, ?)", s.Name, s.TTL)
	})
}

func (a *AggregationManager) loadSeriesMetadata(s *Series) error {
	if err := a.createMetadataTable(); err != nil {
		return err
	}

	return a.withConnection(func(c *sqlite3.Conn) error {
		rs, err := c.Query("SELECT ttl FROM telemetry_metadata WHERE name = ?", s.Name)

		if err != nil {
			return err
		}

		defer rs.Close()

		row := make(sqlite3.RowMap)

		if err = rs.Scan(row); err != nil {
			return err
		}

		s.TTL = row["ttl"].(int)

		return nil
	})
}
