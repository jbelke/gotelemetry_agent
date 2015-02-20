package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ConfigFile struct {
	AllAccounts []AccountConfig
}

func NewConfigFile() (*ConfigFile, error) {
	source, err := ioutil.ReadFile(CLIConfig.ConfigFileLocation)

	if err != nil {
		if CLIConfig.IsPiping {
			return &ConfigFile{[]AccountConfig{AccountConfig{}}}, nil
		}

		return nil, errors.New(fmt.Sprintf("Unable to open configuration file at %s. Did you use --config to specify the right path?\n\n", CLIConfig.ConfigFileLocation))
	}

	result := &AccountConfig{}

	err = yaml.Unmarshal(source, result)

	return &ConfigFile{[]AccountConfig{*result}}, err
}

func (c *ConfigFile) Accounts() []AccountConfig {
	return c.AllAccounts
}
