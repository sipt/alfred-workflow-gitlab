package cmd

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"

	aw "github.com/deanishe/awgo"
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
			wf.SendFeedback()
		},
	}, &cobra.Command{
		Use:   "setprojects",
		Short: "download projects",
		Run: func(cmd *cobra.Command, args []string) {
			wf.Run(DownloadProjects)
			wf.SendFeedback()
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
	query = strings.Replace(strings.ToLower(query), " ", "", -1)
	filtered := make([]*Project, 0)
	for _, project := range projects {
		if query == "" || strings.Contains(strings.Replace(strings.ToLower(project.Name), " ", "", -1), query) {
			filtered = append(filtered, project)
		}
	}
	items := make([]Item, 0)
	_ = wf.Cache.LoadJSON("mru", &items)
	itemMap := make(map[string]Item)
	for _, item := range items {
		itemMap[item.URL] = item
	}
	sort.Slice(filtered, func(i, j int) bool {
		if itemMap[filtered[i].WebUrl].Latest > itemMap[filtered[j].WebUrl].Latest {
			return true
		} else if itemMap[filtered[i].WebUrl].Latest == itemMap[filtered[j].WebUrl].Latest {
			return filtered[i].Name < filtered[j].Name
		}
		return false
	})
	for _, project := range filtered {
		wf.NewItem(project.Name).Arg(project.WebUrl).Valid(true).Icon(&aw.Icon{
			Value: "icon.png",
		})
		count++
		if count >= 20 {
			break
		}
	}
}

func DownloadProjectsInBg() {
	if wf.IsRunning("download_projects") {
		return
	}
	execCmd := exec.Command("./awgitlab", "setprojects")
	err := wf.RunInBackground("download_projects", execCmd)
	if err != nil {
		log.Println("run download_projects job failed.", err)
	}
}

func DownloadProjects() {
	if err := valid(); err != nil {
		return
	}
	git, err := gitlab.NewClient(accessToken, gitlab.WithBaseURL(fmt.Sprintf("https://%s/api/v4", domain)))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	projects, resp, err := fetchProjects(git, 0)
	if err != nil {
		log.Fatalf("Failed to fetchProjects: %v", err)
	}
	results := make(chan []*Project, resp.TotalPages)
	wg := &sync.WaitGroup{}
	for i := resp.NextPage; i <= resp.TotalPages; i++ {
		wg.Add(1)
		go func(page int) {
			projects, _, err := fetchProjects(git, page)
			if err != nil {
				log.Fatalf("Failed to fetchProjects: %v", err)
			}
			results <- projects
			wg.Done()
		}(i)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	for result := range results {
		projects = append(projects, result...)
	}
	err = wf.Cache.StoreJSON("projects", projects)
	if err != nil {
		log.Println("store projects failed.", err.Error())
	}

	return
}

func fetchProjects(git *gitlab.Client, page int) ([]*Project, *gitlab.Response, error) {
	var simple = true
	list, resp, err := git.Projects.ListProjects(&gitlab.ListProjectsOptions{ListOptions: gitlab.ListOptions{Page: page, PerPage: 100}, Simple: &simple})
	if err != nil {
		log.Println("Get projects failed.", err.Error())
		return nil, nil, err
	}
	projects := make([]*Project, len(list))
	for i, project := range list {
		projects[i] = &Project{
			Name:   project.NameWithNamespace,
			Desc:   project.Description,
			WebUrl: project.WebURL,
		}
	}
	return projects, resp, nil
}

func valid() error {
	data, err := wf.Cache.Load("domain")
	if err != nil || len(data) == 0 {
		log.Println("[domain] not found.", err)
		wf.NewWarningItem("Please set GitLab Domain.", "").Icon(&aw.Icon{
			Value: "icon.png",
		})
		return errors.New("[domain] not found")
	}
	domain = string(data)
	data, err = wf.Cache.Load("access_token")
	if err != nil || len(data) == 0 {
		log.Println("[access_token] not found.")
		wf.NewWarningItem("Please set GitLab AccessToken.", "").Icon(&aw.Icon{
			Value: "icon.png",
		})
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
