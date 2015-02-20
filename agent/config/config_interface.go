package config

import (
	"errors"
	"os"
)

func MapFromYaml(from interface{}) interface{} {
	switch from.(type) {
	case map[interface{}]interface{}:
		result := map[string]interface{}{}

		for index, value := range from.(map[interface{}]interface{}) {
			result[index.(string)] = MapFromYaml(value)
		}

		return result

	case []interface{}:
		f := from.([]interface{})

		for index, value := range f {
			f[index] = MapFromYaml(value)
		}

		return f

	default:
		return from
	}
}

type Job struct {
	ID     string                 `yaml:"id"`
	Plugin string                 `yaml:"plugin"`
	Config map[string]interface{} `yaml:"config"`
	Then   []Job                  `yaml:"then"`
}

type AccountConfig struct {
	APIKey             string  `yaml:"api_key"`
	APIToken           string  `yaml:"api_token"`
	SubmissionInterval float64 `yaml:"submission_interval"`
	Jobs               []Job   `yaml:"jobs"`
}

type ConfigInterface interface {
	Accounts() []AccountConfig
}

func (a AccountConfig) GetAPIKey() (string, error) {
	result := a.APIKey

	if result == "" {
		result = a.APIToken
	}

	if result == "" {
		result = os.ExpandEnv("$TELEMETRY_API_TOKEN")
	}

	if result != "" {
		return result, nil
	}

	return "", errors.New("No API Token found in the configuration file or in the TELEMETRY_API_TOKEN environment variable.")
}
