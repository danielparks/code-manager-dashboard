package command

import (
	"github.com/danielparks/code-manager-dashboard/codemanager"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	getapiCommand.PersistentFlags().StringP("state-file", "f", "", "File to store state in.")
	getapiCommand.PersistentFlags().BoolP("show", "S", false, "Show state.")
	RootCommand.AddCommand(getapiCommand)
}

var getapiCommand = &cobra.Command{
	Use:   "getapi",
	Short: "Load current state from the Code Manager API",
	Run:   func(command *cobra.Command, args []string) {
		stateFile := getFlagString(command, "state-file")
		show := getFlagBool(command, "show")

		server := "pe-mom1-prod.ops.puppetlabs.net"
		caPath := "/Users/daniel/work/puppetca.ops.puppetlabs.net.pem"

		var codeState codemanager.CodeState
		var err error

		if stateFile != "" {
			codeState, err = loadOptionalCodeState(stateFile)
			if err != nil {
				log.Fatal(err)
			}
		}

		apiClient := codemanager.TypicalApiClient(server, os.Getenv("pe_token"), caPath)
		rawCodeState := apiClient.GetRawCodeState()
		codeState.UpdateFromRawCodeState(rawCodeState)

		if show {
			ShowEnvironments(&codeState)
		}

		if stateFile != "" {
			err = codemanager.SaveCodeState(&codeState, stateFile)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}
