package main

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type CodeState struct {
	Environments map[string]EnvironmentState
}

func (codeState *CodeState) UpdateFromRawCodeState(rawCodeState JsonObject) {
	log.Debug("CodeState<>.UpdateFromRawCodeState(<>)")

	newDeploys := map[string][]Deploy{}

	deploysStatus := rawCodeState.GetObject("deploys-status")
	mappings := map[string]DeployStatus{
		"new":       New,
		"queued":    Queued,
		"deploying": Deploying,
		"failed":    Failed,
	}

	for key, status := range mappings {
		rawDeploys := deploysStatus.GetArray(key)
		convertRawDeploys(rawDeploys, status, &newDeploys)
	}

	fileSyncStatus := rawCodeState.GetObject("file-sync-storage-status")
	rawDeploys := fileSyncStatus.GetArray("deployed")
	convertRawDeploys(rawDeploys, Deployed, &newDeploys)

	environmentsSeen := map[string]bool{}
	for name, environmentState := range codeState.Environments {
		environmentsSeen[name] = true
		if newDeploys[name] != nil {
			environmentState.AddDeploys(newDeploys[name])
		} else if environmentState.Deploys[0].Status != Deleted {
			log.Debugf("Environment %q not in the latest status update", name)
			// This environment wasn't in the current update, and its last recorded
			// status isn't Deleted. So, it needs a Deleted record.
			environmentState.AddDeploys([]Deploy{
				Deploy{
					Environment:   name,
					Status:        Deleted,
					EstimatedTime: time.Now(),
				},
			})
		}

		//		environmentState.RemoveDuplicateDeploys()
		// RemoveDuplicateDeploys will sort
		// environmentState.SortDeploys(Descending)

		codeState.Environments[name] = environmentState
	}

	if codeState.Environments == nil {
		codeState.Environments = map[string]EnvironmentState{}
	}

	// Look for all the environments we haven't already seen.
	for name, deploys := range newDeploys {
		if environmentsSeen[name] {
			continue
		}

		newEnvironmentState := EnvironmentState{Environment: name}
		newEnvironmentState.AddDeploys(deploys)
		codeState.Environments[name] = newEnvironmentState
	}
}

func convertRawDeploys(rawDeploys []interface{}, status DeployStatus, environments *map[string][]Deploy) {
	log.Debug("convertRawDeploys ", len(rawDeploys), " ", status, " deploys")
	for _, _rawDeploy := range rawDeploys {
		rawDeploy := JsonObject(_rawDeploy.(map[string]interface{}))
		deploy := convertRawDeploy(rawDeploy, status)

		deploys := (*environments)[deploy.Environment]
		(*environments)[deploy.Environment] = append(deploys, deploy)
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
