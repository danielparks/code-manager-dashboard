package command

import (
	"encoding/json"
	"github.com/danielparks/code-manager-dashboard/codemanager"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
)

func init() {
	getfileCommand.PersistentFlags().StringP("state-file", "f", "", "File to store state in.")
	getfileCommand.PersistentFlags().BoolP("show", "S", false, "Show state.")
	RootCommand.AddCommand(getfileCommand)
}

var getfileCommand = &cobra.Command{
	Use:   "getfile",
	Short: "Load current state from a file",
	Run:   func(command *cobra.Command, args []string) {
		stateFile := getFlagString(command, "state-file")
		show := getFlagBool(command, "show")

		var codeState codemanager.CodeState
		var err error

		if stateFile != "" {
			codeState, err = loadOptionalCodeState(stateFile)
			if err != nil {
				log.Fatal(err)
			}
		}

		for _, source := range args {
			codeState.UpdateFromRawCodeState(loadRawCodeState(source))
		}

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


// Get deploy status from file
func loadRawCodeState(source string) map[string]interface{} {
	codeStateJson, err := ioutil.ReadFile(source)
	if err != nil {
		log.Fatal(err)
	}

	rawCodeState := map[string]interface{}{}
	err = json.Unmarshal(codeStateJson, &rawCodeState)
	if err != nil {
		log.Fatal(err)
	}

	return rawCodeState
}
