package command

import (
	"fmt"
	"github.com/danielparks/code-manager-dashboard/codemanager"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	showCommand.PersistentFlags().StringP("state-file", "f", "", "File to store state in.")
	showCommand.MarkPersistentFlagRequired("state-file")
	RootCommand.AddCommand(showCommand)
}

var showCommand = &cobra.Command{
	Use:   "show",
	Short: "Show deployment status recorded in the state file",
	Args:  cobra.ArbitraryArgs,
	Run:   func(command *cobra.Command, args []string) {
		stateFile := getFlagString(command, "state-file")

		codeState, err := codemanager.LoadCodeState(stateFile)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			ShowEnvironments(&codeState)
		} else {
			for _, name := range args {
				environmentState := codeState.Environments[name]
				ShowEnvironmentState(&environmentState)
			}
		}
	},
}

func getLocation() *time.Location {
	now := time.Now().Truncate(time.Second)
	localZone, localZoneOffset := now.Zone()
	return time.FixedZone(localZone, localZoneOffset)
}

// FIXME: use ShowEnvironmentState? rename?
func ShowEnvironments(codeState *codemanager.CodeState) {
	environments := codeState.SortedEnvironments()
	location := getLocation()

	for _, environmentState := range environments {
		environment := environmentState.Environment

		environmentState.SortDeploys(codemanager.Descending)
		for _, deploy := range environmentState.Deploys {
			localDate := deploy.MatchTime().Truncate(time.Second).In(location)
			fmt.Printf("%-45s  %-9s  %s\n", environment, deploy.Status, localDate)
			environment = ""
		}
	}
}

// FIXME: how do we handle non-existent environments?
func ShowEnvironmentState(environmentState *codemanager.EnvironmentState) {
	environment := environmentState.Environment
	location := getLocation()

	environmentState.SortDeploys(codemanager.Descending)
	for _, deploy := range environmentState.Deploys {
		localDate := deploy.MatchTime().Truncate(time.Second).In(location)
		fmt.Printf("%-45s  %-9s  %s\n", environment, deploy.Status, localDate)
		environment = ""
	}
}
