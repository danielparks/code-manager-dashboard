package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

type Deploy struct {
	Environment   string
	Status        DeployStatus
	Sha           string
	FinishedAt    time.Time
	QueuedAt      time.Time
	EstimatedTime time.Time
	Error         JsonObject
}

type EnvironmentState struct {
	Environment string
	Deploys     []Deploy
}

type CodeState struct {
	Environments map[string]EnvironmentState
}

func (codeState *CodeState) UpdateFromRawCodeState(rawCodeState JsonObject) {
	environmentsSeen := map[string]bool{}

	codeState.ClearUnfinishedDeploys()

	deploysStatus := rawCodeState.GetObject("deploys-status")
	mappings := map[string]DeployStatus{
		"new": New,
		"queued": Queued,
		"deploying": Deploying,
		"failed": Failed,
	}

	for key, status := range mappings {
		rawDeploys := deploysStatus.GetArray(key)
		codeState.addRawDeploys(rawDeploys, status, environmentsSeen)
	}

	fileSyncStatus := rawCodeState.GetObject("file-sync-storage-status")
	rawDeploys := fileSyncStatus.GetArray("deployed")
	codeState.addRawDeploys(rawDeploys, Deployed, environmentsSeen)

	for name, environmentState := range codeState.Environments {
		if !environmentsSeen[name] && environmentState.Deploys[0].Status != Deleted {
			// This environment is wasn't in the current update, and its last recorded
			// status isn't Deleted. So, it needs a Deleted record.
			environmentState.AddDeploy(
				Deploy{
					Environment:   name,
					Status:        Deleted,
					EstimatedTime: time.Now(),
				})
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
		key := fmt.Sprintf("%s %s %s", deploy.Status, deploy.Sha, deploy.Time())
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
			return a.Time().After(b.Time())
		} else if a.Status == b.Status {
			// Same status, so sort on date.
			return a.Time().After(b.Time())
		} else {
			return b.Status > a.Status
		}
	})
}

func convertRawDate(rawDate interface{}) time.Time {
	if rawDate == nil {
		return time.Time{}
	}

	date, err := time.Parse(RFC3339Micro, rawDate.(string))
	if err != nil {
		log.Fatal(err)
	}

	return date
}

func convertRawDeploy(rawDeploy JsonObject, status DeployStatus) Deploy {
	var deploy Deploy

	deploy.Environment = rawDeploy["environment"].(string)
	deploy.Status = status
	deploy.QueuedAt = convertRawDate(rawDeploy["queued-at"])
	deploy.FinishedAt = convertRawDate(rawDeploy["date"])

	if rawDeploy["deploy-signature"] != nil {
		deploy.Sha = rawDeploy["deploy-signature"].(string)
	}

	if rawDeploy["error"] != nil {
		deploy.Error = rawDeploy.GetObject("error")

		if deploy.Status == Failed && deploy.Error != nil && deploy.Error["msg"] != nil && strings.Contains(deploy.Error["msg"].(string), "cannot be found in any source and will not be deployed.") {
			// Check for Environment(s) 'combined_minor_changes' cannot be found in any source and will not be deployed.
			// if it's in the error.msg, then convert this to Delete
			deploy.Status = Deleted
		}
	}

	return deploy
}

func (codeState *CodeState) addRawDeploys(rawDeploys []interface{}, status DeployStatus, environmentsSeen map[string]bool) {
	for _, _rawDeploy := range rawDeploys {
		rawDeploy := JsonObject(_rawDeploy.(map[string]interface{}))
		deploy := convertRawDeploy(rawDeploy, status)

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

func (deploy Deploy) Time() time.Time {
	time0 := time.Time{}

	if deploy.FinishedAt.After(time0) {
		return deploy.FinishedAt
	} else if deploy.QueuedAt.After(time0) {
		return deploy.QueuedAt
	} else {
		return deploy.EstimatedTime
	}
}
