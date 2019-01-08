package codemanager

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ApiClient struct {
	Host       string
	Port       uint16
	RbacToken  string
	HttpClient *http.Client
}

func TypicalApiClient(host string, rbacToken string, caPath string) *ApiClient {
	return &ApiClient{
		Host:       host,
		Port:       8170,
		RbacToken:  rbacToken,
		HttpClient: ApiHttpClient(LoadCaCert(caPath)),
	}
}

// Create a tls.Config that recognizes a named CA cert
func LoadCaCert(path string) *tls.Config {
	caCert, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		RootCAs: caCertPool,
	}
}

func ApiHttpClient(tlsConfig *tls.Config) *http.Client {
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   5 * time.Minute,
	}
}

func getRawCodeStateJson(client *ApiClient) []byte {
	url := fmt.Sprintf("https://%s:%d/code-manager/v1/deploys/status",
		client.Host, client.Port)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	request.Header.Set("Accept", "application/json")

	if client.RbacToken != "" {
		request.Header.Set("X-Authentication", client.RbacToken)
	}

	response, err := client.HttpClient.Do(request)
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

func (client *ApiClient) GetRawCodeState() map[string]interface{} {
	codeState := map[string]interface{}{}
	err := json.Unmarshal(getRawCodeStateJson(client), &codeState)
	if err != nil {
		log.Fatal(err)
	}

	return codeState
}
