package command

import (
	"github.com/danielparks/code-manager-dashboard/codemanager"
	log "github.com/sirupsen/logrus"
  "github.com/spf13/cobra"
  "os"
)

var (
	stateFile = getapiCommand.PersistentFlags().StringP("state-file", "f", "", "File to store state in.")
	show = getapiCommand.PersistentFlags().BoolP("show", "S", false, "Show state.")
)

func init() {
  RootCommand.AddCommand(getapiCommand)
}

var getapiCommand = &cobra.Command{
  Use:   "getapi",
  Short: "Load current state from the Code Manager API.",
  Run:   func(command *cobra.Command, args []string) {
		server := "pe-mom1-prod.ops.puppetlabs.net"
		caPath := "/Users/daniel/work/puppetca.ops.puppetlabs.net.pem"

		var codeState codemanager.CodeState
		var err error

		if *stateFile != "" {
			codeState, err = codemanager.LoadCodeState(*stateFile)
			if err != nil {
				log.Fatal(err)
			}
		}

		apiClient := codemanager.TypicalApiClient(server, os.Getenv("pe_token"), caPath)
		rawCodeState := apiClient.GetRawCodeState()
		codeState.UpdateFromRawCodeState(rawCodeState)

		if *show {
			ShowEnvironments(&codeState)
		}

		if *stateFile != "" {
			err = codemanager.SaveCodeState(&codeState, *stateFile)
			if err != nil {
				log.Fatal(err)
			}
		}
  },
}
