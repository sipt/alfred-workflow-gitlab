package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "accesstoken",
		Short: "set access_token",
		Run: func(cmd *cobra.Command, args []string) {
			wf.Run(func() {
				if len(args) == 0 || len(args[0]) == 0 {
					wf.NewWarningItem("Invalid domain", "")
					return
				}
				err := wf.Cache.Store("access_token", []byte(args[0]))
				if err != nil {
					log.Println("set [access_token] failed.", err.Error())
					wf.NewWarningItem("Set GitLab AccessToken failed", err.Error())
					return
				}
			})
		},
	})
}
