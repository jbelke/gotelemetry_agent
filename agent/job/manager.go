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
	completionChannel    chan bool
	jobCompletionChannel chan string
}

func createJob(credentials gotelemetry.Credentials, accountStream *gotelemetry.BatchStream, errorChannel chan error, jobDescription config.Job, jobCompletionChannel chan string, wait bool) (*Job, error) {
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

func NewJobManager(jobConfig config.ConfigInterface, errorChannel chan error, completionChannel chan bool) (*JobManager, error) {
	result := &JobManager{
		credentials:          map[string]gotelemetry.Credentials{},
		accountStreams:       map[string]*gotelemetry.BatchStream{},
		jobs:                 map[string]*Job{},
		completionChannel:    completionChannel,
		jobCompletionChannel: make(chan string),
	}

	for _, account := range jobConfig.Accounts() {
		var err error

		apiKey, err := account.GetAPIKey()

		if err != nil {
			return nil, err
		}

		credentials, success := result.credentials[apiKey]

		if !success {
			credentials, err = gotelemetry.NewCredentials(apiKey)

			if err != nil {
				return nil, err
			}

			credentials.SetDebugChannel(errorChannel)

			result.credentials[apiKey] = credentials
		}

		accountStream, success := result.accountStreams[apiKey]

		if !success {
			submissionInterval := time.Duration(account.SubmissionInterval) * time.Second

			if submissionInterval < time.Second {
				errorChannel <- gotelemetry.NewLogError("Submission interval automatically set to 1s. You can change this value by adding a `submission_interval` property to your configuration file.")
				submissionInterval = time.Second
			}

			accountStream, err = gotelemetry.NewBatchStream(credentials, submissionInterval, errorChannel)

			if err != nil {
				return nil, err
			}

			result.accountStreams[apiKey] = accountStream
		}

		for _, jobDescription := range account.Jobs {
			jobId := jobDescription.ID

			if jobId == "" {
				if tag, ok := jobDescription.Config["flow_tag"].(string); ok {
					jobId = tag
					jobDescription.ID = tag
				} else {
					return nil, gotelemetry.NewError(500, "Job ID missing and no `flow_tag` provided.")
				}
			}

			if !config.CLIConfig.Filter.MatchString(jobId) {
				continue
			}

			if config.CLIConfig.ForceRunOnce {
				delete(jobDescription.Config, "refresh")
			}

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

	if len(result.jobs) == 0 {
		return nil, gotelemetry.NewError(400, "No jobs to run. Exiting.")
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
				for _, stream := range m.accountStreams {
					stream.Flush()
				}

				m.completionChannel <- true
				return
			}
		}
	}
}
