package job

import (
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
	"time"
)

type JobManager struct {
	credentials    map[string]gotelemetry.Credentials
	accountStreams map[string]*gotelemetry.BatchStream
	jobs           map[string]*Job
}

func NewJobManager(config config.ConfigInterface, errorChannel *chan error) (*JobManager, error) {
	result := &JobManager{
		credentials:    map[string]gotelemetry.Credentials{},
		accountStreams: map[string]*gotelemetry.BatchStream{},
		jobs:           map[string]*Job{},
	}

	for _, account := range config.Accounts() {
		var err error

		apiKey := account.APIKey

		credentials, success := result.credentials[apiKey]

		if !success {
			credentials, err = gotelemetry.NewCredentials(account.APIKey)

			if err != nil {
				return nil, err
			}

			result.credentials[apiKey] = credentials
		}

		accountStream, success := result.accountStreams[apiKey]

		if !success {
			accountStream, err = gotelemetry.NewBatchStream(credentials, time.Duration(account.SubmissionInterval)*time.Second, errorChannel)

			if err != nil {
				return nil, err
			}

			result.accountStreams[apiKey] = accountStream
		}

		for _, jobDescription := range account.Jobs {
			jobId := jobDescription.ID

			_, success := result.jobs[jobId]

			if success {
				return nil, gotelemetry.NewError(500, "Duplicate job `"+jobId+"`")
			}

			pluginFactory, err := GetPlugin(jobDescription.Plugin)

			if err != nil {
				return nil, err
			}

			pluginInstance := pluginFactory.NewInstance()

			job, err := newJob(credentials, accountStream, jobDescription.ID, jobDescription.Config, pluginInstance, errorChannel)

			if err != nil {
				return nil, err
			}

			result.jobs[job.ID] = job
		}
	}

	return result, nil
}
