package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ConfigFile struct {
	AllAccounts []AccountConfig `yaml:"accounts"`
}

func NewConfigFile() (*ConfigFile, error) {
	source, err := ioutil.ReadFile(CLIConfig.ConfigFileLocation)

	if err != nil {
		return nil, err
	}

	result := &ConfigFile{}

	err = yaml.Unmarshal(source, result)

	return result, err
}

func (c *ConfigFile) Accounts() []AccountConfig {
	return c.AllAccounts
}
