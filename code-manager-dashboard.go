package main

import (
	"encoding/json"
	"fmt"
	"github.com/pborman/getopt/v2"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"
)

const RFC3339Micro = "2006-01-02T15:04:05.999Z07:00"

type Deploy struct {
	Environment string
	Status      DeployStatus
	Sha         string
	Date        time.Time
	Error       map[string]interface{}
}

func convertRawDeploy(rawDeploy map[string]interface{}, status DeployStatus, dateKey string) Deploy {
	var sha string
	var date time.Time
	var err error

	environment := rawDeploy["environment"].(string)
	if rawDeploy["deploy-signature"] != nil {
		sha = rawDeploy["deploy-signature"].(string)
	}

	if rawDeploy[dateKey] != nil {
		date, err = time.Parse(RFC3339Micro, rawDeploy[dateKey].(string))
		if err != nil {
			log.Fatal(err)
		}
	}

	var deployError map[string]interface{}

	if rawDeploy["error"] != nil {
		deployError = jsonGetObject(rawDeploy, "error")

		if status == Failed && deployError != nil && deployError["msg"] != nil && strings.Contains(deployError["msg"].(string), "cannot be found in any source and will not be deployed.") {
			// Check for Environment(s) 'combined_minor_changes' cannot be found in any source and will not be deployed.
			// if it's in the error.msg, then convert this to Delete
			status = Deleted
		}
	}

	return Deploy{environment, status, sha, date, deployError}
}

func convertRawDeploys(deploys *map[string][]Deploy, rawDeploys []interface{}, status DeployStatus, dateKey string, environmentsSeen map[string]bool) {
	for _, _rawDeploy := range rawDeploys {
		rawDeploy := _rawDeploy.(map[string]interface{})
		deploy := convertRawDeploy(rawDeploy, status, dateKey)
		environmentDeploys := (*deploys)[deploy.Environment]
		(*deploys)[deploy.Environment] = append(environmentDeploys, deploy)
		environmentsSeen[deploy.Environment] = true
	}
}

func jsonGetArray(parent map[string]interface{}, key string) []interface{} {
	return parent[key].([]interface{})
}

func jsonGetObject(parent map[string]interface{}, key string) map[string]interface{} {
	return parent[key].(map[string]interface{})
}

func updateEnvironmentMap(environmentMap *map[string][]Deploy, rawDeployStatus map[string]interface{}) {
	// Clear all deployments except for the finished ones — others will be
	// replaced from the current data.
	for environment, deploys := range *environmentMap {
		cleanedDeploys := []Deploy{}
		for _, deploy := range deploys {
			if deploy.Status >= Deployed {
				// Either Deployed or Failed.
				cleanedDeploys = append(cleanedDeploys, deploy)
			}
		}
		(*environmentMap)[environment] = cleanedDeploys
	}

	var rawDeploys []interface{}
	environmentsSeen := map[string]bool{}

	fileSyncStatus := jsonGetObject(rawDeployStatus, "file-sync-storage-status")
	deploysStatus := jsonGetObject(rawDeployStatus, "deploys-status")

	rawDeploys = jsonGetArray(deploysStatus, "new")
	convertRawDeploys(environmentMap, rawDeploys, New, "queued-at", environmentsSeen)

	rawDeploys = jsonGetArray(deploysStatus, "queued")
	convertRawDeploys(environmentMap, rawDeploys, Queued, "queued-at", environmentsSeen)

	rawDeploys = jsonGetArray(deploysStatus, "deploying")
	convertRawDeploys(environmentMap, rawDeploys, Deploying, "queued-at", environmentsSeen)

	rawDeploys = jsonGetArray(fileSyncStatus, "deployed")
	convertRawDeploys(environmentMap, rawDeploys, Deployed, "date", environmentsSeen)

	rawDeploys = jsonGetArray(deploysStatus, "failed")
	convertRawDeploys(environmentMap, rawDeploys, Failed, "queued-at", environmentsSeen)

	for environment, deploys := range *environmentMap {
		if !environmentsSeen[environment] && deploys[0].Status != Deleted {
			// This environment is wasn't in the current update, and its last recorded
			// status isn't Deleted.
			deploys = append(deploys, Deploy{environment, Deleted, "", time.Now(), nil})
		}

		uniqueDeploys := []Deploy{}

		// Remove duplicates
		seen := map[string]bool{}
		for _, deploy := range deploys {
			if deploy.Status >= Deployed {
				key := fmt.Sprintf("%s %s %s", deploy.Status, deploy.Sha, deploy.Date)
				if seen[key] {
					continue
				}

				seen[key] = true
			}

			uniqueDeploys = append(uniqueDeploys, deploy)
		}

		// Sort
		sort.Slice(uniqueDeploys, func(i, j int) bool {
			a := uniqueDeploys[i]
			b := uniqueDeploys[j]
			if a.Status >= Deployed && b.Status >= Deployed {
				// Either Deployed or Failed. These should be sorted together by date.
				return a.Date.After(b.Date)
			} else if a.Status == b.Status {
				// Same status, so sort on date.
				return a.Date.After(b.Date)
			} else {
				return b.Status > a.Status
			}
		})

		(*environmentMap)[environment] = uniqueDeploys
	}
}

func sortedEnvironments(environmentMap map[string][]Deploy) [][]Deploy {
	environments := make([][]Deploy, len(environmentMap))
	i := 0
	for _, value := range environmentMap {
		environments[i] = value
		i++
	}

	sort.Slice(environments, func(i, j int) bool {
		a := environments[i]
		b := environments[j]
		return strings.ToLower(a[0].Environment) < strings.ToLower(b[0].Environment)
	})

	return environments
}

func displayEnvironments(environmentMap map[string][]Deploy) {
	environments := sortedEnvironments(environmentMap)

	now := time.Now().Truncate(time.Second)
	localZone, localZoneOffset := now.Zone()
	location := time.FixedZone(localZone, localZoneOffset)

	for _, deploys := range environments {
		environment := deploys[0].Environment

		for _, deploy := range deploys {
			localDate := deploy.Date.Truncate(time.Second).In(location)
			elapsed := deploy.Date.Truncate(time.Second).Sub(now)

			fmt.Printf("%-45s  %-9s  %s  %v\n", environment, deploy.Status, localDate, elapsed)
			environment = ""
		}
	}
}

// Get deploy status from API
func getRawDeployStatus() map[string]interface{} {
	rawDeployStatus := map[string]interface{}{}
	err := json.Unmarshal(GetDeployStatus(), &rawDeployStatus)
	if err != nil {
		log.Fatal(err)
	}

	return rawDeployStatus
}

// Get deploy status from file
func loadRawDeployStatus(source string) map[string]interface{} {
	deployStatusJSON, err := ioutil.ReadFile(source)
	if err != nil {
		log.Fatal(err)
	}

	rawDeployStatus := map[string]interface{}{}
	err = json.Unmarshal(deployStatusJSON, &rawDeployStatus)
	if err != nil {
		log.Fatal(err)
	}

	return rawDeployStatus
}

func dumpState(environmentMap map[string][]Deploy, path string) error {
	environmentsJSON, err := json.MarshalIndent(environmentMap, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("Dumping state into %q", path)

	/// FIXME should we lock this?
	return ioutil.WriteFile(path, append(environmentsJSON, '\n'), 0644)
}

func main() {
	var fakeStatus bool
	getopt.FlagLong(&fakeStatus, "fake-status", 0,
		"Treat arguments as list of files to load deploy statuses from.")

	var stateFile string
	getopt.FlagLong(&stateFile, "state-file", 's',
		"File to store state in.")

	getopt.Parse()
	args := getopt.Args()

	environmentMap := map[string][]Deploy{}

	if fakeStatus {
		for _, source := range args {
			updateEnvironmentMap(&environmentMap, loadRawDeployStatus(source))
		}
	} else {
		updateEnvironmentMap(&environmentMap, getRawDeployStatus())
	}

	displayEnvironments(environmentMap)

	if stateFile != "" {
		err := dumpState(environmentMap, stateFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}
