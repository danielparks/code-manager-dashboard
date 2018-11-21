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

// Create a tls.Config that recognizes a named CA cert
func LoadCaCert(path string) tls.Config {
	caCert, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return tls.Config{
		RootCAs: caCertPool,
	}
}

func httpClient(tlsConfig *tls.Config) *http.Client {
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   60 * time.Second,
	}
}

func getRawCodeStateJson(server string, port uint16, tlsConfig *tls.Config) []byte {
	url := fmt.Sprintf("https://%s:%d/code-manager/v1/deploys/status", server, port)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	pe_token := os.Getenv("pe_token")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Authentication", pe_token)

	response, err := httpClient(tlsConfig).Do(request)
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

func GetRawCodeState(server string, port uint16, tlsConfig *tls.Config) map[string]interface{} {
	codeState := map[string]interface{}{}
	err := json.Unmarshal(getRawCodeStateJson(server, port, tlsConfig), &codeState)
	if err != nil {
		log.Fatal(err)
	}

	return codeState
}
