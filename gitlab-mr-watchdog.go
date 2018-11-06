package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// WatchdogConfig struct for YAML file
type WatchdogConfig struct {
	GitLab struct {
		Host     string `yaml:"host"`
		Group    string `yaml:"group"`
		Username string `yaml:"username"`
		Project  string `yaml:"project"`
		Token    string `yaml:"token"`
	} `yaml:"GitLab"`
	Watchdog struct {
		Duration float64 `yaml:"duration"`
	} `yaml:"Watchdog"`
}

// Read config YAML file, then return Config
func (config *WatchdogConfig) read(file string) *WatchdogConfig {
	yamlFile, err := ioutil.ReadFile(file)
	printErrorThenExit(err, "Read YAML file error")

	err = yaml.Unmarshal(yamlFile, config)
	printErrorThenExit(err, "YAML unmarshal error")

	return config
}

// Validate YAML file configs
func (config *WatchdogConfig) validate() {
	var err error
	switch {
	case config.GitLab.Host == "":
		err = errors.New("GitLab host is required")
	case config.GitLab.Group == "" && config.GitLab.Username == "":
		err = errors.New("GitLab group or username is required")
	case config.GitLab.Project == "":
		err = errors.New("GitLab project is required")
	case config.GitLab.Token == "":
		err = errors.New("GitLab token is required")
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

// GitLabUtility struct for GitLab uitility properties & funcs
type GitLabUtility struct {
	host     string
	token    string
	username string
	group    string
}

// GitLabGroupResponse for GitLab API response of `group`
type GitLabGroupResponse struct {
	Projects []GitLabProject `json:"projects"`
}

// GitLabProject for project structure in GitLabGroupResponse
type GitLabProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GitLabMergeRequest for GitLab merge requests structure
type GitLabMergeRequest struct {
	IID       string `json:"iid"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	WIP       string `json:"work_in_progress"`
	Author    struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"author"`
}

// Fetch GitLab merge requests by ID
func (utility *GitLabUtility) fetchMergeRequestsByID(id int, params string) ([]GitLabMergeRequest, error) {
	apiURL := utility.host + "/api/v4/projects/" + string(id) + "/merge_requests/" + params
	request, _ := http.NewRequest("GET", apiURL, nil)
	request.Header.Set("Private-Token", utility.token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return []GitLabMergeRequest{}, err
	}

	defer response.Body.Close()

	// Print info when success or failure
	switch response.StatusCode {
	case 200:
		body, _ := ioutil.ReadAll(response.Body)
		var models []GitLabMergeRequest

		json.Unmarshal(body, &models)

		return models, nil
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return []GitLabMergeRequest{}, errors.New("Unknown: " + string(body))
	}
}

// Fetch GitLab group's projects
func (utility *GitLabUtility) fetchGroupProjects() ([]GitLabProject, error) {
	apiURL := utility.host + "/api/v4/groups/" + utility.group
	request, _ := http.NewRequest("GET", apiURL, nil)
	request.Header.Set("Private-Token", utility.token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return []GitLabProject{}, err
	}

	defer response.Body.Close()

	// Print info when success or failure
	switch response.StatusCode {
	case 200:
		body, _ := ioutil.ReadAll(response.Body)
		var model GitLabGroupResponse

		json.Unmarshal(body, &model)

		return model.Projects, nil
	case 404:
		return []GitLabProject{}, errors.New("----")
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return []GitLabProject{}, errors.New("Unknown: " + string(body))
	}
}

func main() {
	// Read config file path from command line
	var configFilePath = flag.String("path", "config.yml", "Setup your configuration file path.")
	flag.Parse()

	// Read & validate config.yml
	var config WatchdogConfig
	config.read(*configFilePath)
	config.validate()

	gitlab := GitLabUtility{}
	gitlab.host = config.GitLab.Host
	gitlab.group = config.GitLab.Group
	gitlab.username = config.GitLab.Username
	gitlab.token = config.GitLab.Token

	projects, err := gitlab.fetchGroupProjects()
	printErrorThenExit(err, "")

	projectID := -1
	// Get GitLab group's project by ID
	for _, project := range projects {
		if project.Name == config.GitLab.Project {
			projectID = project.ID
		}
	}

	if projectID == -1 {
		printErrorThenExit(errors.New(""), "")
	}

	gitlab.fetchMergeRequestsByID(projectID, "?state=opened")

	for {
		time.AfterFunc(time.Duration(config.Watchdog.Duration)*time.Minute, func() {
			// do sth...
		})
	}
}
