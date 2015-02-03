package job

import (
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
	"time"
)

type JobManager struct {
	credentials          map[string]gotelemetry.Credentials
	accountStreams       map[string]*gotelemetry.BatchStream
	jobs                 map[string]*Job
	completionChannel    *chan bool
	jobCompletionChannel chan string
}

func createJob(credentials gotelemetry.Credentials, accountStream *gotelemetry.BatchStream, errorChannel *chan error, jobDescription config.Job, jobCompletionChannel chan string, wait bool) (*Job, error) {
	pluginFactory, err := GetPlugin(jobDescription.Plugin)

	if err != nil {
		return nil, err
	}

	pluginInstance := pluginFactory()

	then := []*Job{}

	for _, jobConfig := range jobDescription.Then {
		job, err := createJob(credentials, accountStream, errorChannel, jobConfig, jobCompletionChannel, true)

		if err != nil {
			return nil, err
		}

		then = append(then, job)
	}

	return newJob(credentials, accountStream, jobDescription.ID, jobDescription.Config, then, pluginInstance, errorChannel, jobCompletionChannel, wait)
}

func NewJobManager(config config.ConfigInterface, errorChannel *chan error, completionChannel *chan bool) (*JobManager, error) {
	result := &JobManager{
		credentials:          map[string]gotelemetry.Credentials{},
		accountStreams:       map[string]*gotelemetry.BatchStream{},
		jobs:                 map[string]*Job{},
		completionChannel:    completionChannel,
		jobCompletionChannel: make(chan string),
	}

	for _, account := range config.Accounts() {
		var err error

		apiKey := account.APIKey

		if apiKey == "" {
			apiKey = account.APIToken
		}

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

			job, err := createJob(credentials, accountStream, errorChannel, jobDescription, result.jobCompletionChannel, false)

			if err != nil {
				return nil, err
			}

			result.jobs[job.ID] = job
		}
	}

	go result.monitorDoneChannel()

	return result, nil
}

func (m *JobManager) monitorDoneChannel() {
	for {
		select {
		case id := <-m.jobCompletionChannel:
			delete(m.jobs, id)

			if len(m.jobs) == 0 {
				*m.completionChannel <- true
				return
			}
		}
	}
}
