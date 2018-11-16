package main

import (
	"encoding/json"
	"fmt"
	"github.com/pborman/getopt/v2"
	"io/ioutil"
	"log"
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
		date = date.Truncate(time.Second)
		if err != nil {
			log.Fatal(err)
		}
	}

	return Deploy{name, status, sha, date, nil}
}

func jsonGetArray(parent map[string]interface{}, key string) []interface{} {
	return parent[key].([]interface{})
}

func jsonGetObject(parent map[string]interface{}, key string) map[string]interface{} {
	return parent[key].(map[string]interface{})
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
	deployedEnvironments := jsonGetArray(fileSyncStatus, "deployed")
	for _, _rawDeploy := range deployedEnvironments {
		rawDeploy := _rawDeploy.(map[string]interface{})
		deploy := convertRawDeploy(rawDeploy, Deployed, "date")
		environments[deploy.name] = append(environments[deploy.name], deploy)
	}

	deploysStatus := jsonGetObject(object, "deploys-status")
//	  "queued": [],
//    "deploying": [],
//    "new": [],
//    "failed": []


	queuedEnvironments := jsonGetArray(deploysStatus, "queued")
	for _, _rawDeploy := range queuedEnvironments {
		rawDeploy := _rawDeploy.(map[string]interface{})
		deploy := convertRawDeploy(rawDeploy, Queued, "queued-at")
		environments[deploy.name] = append(environments[deploy.name], deploy)
	}

	now := time.Now().Truncate(time.Second)
	localZone, localZoneOffset := now.Zone()
	location := time.FixedZone(localZone, localZoneOffset)

	for environment, deploys := range environments {
		for _, deploy := range deploys {
			localDate := deploy.date.In(location)

			fmt.Printf("%-45s  %-9s  %s  %v\n", environment, deploy.status, localDate, deploy.date.Sub(now))
			environment = ""
		}
	}
}
