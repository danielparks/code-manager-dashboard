package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
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

func getDeployStatus() []interface{} {
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

	object := map[string]interface{}{}
	json.Unmarshal(body, &object)

	statuses := object["file-sync-storage-status"].(map[string]interface{})
	return statuses["deployed"].([]interface{})
}

func main() {
	const RFC3339Micro = "2006-01-02T15:04:05.999Z07:00"
	location, err := time.LoadLocation("America/Los_Angeles") ////////////////////////////
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()

	statuses := getDeployStatus()
	for _, _environment := range statuses {
		environment := _environment.(map[string]interface{})
		dateString := environment["date"].(string)

		date, err := time.Parse(RFC3339Micro, dateString)
		if err != nil {
			log.Fatal(err)
		}

		localDate := date.In(location)

		fmt.Printf("%-45s %s  %v\n", environment["environment"], localDate, date.Sub(now))
	}
}
