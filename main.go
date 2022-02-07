package main

// Package is called aw
import (
	"github.com/TTNomi/alfred-workflow-gitlab/cmd"
	aw "github.com/deanishe/awgo"
)

// Workflow is the main API
var wf *aw.Workflow
var domain string
var accessToken string

func init() {
	// Create a new Workflow using default settings.
	// Critical settings are provided by Alfred via environment variables,
	// so this *will* die in flames if not run in an Alfred-like environment.
	wf = aw.New()
}

func main() {
	cmd.Execute()
}
