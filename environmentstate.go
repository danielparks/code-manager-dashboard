package main

import (
	"fmt"
	"sort"
)

type EnvironmentState struct {
	Environment string
	Deploys     []Deploy
}

func (environmentState *EnvironmentState) AddDeploy(deploy Deploy) {
	environmentState.Deploys = append(environmentState.Deploys, deploy)
}

func (environmentState *EnvironmentState) ClearUnfinishedDeploys() {
	cleanedDeploys := []Deploy{}
	for _, deploy := range environmentState.Deploys {
		if deploy.Status.Finished() {
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

		// Group statuses into their own buckets, then sort within them.
		if a.Status == b.Status {
			return a.Time().After(b.Time())
		} else if a.Status.Finished() && b.Status.Finished() {
			// Group all finished statuses into the same bucket.
			return a.Time().After(b.Time())
		} else {
			return b.Status > a.Status
		}
	})
}
