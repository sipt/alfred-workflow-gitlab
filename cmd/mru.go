package cmd

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type Item struct {
	URL    string `json:"url,omitempty"`
	Latest int64  `json:"latest,omitempty"`
}

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "record",
		Short: "record project",
		Run: func(cmd *cobra.Command, args []string) {
			wf.Run(func() {
				url := os.Getenv("repo")
				items := make([]Item, 0)
				err := wf.Cache.LoadJSON("mru", &items)
				if err != nil {
					if !strings.Contains(err.Error(), "no such file or directory") {
						log.Println("set [mru] failed.", err.Error())
						return
					}
				}
				found := false
				for _, item := range items {
					if item.URL == url {
						item.Latest = time.Now().Unix()
						found = true
						break
					}
				}
				if !found {
					items = append(items, Item{
						URL:    url,
						Latest: time.Now().Unix(),
					})
				}
				err = wf.Cache.StoreJSON("mru", items)
				if err != nil {
					log.Println("set [mru] failed.", err.Error())
					return
				}
				if err := valid(); err != nil {
					return
				}
			})
		},
	})
}
