package cmd

import (
	"fmt"
	"os"

	aw "github.com/deanishe/awgo"
	"github.com/spf13/cobra"
)

var wf *aw.Workflow
var domain string
var accessToken string

func init() {
	// Create a new Workflow using default settings.
	// Critical settings are provided by Alfred via environment variables,
	// so this *will* die in flames if not run in an Alfred-like environment.
	wf = aw.New()
}

var rootCmd = &cobra.Command{
	Use:   "gitlab",
	Short: "alfred workflow gitlab",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
