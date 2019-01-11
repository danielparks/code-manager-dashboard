package command

import (
	"github.com/danielparks/code-manager-dashboard/codemanager"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	RootCommand.PersistentFlags().BoolP("verbose", "v", false,
		"Output more information")
	RootCommand.PersistentFlags().BoolP("debug", "d", false,
		"Output debugging information")
	RootCommand.PersistentFlags().Bool("trace", false,
		"Output trace information (more than debug)")
}

func getFlagBool(command *cobra.Command, name string) bool {
	value, err := command.Flags().GetBool(name)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func getFlagInt(command *cobra.Command, name string) int {
	value, err := command.Flags().GetInt(name)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func getFlagString(command *cobra.Command, name string) string {
	value, err := command.Flags().GetString(name)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func loadOptionalCodeState(path string) (codemanager.CodeState, error) {
	state, err := codemanager.LoadCodeState(path)
	if os.IsNotExist(err) {
		return state, nil
	}

	return state, err
}

var RootCommand = &cobra.Command{
	Use:   "code-manager-dashboard",
	Short: "Dashboard for Code Manager deploys",
	Long:  "Dashboard showing the deployment status of all environments",
	PersistentPreRun: func(command *cobra.Command, args []string) {
		if getFlagBool(command, "trace") {
			log.SetLevel(log.TraceLevel)
		} else if getFlagBool(command, "debug") {
			log.SetLevel(log.DebugLevel)
		} else if getFlagBool(command, "verbose") {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.WarnLevel)
		}
	},
}
