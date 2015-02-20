package config

import (
	"github.com/alecthomas/kingpin"
	"github.com/telemetryapp/gotelemetry"
	"log"
	"regexp"
)

type CLIConfigType struct {
	ConfigFileLocation string
	LogLevel           gotelemetry.LogLevel
	Filter             *regexp.Regexp
	ForceRunOnce       bool
	IsPiping           bool
	UseJSONPatch       bool
	UsePOST            bool
}

const AgentVersion = "1.2"

var CLIConfig CLIConfigType

func banner() {
	println()
	println("Telemetry Agent v " + AgentVersion)
	println("Copyright © 2012-2015 Telemetry, Inc.")
	println()
	println("For license information, see the LICENSE file")
	println("---------------------------------------------")
	println()
}

func init() {
	banner()
	kingpin.Version(AgentVersion)

	kingpin.Flag("config", "Path to the configuration file for this agent.").Short('c').Default("./gotelemetry_agent.yaml").StringVar(&CLIConfig.ConfigFileLocation)
	kingpin.Flag("once", "Run all jobs exactly once and exit.").Default("false").BoolVar(&CLIConfig.ForceRunOnce)

	kingpin.Flag("pipe", "Accept a Rails-style HTTP PATCH Telemetry payload from stdin, send it to the API, and then exit.").Default("false").BoolVar(&CLIConfig.IsPiping)
	kingpin.Flag("jsonpatch", "With --pipe, submit the package as a JSON-Patch request instead. Ignored otherwise.").BoolVar(&CLIConfig.UseJSONPatch)
	kingpin.Flag("post", "With --pipe, submit the package as a POST request instead. Ignored otherwise.").BoolVar(&CLIConfig.UsePOST)

	logLevel := kingpin.Flag("verbosity", "Set the verbosity level (`debug`, `log`, `error`).").Short('v').Default("log").Enum("debug", "log", "error")
	filter := kingpin.Flag("filter", "Run only the jobs whose IDs (or tags if no ID is specified) match the given regular expression").Default(".").String()

	kingpin.Parse()

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
