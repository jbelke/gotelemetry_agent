package config

import (
	"flag"
)

type CLIConfigType struct {
	ConfigFileLocation string
}

var CLIConfig CLIConfigType

func init() {
	flag.StringVar(&CLIConfig.ConfigFileLocation, "config", "/var/telemetry/gotelemetry_agent.yaml", "Location of the agent configuration file")

	flag.Parse()
}
