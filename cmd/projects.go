package cmd

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "projects",
		Short: "search projects",
		Run: func(cmd *cobra.Command, args []string) {
			wf.Run(func() {
				var query = ""
				if len(args) > 0 {
					query = args[0]
				}
				run(query)
			})
		},
	}, &cobra.Command{
		Use:   "setprojects",
		Short: "download projects",
		Run: func(cmd *cobra.Command, args []string) {
			wf.Run(DownloadProjects)
		},
	})
}

func run(query string) {
	if err := valid(); err != nil {
		return
	}
	projects := make([]*Project, 0)
	err := wf.Cache.LoadJSON("projects", &projects)
	if err != nil {
		DownloadProjectsInBg()
		log.Println("load projects failed.", err)
		wf.NewItem("Loading ...")
		return
	}
	age, _ := wf.Cache.Age("projects")
	if age > time.Minute*10 {
		go DownloadProjectsInBg()
	}
	var count = 0
	query = strings.ToLower(query)
	for _, project := range projects {
		if query == "" || strings.Contains(strings.ToLower(project.Name), query) {
			wf.NewItem(project.Name).Arg(project.WebUrl).Valid(true)
			count++
			if count >= 20 {
				break
			}
		}
	}
}

func DownloadProjectsInBg() {
	if wf.IsRunning("download_projects") {
		return
	}
	execCmd := exec.Command("awgitlab", "setprojects")
	err := wf.RunInBackground("download_projects", execCmd)
	if err != nil {
		log.Println("run download_projects job failed.", err)
	}
}

func DownloadProjects() {
	git, err := gitlab.NewClient(accessToken, gitlab.WithBaseURL(fmt.Sprintf("https://%s/api/v4", domain)))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	var simple = true
	list, _, err := git.Projects.ListProjects(&gitlab.ListProjectsOptions{Simple: &simple})
	if err != nil {
		log.Println("Get projects failed.", err.Error())
		return
	}
	projects := make([]*Project, len(list))
	for i, project := range list {
		projects[i] = &Project{
			Name:   project.NameWithNamespace,
			Desc:   project.Description,
			WebUrl: project.WebURL,
		}
	}
	err = wf.Cache.StoreJSON("projects", projects)
	if err != nil {
		log.Println("store projects failed.", err.Error())
	}
	return
}

func valid() error {
	data, err := wf.Cache.Load("domain")
	if err != nil || len(data) == 0 {
		log.Println("[domain] not found.")
		wf.NewWarningItem("Please set GitLab Domain.", "")
		return errors.New("[domain] not found")
	}
	domain = string(data)
	data, err = wf.Cache.Load("access_token")
	if err != nil || len(data) == 0 {
		log.Println("[access_token] not found.")
		wf.NewWarningItem("Please set GitLab AccessToken.", "")
		return errors.New("[access_token] not found")
	}
	accessToken = string(data)
	return nil
}

type Project struct {
	Name   string `json:"name,omitempty"`
	Desc   string `json:"desc,omitempty"`
	WebUrl string `json:"web_url,omitempty"`
}
