package main

import (
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

func (deploy Deploy) CorrectFailedStatus() bool {
	// If the status is Failed and the error message contains the below, then it
	// actually the represents environment being deleted.
	const deletionMsg = "cannot be found in any source and will not be deployed."

	if deploy.Status == Failed && deploy.Error != nil && deploy.Error["msg"] != nil {
		msg := deploy.Error["msg"].(string)

		if strings.Contains(msg, deletionMsg) {
			deploy.Status = Deleted
			return true
		}
	}

	return deploy.Status == Deleted
}
