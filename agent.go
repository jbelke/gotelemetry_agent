package main

import (
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
	"log"

	_ "github.com/telemetryapp/gotelemetry_agent/plugin"
)

func main() {
	config, err := config.NewConfigFile()

	if err != nil {
		panic(err)
	}

	errorChannel := make(chan error, 0)
	completionChannel := make(chan bool, 0)

	_, err = job.NewJobManager(config, &errorChannel, &completionChannel)

	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-completionChannel:
			goto Done

		case err := <-errorChannel:
			log.Printf("Error: %s", err.Error())
		}
	}

Done:

	log.Println("No more jobs to run; exiting.")
}
