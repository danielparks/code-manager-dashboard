package command

import (
	"github.com/danielparks/code-manager-dashboard/web"
	"github.com/spf13/cobra"
)

func init() {
	serveCommand.PersistentFlags().StringP("state-file", "f", "",
		"File to store state in.")
	serveCommand.MarkPersistentFlagRequired("state-file")
	serveCommand.PersistentFlags().StringP("listen-on", "l", "localhost:8080",
		"[ADDRESS]:PORT to listen on.")
	RootCommand.AddCommand(serveCommand)
}

var serveCommand = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server",
	Args:  cobra.NoArgs,
	Run: func(command *cobra.Command, args []string) {
		web.Serve(
			getFlagString(command, "listen-on"),
			getFlagString(command, "state-file"),
		)
	},
}
