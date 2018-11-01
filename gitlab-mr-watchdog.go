package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// TriggerConfig struct for YAML file
type TriggerConfig struct {
	GitLab struct {
		Owner   string `yaml:"owner"`
		Project string `yaml:"project"`
	} `yaml:"GitLab"`
}

// Read config YAML file, then return Config
func (config *TriggerConfig) read(file string) *TriggerConfig {
	yamlFile, err := ioutil.ReadFile(file)
	printErrorThenExit(err, "Read YAML file error")

	err = yaml.Unmarshal(yamlFile, config)
	printErrorThenExit(err, "YAML unmarshal error")

	return config
}

// Validate YAML file configs
func (config *TriggerConfig) validate() {
	var err error
	switch {
	case config.GitLab.Owner == "":
		err = errors.New("GitLab owner is required")
	case config.GitLab.Project == "":
		err = errors.New("GitLab project is required")
	}

	if err != nil {
		printErrorThenExit(err, "YAML file configs validate error")
	}
}

// Print error message, then exit program
func printErrorThenExit(err error, message string) {
	if err != nil {
		if message != "" {
			fmt.Fprintf(os.Stderr, fmt.Sprintf(message+": [%v]", err)+"\n")
		}

		flag.Usage()
		os.Exit(1)
	}
}

func main() {

}
