package main

import (
	"fmt"
	"github.com/danielparks/code-manager-dashboard/command"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

func main() {
	err := command.RootCommand.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Stderr.WriteString(err)
		os.Exit(1)
	}
}
