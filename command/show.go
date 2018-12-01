package command

import (
	"github.com/danielparks/code-manager-dashboard/codemanager"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"fmt"
	"sort"
	"strings"
	"time"
)

var stateFile = showCommand.PersistentFlags().StringP("state-file", "f", "", "File to store state in.")

func init() {
	showCommand.MarkFlagRequired("state-file")
	RootCommand.AddCommand(showCommand)
}

var showCommand = &cobra.Command{
	Use:   "show",
	Short: "Show deployment status recorded in the state file.",
	Run:   func(command *cobra.Command, args []string) {
		codeState, err := codemanager.LoadCodeState(*stateFile)
		if err != nil {
			log.Fatal(err)
		}

		ShowEnvironments(&codeState)
	},
}

func sortedEnvironments(codeState *codemanager.CodeState) []codemanager.EnvironmentState {
	environments := make([]codemanager.EnvironmentState, len(codeState.Environments))
	i := 0
	for _, environmentState := range codeState.Environments {
		environments[i] = environmentState
		i++
	}

	sort.Slice(environments, func(i, j int) bool {
		a := environments[i].Deploys[0]
		b := environments[j].Deploys[0]
		return strings.ToLower(a.Environment) < strings.ToLower(b.Environment)
	})

	return environments
}

func ShowEnvironments(codeState *codemanager.CodeState) {
	environments := sortedEnvironments(codeState)

	now := time.Now().Truncate(time.Second)
	localZone, localZoneOffset := now.Zone()
	location := time.FixedZone(localZone, localZoneOffset)

	for _, environmentState := range environments {
		environment := environmentState.Environment

		environmentState.SortDeploys(codemanager.Ascending)
		for _, deploy := range environmentState.Deploys {
			localDate := deploy.MatchTime().Truncate(time.Second).In(location)
			fmt.Printf("%-45s  %-9s  %s\n", environment, deploy.Status, localDate)
			environment = ""
		}
	}
}
