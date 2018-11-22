package main

import (
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
