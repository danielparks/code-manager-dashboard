package main

import (
	"encoding/json"
	"fmt"
	"github.com/pborman/getopt/v2"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

const RFC3339Micro = "2006-01-02T15:04:05.999Z07:00"

func sortedEnvironments(codeState *CodeState) []EnvironmentState {
	environments := make([]EnvironmentState, len(codeState.Environments))
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

func displayEnvironments(codeState *CodeState) {
	environments := sortedEnvironments(codeState)

	now := time.Now().Truncate(time.Second)
	localZone, localZoneOffset := now.Zone()
	location := time.FixedZone(localZone, localZoneOffset)

	for _, environmentState := range environments {
		environment := environmentState.Environment

		for _, deploy := range environmentState.Deploys {
			localDate := deploy.DisplayTime().Truncate(time.Second).In(location)
			elapsed := deploy.DisplayTime().Truncate(time.Second).Sub(now)

			fmt.Printf("%-45s  %-9s  %s  %v\n", environment, deploy.Status, localDate, elapsed)
			environment = ""
		}
	}
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

func readState(path string) (CodeState, error) {
	state := CodeState{}

	stateJson, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return state, nil
	}

	if err != nil {
		return state, err
	}

	err = json.Unmarshal(stateJson, &state)
	return state, err
}

func dumpState(codeState *CodeState, path string) error {
	stateJson, err := json.MarshalIndent(*codeState, "", "  ")
	if err != nil {
		return err
	}

	/// FIXME should we lock this?
	return ioutil.WriteFile(path, append(stateJson, '\n'), 0644)
}

func main() {
	var fakeStatus bool
	getopt.FlagLong(&fakeStatus, "fake-status", 0,
		"Treat arguments as list of files to load deploy statuses from.")

	var stateFile string
	getopt.FlagLong(&stateFile, "state-file", 's',
		"File to store state in.")

	getopt.Parse()
	args := getopt.Args()

	server := "pe-mom1-prod.ops.puppetlabs.net"
	caPath := "/Users/daniel/work/puppetca.ops.puppetlabs.net.pem"

	var codeState CodeState
	var err error

	if stateFile != "" {
		codeState, err = readState(stateFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	if fakeStatus {
		for _, source := range args {
			codeState.UpdateFromRawCodeState(loadRawCodeState(source))
		}
	} else {
		apiClient := TypicalApiClient(server, os.Getenv("pe_token"), caPath)
		rawCodeState := apiClient.GetRawCodeState()
		codeState.UpdateFromRawCodeState(rawCodeState)
	}

	displayEnvironments(&codeState)

	if stateFile != "" {
		err = dumpState(&codeState, stateFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}
