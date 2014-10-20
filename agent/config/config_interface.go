package config

type Job struct {
	ID     string                 `yaml:"id"`
	Plugin string                 `yaml:"plugin"`
	Config map[string]interface{} `yaml:"config"`
}

type AccountConfig struct {
	APIKey             string  `yaml:"api_key"`
	SubmissionInterval float64 `yaml:"submission_interval"`
	Jobs               []Job   `yaml:"jobs"`
}

type ConfigInterface interface {
	Accounts() []AccountConfig
}
