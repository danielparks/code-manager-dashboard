package command

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	verbose = RootCommand.PersistentFlags().BoolP("verbose", "v", false, "Output more information.")
	debug = RootCommand.PersistentFlags().BoolP("debug", "d", false, "Output debugging information.")
	trace = RootCommand.PersistentFlags().Bool("trace", false, "Output trace information (more than debug).")
)

var RootCommand = &cobra.Command{
	Use:   "code-manager-dashboard",
	Short: "Dashboard for Code Manager deploys",
	Long:  "Dashboard showing the deployment status of all environments",
	Run:   func(command *cobra.Command, args []string) {
		if *trace {
			log.SetLevel(log.TraceLevel)
		} else if *debug {
			log.SetLevel(log.DebugLevel)
		} else if *verbose {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.WarnLevel)
		}
	},
}
