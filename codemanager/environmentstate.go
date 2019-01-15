package codemanager

import (
	log "github.com/sirupsen/logrus"
	"sort"
)

type EnvironmentState struct {
	Environment string
	Deploys     []Deploy
}

type SortOrder int

const (
	Ascending  = iota
	Descending = iota
)

func (environmentState *EnvironmentState) AddDeploys(newDeploys []Deploy) {
	log.Debugf("%s AddDeploys([%d]Deploy)",
		environmentState.Environment,
		len(newDeploys))

	deploysToAdd := []Deploy{}

	environmentState.SortDeploys(Descending)
	oldDeploysMatched := make(map[int]bool, len(environmentState.Deploys))

	_sortDeploys(newDeploys, Descending)
	for _, newDeploy := range newDeploys {
		possibleMatch := -1
		found := false
		for i, oldDeploy := range environmentState.Deploys {
			if oldDeploysMatched[i] {
				// An old deploy can only be updated be one new record.
				continue
			}

			match := oldDeploy.Match(newDeploy)
			if match == Yes {
				oldDeploy.Update(newDeploy)
				oldDeploysMatched[i] = true
				found = true
				break
			} else if match == Maybe && possibleMatch < 0 {
				log.Tracef("First possible match found")
				possibleMatch = i
			}
		}

		if found {
			continue
		}

		if possibleMatch >= 0 {
			log.Tracef("Using possible match")
			environmentState.Deploys[possibleMatch].Update(newDeploy)
			oldDeploysMatched[possibleMatch] = true
			continue
		}

		// It's new
		deploysToAdd = append(deploysToAdd, newDeploy)
	}

	for i, oldDeploy := range environmentState.Deploys {
		if !oldDeploysMatched[i] && !oldDeploy.Status.Finished() {
			log.Tracef("Found ghost deploy")
			environmentState.Deploys[i].Status = Ghost
		}
	}

	environmentState.Deploys = append(environmentState.Deploys, deploysToAdd...)
}

func _sortDeploys(deploys []Deploy, order SortOrder) {
	sort.Slice(deploys, func(i, j int) bool {
		a := deploys[i].MatchTime()
		b := deploys[j].MatchTime()
		if order == Descending {
			return a.After(b)
		} else {
			return a.Before(b)
		}
	})
}

func (environmentState *EnvironmentState) SortDeploys(order SortOrder) {
	_sortDeploys(environmentState.Deploys, order)
}

// Convenience, consistency, clarity
func (environmentState *EnvironmentState) SortedDeploys(order SortOrder) []Deploy {
	_sortDeploys(environmentState.Deploys, order)
	return environmentState.Deploys
}
