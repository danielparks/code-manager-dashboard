package codemanager

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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

func (deploy Deploy) DisplayTime() time.Time {
	// We really care about when it was finished.
	if deploy.HasFinishedTime() {
		return deploy.FinishedAt
	} else if deploy.HasQueuedTime() {
		return deploy.QueuedAt
	} else {
		return deploy.EstimatedTime
	}
}

func (deploy Deploy) MatchTime() time.Time {
	// Match on queued time, if possible
	if deploy.HasQueuedTime() {
		return deploy.QueuedAt
	} else if deploy.HasFinishedTime() {
		return deploy.FinishedAt
	} else {
		return deploy.EstimatedTime
	}
}

func (deploy Deploy) HasQueuedTime() bool {
	return deploy.QueuedAt.After(time.Time{})
}

func (deploy Deploy) HasFinishedTime() bool {
	return deploy.FinishedAt.After(time.Time{})
}

func (deploy Deploy) HasEstimatedTime() bool {
	return deploy.EstimatedTime.After(time.Time{})
}

func (deploy Deploy) String() string {
	return fmt.Sprintf("%s %s (%s)", deploy.Environment, deploy.Status, deploy.MatchTime())
}

func boolToEqual(value bool) string {
	if value {
		return "="
	} else {
		return "≠"
	}
}

// This attempts to match a deploy record, a, with an updated deploy record, b.
func (a Deploy) Match(b Deploy) Trinary {
	if a.Environment != b.Environment {
		log.Tracef("Match No: environment: %q ≠ %q", a.Environment, b.Environment)
		return No
	}
	/// FIXME, can't match if the sha isn't the same /////////////////

	if a.Status != b.Status && a.Status.Finished() && b.Status.Finished() {
		log.Tracef("Match No: different finished statuses: %s ≠ %s", a.Status, b.Status)
		// The statuses aren't actually the same, and deployments can't move from
		// one finished state to another.
		return No
	}

	time0 := time.Time{}
	sameTime := false
	debugTimeMatchString := ""

	if a.QueuedAt.After(time0) {
		// a.QueuedAt has been set. If they match, then they almost certainly refer
		// to the same deployment.
		sameTime = a.QueuedAt.Equal(b.QueuedAt)
		debugTimeMatchString = fmt.Sprintf("queued at: %s %s %s", a.QueuedAt, boolToEqual(sameTime), b.QueuedAt)
	} else if a.FinishedAt.After(time0) {
		// a.FinishedAt has been set. If they match, then they almost certainly
		// refer to the same deployment.
		sameTime = a.FinishedAt.Equal(b.FinishedAt)
		debugTimeMatchString = fmt.Sprintf("finished at: %s %s %s", a.FinishedAt, boolToEqual(sameTime), b.FinishedAt)
	} else {
		// We should never get an update to a deployment with only an estimated time,
		// since that's only used when an environment disappears.
		panic(fmt.Sprintf("Matching A has no finished or queued at time"))
	}

	if a.Status == b.Status {
		// If they have the same status, then they would have to have the same time.
		log.Tracef("Match %s: same status (%s). %s", BoolToTrinary(sameTime), a.Status, debugTimeMatchString)
		return BoolToTrinary(sameTime)
	}

	if sameTime {
		// Different unfinished statuses with the same time.
		log.Tracef("Match Yes: unfinished statuses (%s, %s) with %s", a.Status, b.Status, debugTimeMatchString)
		return Yes
	}

	if a.Status.Finished() == b.Status.Finished() {
		// Two finished deployments can only match if they have the same time.
		// Two unfinished deployments can only match if they have the same time.
		log.Tracef("Match No: one must be finished and one must be unfinished (%s, %s)", a.Status, b.Status)
		return No
	}

	// One status is finished and the other is unfinished. They have different
	// times (as one would expect).
	if a.Status.Finished() {
		log.Tracef("Match No: finished (%s) can't be updated with unfinished (%s)", a.Status, b.Status)
		// A finished before B was queued.
		return No
	}

	if a.MatchTime().After(b.MatchTime()) {
		log.Tracef("Match No: %s < %s", a.MatchTime(), b.MatchTime())
		// B finished before A was queued.
		return No
	}

	log.Trace("Match Maybe:")
	log.Tracef("  A: %s", a)
	log.Tracef("  B: %s", b)

	// Match candidates by sorting on time and complementary finishedness.
	return Maybe
}

func (deploy *Deploy) Update(newDeploy Deploy) {
	log.Debugf("Updating %q deploy from %s to %s",
		deploy.Environment, deploy.Status, newDeploy.Status)
	deploy.Status = newDeploy.Status

	if newDeploy.HasFinishedTime() {
		deploy.FinishedAt = newDeploy.FinishedAt
	}

	if newDeploy.HasEstimatedTime() {
		deploy.EstimatedTime = newDeploy.EstimatedTime
	}

	if newDeploy.Sha != "" {
		deploy.Sha = newDeploy.Sha
	}

	if newDeploy.Error != nil {
		deploy.Error = newDeploy.Error
	}
}
