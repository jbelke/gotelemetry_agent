package main

import (
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
	"log"

	_ "github.com/telemetryapp/gotelemetry_agent/plugin"
)

func banner() {
	println()
	println("Telemetry Agent v " + config.AgentVersion)
	println("Copyright © 2012-2015 Telemetry, Inc.")
	println()
	println("For license information, see the LICENSE file")
	println("---------------------------------------------")
	println()
}

func main() {
	banner()

	configFile, err := config.NewConfigFile()

	if err != nil {
		log.Fatalf("Initialization error: %s", err)
	}

	errorChannel := make(chan error, 0)
	completionChannel := make(chan bool, 0)

	_, err = job.NewJobManager(configFile, &errorChannel, &completionChannel)

	if err != nil {
		log.Fatalf("Initialization error: %s", err)
	}

	for {
		select {
		case <-completionChannel:
			goto Done

		case err := <-errorChannel:
			if e, ok := err.(*gotelemetry.Error); ok {
				logLevel := e.GetLogLevel()

				if logLevel >= config.CLIConfig.LogLevel {
					prefix := "Error"

					switch logLevel {
					case gotelemetry.LogLevelLog:
						prefix = "Log  "

					case gotelemetry.LogLevelDebug:
						prefix = "Debug"
					}

					log.Printf("%s: %s", prefix, err)
				}

				continue
			}

			log.Printf("Error: %s", err.Error())
		}
	}

Done:

	log.Println("No more jobs to run; exiting.")
}
