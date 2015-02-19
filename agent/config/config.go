package config

import (
	"flag"
	"github.com/telemetryapp/gotelemetry"
	"log"
	"regexp"
)

type CLIConfigType struct {
	ConfigFileLocation string
	LogLevel           gotelemetry.LogLevel
	Filter             *regexp.Regexp
	ForceRunOnce       bool
}

var CLIConfig CLIConfigType

func init() {
	flag.StringVar(&CLIConfig.ConfigFileLocation, "config", "/var/telemetry/gotelemetry_agent.yaml", "Location of the agent configuration file")
	flag.BoolVar(&CLIConfig.ForceRunOnce, "once", false, "Run all jobs exactly once and exit")

	filter := flag.String("filter", "", "Run only the jobs whose IDs (or tags if no ID is specified) match the given regular expression")
	logLevel := flag.String("v", "log", "Set the verbosity level (`debug`, `log`, `error`; default `log`)")

	flag.Parse()

	switch *logLevel {
	case "debug":
		CLIConfig.LogLevel = gotelemetry.LogLevelDebug

	case "log":
		CLIConfig.LogLevel = gotelemetry.LogLevelLog

	case "error":
		CLIConfig.LogLevel = gotelemetry.LogLevelError

	default:
		log.Fatalf("Invalid verbosity level `%s`", logLevel)
	}

	rx, err := regexp.Compile(*filter)

	if err != nil {
		log.Fatalf("Invalid regular expression provided for -filter: %s", err)
	}

	CLIConfig.Filter = rx
}
