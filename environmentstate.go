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
