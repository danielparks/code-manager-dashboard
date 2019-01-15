package command

import (
	"github.com/danielparks/code-manager-dashboard/codemanager"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"fmt"
)

func init() {
	trimCommand.PersistentFlags().StringP("state-file", "f", "", "File to store state in.")
	trimCommand.MarkPersistentFlagRequired("state-file")
	trimCommand.PersistentFlags().IntP("count", "c", 5, "Number of deploys to keep for each environment.")
	trimCommand.MarkPersistentFlagRequired("count")
	trimCommand.PersistentFlags().BoolP("show", "S", false, "Show state.")
	RootCommand.AddCommand(trimCommand)
}

var trimCommand = &cobra.Command{
	Use:   "trim",
	Short: "Trim deploys down to a specified size (per environment)",
	Args:  cobra.NoArgs,
	Run: func(command *cobra.Command, args []string) {
		stateFile := getFlagString(command, "state-file")
		count := getFlagInt(command, "count")
		show := getFlagBool(command, "show")

		codeState, err := codemanager.LoadCodeState(stateFile)
		if err != nil {
			log.Fatal(err)
		}

		TrimEnvironments(&codeState, count)

		if show {
			ShowEnvironments(&codeState)
		}

		err = codemanager.SaveCodeState(&codeState, stateFile)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func TrimEnvironments(codeState *codemanager.CodeState, count int) {
	for i, environmentState := range codeState.Environments {
		environmentState.SortDeploys(codemanager.Descending)
		if len(environmentState.Deploys) > count {
			environmentState.Deploys = environmentState.Deploys[0:count]
			fmt.Printf("env %s: %d\n", environmentState.Environment, len(environmentState.Deploys))
		}
		codeState.Environments[i] = environmentState
	}
}
