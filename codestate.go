package main

import (
	"log"
	"time"
)

type CodeState struct {
	Environments map[string]EnvironmentState
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

func (codeState *CodeState) addRawDeploys(rawDeploys []interface{}, status DeployStatus, environmentsSeen map[string]bool) {
	for _, _rawDeploy := range rawDeploys {
		rawDeploy := JsonObject(_rawDeploy.(map[string]interface{}))
		deploy := convertRawDeploy(rawDeploy, status)

		codeState.AddDeploy(deploy)
		environmentsSeen[deploy.Environment] = true
	}
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
	}

	deploy.CorrectFailedStatus()

	return deploy
}
