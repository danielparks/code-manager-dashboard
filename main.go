package main

import (
	"fmt"
	"github.com/danielparks/code-manager-dashboard/command"
	"os"
)

func main() {
	err := command.RootCommand.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
