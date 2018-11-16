package main

import (
	"encoding/json"
	"fmt"
	"github.com/pborman/getopt/v2"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"
)

const RFC3339Micro = "2006-01-02T15:04:05.999Z07:00"

type DeployStatus int

const (
	New       DeployStatus = iota
	Queued    DeployStatus = iota
	Deploying DeployStatus = iota
	Deployed  DeployStatus = iota
	Failed    DeployStatus = iota
)

func (status DeployStatus) String() string {
	names := [...]string{
		"new",
		"queued",
		"deploying",
		"deployed",
		"failed",
	}
	return names[status]
}

type Deploy struct {
	name   string
	status DeployStatus
	sha    string
	date   time.Time
	error  map[string]interface{}
}

func convertRawDeploy(rawDeploy map[string]interface{}, status DeployStatus, dateKey string) Deploy {
	var sha string
	var date time.Time
	var err error

	name := rawDeploy["environment"].(string)
	if rawDeploy["deploy-signature"] != nil {
		sha = rawDeploy["deploy-signature"].(string)
	}

	if rawDeploy[dateKey] != nil {
		date, err = time.Parse(RFC3339Micro, rawDeploy[dateKey].(string))
		if err != nil {
			log.Fatal(err)
		}
	}

	return Deploy{name, status, sha, date, nil}
}

func convertRawDeploys(deploys *map[string][]Deploy, rawDeploys []interface{}, status DeployStatus, dateKey string) {
	for _, _rawDeploy := range rawDeploys {
		rawDeploy := _rawDeploy.(map[string]interface{})
		deploy := convertRawDeploy(rawDeploy, status, dateKey)
		(*deploys)[deploy.name] = append((*deploys)[deploy.name], deploy)
	}
}

func jsonGetArray(parent map[string]interface{}, key string) []interface{} {
	return parent[key].([]interface{})
}

func jsonGetObject(parent map[string]interface{}, key string) map[string]interface{} {
	return parent[key].(map[string]interface{})
}

func environmentsValues(theMap map[string][]Deploy) [][]Deploy {
	values := make([][]Deploy, len(theMap))
	i := 0
	for _, value := range theMap {
		values[i] = value
		i++
	}
	return values
}

func main() {
	var deployStatusResponse []byte
	var err error

	deployStatusSource := getopt.StringLong("status-source", 'S', "",
		"File to use instead of deploy status API endpoint")
	getopt.Parse()

	if *deployStatusSource == "" {
		deployStatusResponse = GetDeployStatus()
	} else {
		deployStatusResponse, err = ioutil.ReadFile(*deployStatusSource)
		if err != nil {
			log.Fatal(err)
		}
	}

	environments := map[string][]Deploy{}

	object := map[string]interface{}{}
	json.Unmarshal(deployStatusResponse, &object)

	fileSyncStatus := jsonGetObject(object, "file-sync-storage-status")
	rawDeploys := jsonGetArray(fileSyncStatus, "deployed")
	convertRawDeploys(&environments, rawDeploys, Deployed, "date")

	deploysStatus := jsonGetObject(object, "deploys-status")
	rawDeploys = jsonGetArray(deploysStatus, "deploying")
	convertRawDeploys(&environments, rawDeploys, Deploying, "queued-at")

	rawDeploys = jsonGetArray(deploysStatus, "queued")
	convertRawDeploys(&environments, rawDeploys, Queued, "queued-at")

	rawDeploys = jsonGetArray(deploysStatus, "new")
	convertRawDeploys(&environments, rawDeploys, New, "queued-at")

	rawDeploys = jsonGetArray(deploysStatus, "failed")
	convertRawDeploys(&environments, rawDeploys, Failed, "queued-at")

	sortedEnvironments := environmentsValues(environments)
	sort.Slice(sortedEnvironments, func(i, j int) bool {
		a := sortedEnvironments[i]
		b := sortedEnvironments[j]
		return strings.ToLower(a[0].name) < strings.ToLower(b[0].name)
	})

	now := time.Now().Truncate(time.Second)
	localZone, localZoneOffset := now.Zone()
	location := time.FixedZone(localZone, localZoneOffset)

	for _, deploys := range sortedEnvironments {
		environment := deploys[0].name

		sort.Slice(deploys, func(i, j int) bool {
			a := deploys[i]
			b := deploys[j]
			if a.status >= Deployed && b.status >= Deployed {
				// Either Deployed or Failed. These should be sorted together by date.
				return a.date.After(b.date)
			} else if a.status == b.status {
				// Same status, so sort on date.
				return a.date.After(b.date)
			} else {
				return b.status > a.status
			}
		})

		for _, deploy := range deploys {
			localDate := deploy.date.Truncate(time.Second).In(location)
			elapsed := deploy.date.Truncate(time.Second).Sub(now)

			fmt.Printf("%-45s  %-9s  %s  %v\n", environment, deploy.status, localDate, elapsed)
			environment = ""
		}
	}
}
