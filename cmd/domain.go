package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "domain",
		Short: "set domain",
		Run: func(cmd *cobra.Command, args []string) {
			wf.Run(func() {
				if len(args) == 0 || len(args[0]) == 0 {
					wf.NewWarningItem("Invalid domain", "")
					return
				}
				err := wf.Cache.Store("domain", []byte(args[0]))
				if err != nil {
					log.Println("set [domain] failed.", err.Error())
					wf.NewWarningItem("Set GitLab Domain failed", err.Error())
					return
				}
			})
		},
	})
}
