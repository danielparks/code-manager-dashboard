package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type DeployStatus int

const (
	New       DeployStatus = iota
	Queued    DeployStatus = iota
	Deploying DeployStatus = iota
	Deployed  DeployStatus = iota
	Failed    DeployStatus = iota
	Deleted   DeployStatus = iota
)

var DeployStatusNames = [...]string{
	"new",
	"queued",
	"deploying",
	"deployed",
	"failed",
	"deleted",
}

/// FIXME needs testing
func stringToDeployStatus(status string) (DeployStatus, error) {
	for index, name := range DeployStatusNames {
		if name == status {
			return DeployStatus(index), nil
		}
	}

	return New, errors.New(fmt.Sprintf("Invalid status name %q", status))
}

func (status DeployStatus) String() string {
	return DeployStatusNames[status]
}

func (status DeployStatus) Finished() bool {
	return status >= Deployed
}

func (status *DeployStatus) UnmarshalJSON(raw []byte) error {
	var rawString string
	err := json.Unmarshal(raw, &rawString)
	if err != nil {
		return err
	}

	_status, err := stringToDeployStatus(rawString)
	if err != nil {
		return err
	}

	*status = _status
	return nil
}

func (status DeployStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}
