package main

import (
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent"
	"github.com/telemetryapp/gotelemetry_agent/agent/aggregations"
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/telemetryapp/gotelemetry_agent/plugin"
)

var configFile *config.ConfigFile
var errorChannel chan error
var completionChannel chan bool

func main() {
	var err error

	configFile, err = config.NewConfigFile()

	if err != nil {
		log.Fatalf("Initialization error: %s", err)
	}

	errorChannel = make(chan error, 0)
	completionChannel = make(chan bool, 0)

	go run()

	for {
		select {
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

		case <-completionChannel:
			goto Done
		}
	}

Done:

	log.Println("No more jobs to run; exiting.\n")
}

func run() {
	err := aggregations.Init(configFile, errorChannel)

	if err != nil {
		log.Fatalf("Initialization error: %s", err)
	}

	if config.CLIConfig.IsPiping {
		payload, err := ioutil.ReadAll(os.Stdin)

		if err != nil {
			errorChannel <- err
		}

		agent.ProcessPipeRequest(configFile, errorChannel, completionChannel, payload)
	} else if config.CLIConfig.IsNotifying {
		agent.ProcessNotificationRequest(configFile, errorChannel, completionChannel, config.CLIConfig.NotificationChannel, config.CLIConfig.Notification)
	} else {
		_, err := job.NewJobManager(configFile, errorChannel, completionChannel)

		if err != nil {
			log.Fatalf("Initialization error: %s", err)
		}
	}
}
