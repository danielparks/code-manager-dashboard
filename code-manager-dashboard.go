package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/pborman/getopt/v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// Create an http.Client that recognizes the Puppet CA.
func httpClient() *http.Client {
	caCertPath := "/Users/daniel/work/puppetca.ops.puppetlabs.net.pem"

	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}

func getDeployStatus() []byte {
	server := "pe-mom1-prod.ops.puppetlabs.net"
	port := "8170"

	url := fmt.Sprintf("https://%s:%s/code-manager/v1/deploys/status", server, port)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	pe_token := os.Getenv("pe_token")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Authentication", pe_token)

	response, err := httpClient().Do(request)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		log.Fatalf("Unexpected status %q checking deployment status.", response.Status)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return body
}

func main() {
	const RFC3339Micro = "2006-01-02T15:04:05.999Z07:00"
	var deployStatusResponse []byte
	var err error

	deployStatusSource := getopt.StringLong("status-source", 'S', "",
		"File to use instead of deploy status API endpoint")
	getopt.Parse()

	if *deployStatusSource == "" {
		deployStatusResponse = getDeployStatus()
	} else {
		deployStatusResponse, err = ioutil.ReadFile(*deployStatusSource)
		if err != nil {
			log.Fatal(err)
		}
	}

	object := map[string]interface{}{}
	json.Unmarshal(deployStatusResponse, &object)

	fileSyncStatus := object["file-sync-storage-status"].(map[string]interface{})
	deployedEnvironments := fileSyncStatus["deployed"].([]interface{})

	now := time.Now()
	localZone, localZoneOffset := now.Zone()
	location := time.FixedZone(localZone, localZoneOffset)

	for _, _environment := range deployedEnvironments {
		environment := _environment.(map[string]interface{})
		dateString := environment["date"].(string)

		date, err := time.Parse(RFC3339Micro, dateString)
		if err != nil {
			log.Fatal(err)
		}

		localDate := date.In(location)

		fmt.Printf("%-45s %s	%v\n", environment["environment"], localDate, date.Sub(now))
	}
}
