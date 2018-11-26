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
		Host    string `yaml:"host"`
		Owner   string `yaml:"owner"`
		Project string `yaml:"project"`
		Token   string `yaml:"token"`
	} `yaml:"GitLab"`
	Watchdog struct {
		Duration int `yaml:"duration"`
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
	case config.GitLab.Owner == "":
		err = errors.New("GitLab owner is required")
	case config.GitLab.Project == "":
		err = errors.New("GitLab project is required")
	case config.GitLab.Token == "":
		err = errors.New("GitLab token is required")
		// case config.Watchdog.Duration == nil:
		// 	err = errors.New("GitLab token is required")
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
	host  string
	token string
	owner string
}

// GitLabProjectsResponse for GitLab API response of `group`
type GitLabProjectsResponse struct {
	Projects []GitLabProject `json:"projects"`
}

// GitLabProject for project structure in GitLabGroupResponse
type GitLabProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GitLabMergeRequest for GitLab merge requests structure
type GitLabMergeRequest struct {
	IID       int    `json:"iid"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	WIP       string `json:"work_in_progress"`
	WebURL    string `json:"web_url"`
	Author    struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"author"`
}

// Fetch GitLab merge requests by ID
func (utility *GitLabUtility) fetchMergeRequestsByID(id int, params string) ([]GitLabMergeRequest, error) {
	apiURL := utility.host + "/api/v4/projects/" + fmt.Sprint(id) + "/merge_requests/" + params
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

// Fetch GitLab projects
func (utility *GitLabUtility) fetchProjectIDByName(isByGroup bool, name string) (int, error) {
	if isByGroup {
		projects, _ := utility.fetchGroupProjects()

		// Get GitLab group's project by ID
		for _, project := range projects {
			if project.Name == name {
				return project.ID, nil
			}
		}

		return -1, errors.New("")
	}

	userID, err := utility.fetchUserIDByUsername()
	projectID, err := utility.fetchProjectIDByUserIDAndProjectName(userID, name)

	if err != nil {
		return -1, nil
	}

	return projectID, nil
}

// Fetch GitLab group's projects
func (utility *GitLabUtility) fetchGroupProjects() ([]GitLabProject, error) {

	apiURL := utility.host + "/api/v4/groups/" + utility.owner
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
		var model GitLabProjectsResponse

		json.Unmarshal(body, &model)

		return model.Projects, nil
	case 404:
		return []GitLabProject{}, errors.New("----")
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return []GitLabProject{}, errors.New("Unknown: " + string(body))
	}
}

// Fetch GitLab user ID by username
func (utility *GitLabUtility) fetchUserIDByUsername() (int, error) {
	type GitLabUserResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	apiURL := utility.host + "/api/v4/users?username=" + utility.owner
	request, _ := http.NewRequest("GET", apiURL, nil)
	request.Header.Set("Private-Token", utility.token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return -1, err
	}

	defer response.Body.Close()

	// Print info when success or failure
	switch response.StatusCode {
	case 200:
		body, _ := ioutil.ReadAll(response.Body)
		var models []GitLabUserResponse

		json.Unmarshal(body, &models)

		if len(models) == 1 {
			return models[0].ID, nil
		}

		return -1, errors.New("----")
	case 404:
		return -1, errors.New("----")
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return -1, errors.New("Unknown: " + string(body))
	}
}

func (utility *GitLabUtility) fetchProjectIDByUserIDAndProjectName(userID int, projectName string) (int, error) {
	apiURL := utility.host + "/api/v4/users/" + fmt.Sprint(userID) + "/projects?search=" + projectName
	request, _ := http.NewRequest("GET", apiURL, nil)
	request.Header.Set("Private-Token", utility.token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return -1, err
	}

	defer response.Body.Close()

	// Print info when success or failure
	switch response.StatusCode {
	case 200:
		body, _ := ioutil.ReadAll(response.Body)
		var models []GitLabProject

		json.Unmarshal(body, &models)

		for _, model := range models {
			if model.Name == projectName {
				return model.ID, nil
			}
		}

		return -1, errors.New("----")
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return -1, errors.New("Unknown: " + string(body))
	}
}

func timeDuring(start string, end string, format string) {

}

func main() {
	// Read config file path from command line
	var configFilePath = flag.String("path", "config.yml", "Setup your configuration file path.")
	var isByGroup = flag.Bool("group", true, "Setup project owner type (owned by a group or user).")
	flag.Parse()

	// Read & validate config.yml
	var config WatchdogConfig
	config.read(*configFilePath)
	config.validate()

	gitlab := GitLabUtility{config.GitLab.Host, config.GitLab.Token, config.GitLab.Owner}

	projectID, err := gitlab.fetchProjectIDByName(*isByGroup, config.GitLab.Project)
	printErrorThenExit(err, "")

	ticker := time.NewTicker(time.Duration(config.Watchdog.Duration) * time.Second)

	num := 0
	go func() {
		for {
			<-ticker.C
			num++
			fmt.Println("No.", num)

			mergeRequests, _ := gitlab.fetchMergeRequestsByID(projectID, "?state=opened")

			for _, mergeRequest := range mergeRequests {
				username := mergeRequest.Author.Username
				createdAt := mergeRequest.CreatedAt
				updatedAt := mergeRequest.UpdatedAt
				fmt.Println(username, createdAt, updatedAt)
			}
		}
	}()

	// For infinite goroutine
	for {
	}
}
