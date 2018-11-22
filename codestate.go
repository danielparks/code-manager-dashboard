package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

type Deploy struct {
	Environment string
	Status      DeployStatus
	Sha         string
	Date        time.Time
	Error       JsonObject
}

type EnvironmentState struct {
	Environment string
	Deploys     []Deploy
}

type CodeState struct {
	Environments map[string]EnvironmentState
}

type JsonObject map[string]interface{}

func (codeState *CodeState) UpdateFromRawCodeState(rawCodeState JsonObject) {
	var rawDeploys []interface{}
	environmentsSeen := map[string]bool{}

	codeState.ClearUnfinishedDeploys()

	fileSyncStatus := rawCodeState.GetObject("file-sync-storage-status")
	deploysStatus := rawCodeState.GetObject("deploys-status")

	rawDeploys = deploysStatus.GetArray("new")
	codeState.convertRawDeploys(rawDeploys, New, "queued-at", environmentsSeen)

	rawDeploys = deploysStatus.GetArray("queued")
	codeState.convertRawDeploys(rawDeploys, Queued, "queued-at", environmentsSeen)

	rawDeploys = deploysStatus.GetArray("deploying")
	codeState.convertRawDeploys(rawDeploys, Deploying, "queued-at", environmentsSeen)

	rawDeploys = fileSyncStatus.GetArray("deployed")
	codeState.convertRawDeploys(rawDeploys, Deployed, "date", environmentsSeen)

	rawDeploys = deploysStatus.GetArray("failed")
	codeState.convertRawDeploys(rawDeploys, Failed, "queued-at", environmentsSeen)

	for name, environmentState := range codeState.Environments {
		if !environmentsSeen[name] && environmentState.Deploys[0].Status != Deleted {
			// This environment is wasn't in the current update, and its last recorded
			// status isn't Deleted. So, it needs a Deleted record.
			environmentState.AddDeploy(Deploy{name, Deleted, "", time.Now(), nil})
		}

		environmentState.RemoveDuplicateDeploys()
		environmentState.SortDeploys()

		codeState.Environments[name] = environmentState
	}
}

func (codeState *CodeState) ClearUnfinishedDeploys() {
	for _, environmentState := range codeState.Environments {
		environmentState.ClearUnfinishedDeploys()
	}
}

func (environmentState *EnvironmentState) ClearUnfinishedDeploys() {
	cleanedDeploys := []Deploy{}
	for _, deploy := range environmentState.Deploys {
		if deploy.Status >= Deployed {
			// Either Deployed or Failed.
			cleanedDeploys = append(cleanedDeploys, deploy)
		}
	}
	environmentState.Deploys = cleanedDeploys
}

func (environmentState *EnvironmentState) RemoveDuplicateDeploys() {
	uniqueDeploys := []Deploy{}

	seen := map[string]bool{}
	for _, deploy := range environmentState.Deploys {
		key := fmt.Sprintf("%s %s %s", deploy.Status, deploy.Sha, deploy.Date)
		if seen[key] {
			continue
		}

		seen[key] = true
		uniqueDeploys = append(uniqueDeploys, deploy)
	}

	environmentState.Deploys = uniqueDeploys
}

func (environmentState *EnvironmentState) SortDeploys() {
	sort.Slice(environmentState.Deploys, func(i, j int) bool {
		a := environmentState.Deploys[i]
		b := environmentState.Deploys[j]

		if a.Status >= Deployed && b.Status >= Deployed {
			// Either Deployed or Failed. These should be sorted together by date.
			return a.Date.After(b.Date)
		} else if a.Status == b.Status {
			// Same status, so sort on date.
			return a.Date.After(b.Date)
		} else {
			return b.Status > a.Status
		}
	})
}

func convertRawDeploy(rawDeploy JsonObject, status DeployStatus, dateKey string) Deploy {
	var sha string
	var date time.Time
	var err error

	environment := rawDeploy["environment"].(string)
	if rawDeploy["deploy-signature"] != nil {
		sha = rawDeploy["deploy-signature"].(string)
	}

	if rawDeploy[dateKey] != nil {
		date, err = time.Parse(RFC3339Micro, rawDeploy[dateKey].(string))
		if err != nil {
			log.Fatal(err)
		}
	}

	var deployError JsonObject

	if rawDeploy["error"] != nil {
		deployError = rawDeploy.GetObject("error")

		if status == Failed && deployError != nil && deployError["msg"] != nil && strings.Contains(deployError["msg"].(string), "cannot be found in any source and will not be deployed.") {
			// Check for Environment(s) 'combined_minor_changes' cannot be found in any source and will not be deployed.
			// if it's in the error.msg, then convert this to Delete
			status = Deleted
		}
	}

	return Deploy{environment, status, sha, date, deployError}
}

func (codeState *CodeState) convertRawDeploys(rawDeploys []interface{}, status DeployStatus, dateKey string, environmentsSeen map[string]bool) {
	for _, _rawDeploy := range rawDeploys {
		rawDeploy := JsonObject(_rawDeploy.(map[string]interface{}))
		deploy := convertRawDeploy(rawDeploy, status, dateKey)

		codeState.AddDeploy(deploy)
		environmentsSeen[deploy.Environment] = true
	}
}

func (environmentState *EnvironmentState) AddDeploy(deploy Deploy) {
	environmentState.Deploys = append(environmentState.Deploys, deploy)
}

func (codeState *CodeState) AddDeploy(deploy Deploy) {
	if codeState.Environments == nil {
		codeState.Environments = map[string]EnvironmentState{}
	}

	environmentState := codeState.Environments[deploy.Environment]
	if environmentState.Environment == "" {
		// We haven't seen this environment before; initialize it.
		environmentState.Environment = deploy.Environment
	}

	environmentState.AddDeploy(deploy)

	codeState.Environments[deploy.Environment] = environmentState
}

func (parent JsonObject) GetArray(key string) []interface{} {
	return parent[key].([]interface{})
}

func (parent JsonObject) GetObject(key string) JsonObject {
	return JsonObject(parent[key].(map[string]interface{}))
}
