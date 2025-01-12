package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// sumoRESTAPIURL is the URL of the SumoLogic REST API. Note: this is the URL for the SumoLogic DE environment.
const sumoRESTAPIURL = "https://api.de.sumologic.com/api/v1"
const collectorNamePrefix = "jsumo-collector-"
const sumoAccessIDEnvVar = "SUMO_ACCESSID"
const sumoAccessKeyEnvVar = "SUMO_ACCESSKEY"

// GetReceiverURL returns the URL of the SumoLogic receiver which is used to send
// logs to SumoLogic. The URL is constructed using the hostname of the machine.
// If it doesn't exist, a new collector and source are created in SumoLogic.
func GetReceiverURL() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	DebugLogger.Println("System hostname:", hostname)

	// Get the collector ID from the collector name
	collectorID, err := getSumoCollectorIDFromName(hostname)
	if err != nil {
		// Create a new collector if it doesn't exist
		DebugLogger.Println(red(err))
		collectorID, err = createSumoCollector(hostname)
		if err != nil {
			return "", err
		}
	}

	// Get the source receiver URL from the source name
	receiverURL, err := getSumoHTTPSourceReceiverURLFromName(collectorID, hostname)
	if err != nil {
		// Create a new source if it doesn't exist
		DebugLogger.Println(red(err))
		receiverURL, err = createSumoHTTPSource(collectorID, hostname)
		if err != nil {
			return "", err
		}
	}

	return receiverURL, nil

}

// createSumoCollector creates a new collector in SumoLogic
func createSumoCollector(name string) (int, error) {
	DebugLogger.Println(green(fmt.Sprintf("Creating collector with name: %s", name)))
	url := sumoRESTAPIURL + "/collectors"
	body := map[string]interface{}{
		"collector": map[string]interface{}{
			"name":          name,
			"description":   "Created by jsumo",
			"collectorType": "Hosted",
		},
	}

	respBodyBytes, err := makeRequest("POST", url, body)
	if err != nil {
		return 0, err
	}

	var response CollectorResponse
	if err := json.Unmarshal(respBodyBytes, &response); err != nil {
		return 0, err
	}
	return response.Collector.ID, nil
}

// getSumoCollectorIDFromName returns the ID of the collector with the given name
func getSumoCollectorIDFromName(name string) (int, error) {
	DebugLogger.Println(green(fmt.Sprintf("Getting collector ID with name: %s", name)))
	url := sumoRESTAPIURL + "/collectors"
	respBodyBytes, err := makeRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	var collectors CollectorsListResponse
	if err := json.Unmarshal(respBodyBytes, &collectors); err != nil {
		return 0, err
	}

	for _, collector := range collectors.Collectors {
		if collector.Name == name {
			return collector.ID, nil
		}
	}

	return 0, fmt.Errorf("collector with name %s not found", name)
}

// createSumoHTTPSource creates a new HTTP source in SumoLogic
func createSumoHTTPSource(collectorID int, sourceName string) (string, error) {
	DebugLogger.Println(green(fmt.Sprintf("Creating HTTP source with name: %s", sourceName)))
	url := sumoRESTAPIURL + "/collectors/" + fmt.Sprint(collectorID) + "/sources"

	// Ref for unique params: https://help.sumologic.com/docs/send-data/use-json-configure-sources/json-parameters-hosted-sources/#http-source
	// Ref for common params: https://help.sumologic.com/docs/send-data/use-json-configure-sources/#common-parameters-for-log-source-types
	body := map[string]interface{}{
		"source": map[string]interface{}{
			"name":                       sourceName,
			"description":                "Created by jsumo",
			"sourceType":                 "HTTP",
			"messagePerRequest":          false,
			"multilineProcessingEnabled": true,
			"hostName":                   sourceName,
			"category":                   sourceName,
		},
	}

	respBodyBytes, err := makeRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	var response SourceResponse
	if err := json.Unmarshal(respBodyBytes, &response); err != nil {
		return "", err
	}
	return response.Source.URL, nil
}

// getSumoHTTPSourceReceiverURLFromName returns the URL of the HTTP source with the given name
func getSumoHTTPSourceReceiverURLFromName(collectorID int, sourceName string) (string, error) {
	DebugLogger.Println(green(fmt.Sprintf("Getting HTTP source URL with name: %s", sourceName)))
	url := sumoRESTAPIURL + "/collectors/" + fmt.Sprint(collectorID) + "/sources"
	respBodyBytes, err := makeRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	var sources SourcesListResponse
	if err := json.Unmarshal(respBodyBytes, &sources); err != nil {
		return "", err
	}

	for _, source := range sources.Sources {
		if source.Name == sourceName {
			return source.URL, nil
		}
	}

	return "", fmt.Errorf("source with name %s not found", sourceName)
}

// uploadFileToSumoSource uploads a file to the SumoLogic source receiver URL
// Ref: https://help.sumologic.com/docs/send-data/hosted-collectors/http-source/logs-metrics/upload-logs/
func uploadFileToSumoSource(filename, receiverURL string) error {
	startedAt := time.Now()
	defer func() {
		DebugLogger.Printf("File uploaded %s, took %s\n", filename, time.Since(startedAt))
	}()
	DebugLogger.Println(green(fmt.Sprintf("Uploading file %s ...", filename)))
	file, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			DebugLogger.Println(yellow(fmt.Sprintf("File %s not found. Skipping upload.", filename)))
			return nil
		}
		return err
	}
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	req, err := http.NewRequest("POST", receiverURL, bytes.NewBuffer(file))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Encoding", "zstd")
	if FlagSourceCategory != "" {
		req.Header.Set("X-Sumo-Category", FlagSourceCategory)
	}

	DebugLogger.Println(blue(fmt.Sprintf("-- Making request POST %s", receiverURL)))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	DebugLogger.Println(blue(fmt.Sprintf("Response status: %s", resp.Status)))

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	DebugLogger.Println(blue(fmt.Sprintf("Response body: %s", string(respBody))))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error: status %s, %s", resp.Status, string(respBody))
	}

	Logger.Printf("Uploaded %d bytes\n", len(file))

	DebugLogger.Printf("Removing file %s\n", filename)
	err = os.Remove(filename)
	if err != nil {
		Logger.Println(err)
	}
	return nil
}

// makeRequest makes an HTTP request to the SumoLogic REST API
func makeRequest(method, url string, body map[string]interface{}) ([]byte, error) {
	DebugLogger.Println(blue(fmt.Sprintf("-- Making request %s %s", method, url)))
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	DebugLogger.Println(blue(fmt.Sprintf("Request body: %s", string(jsonBody))))

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	accessIDStr := os.Getenv(sumoAccessIDEnvVar)
	if accessIDStr == "" {
		return nil, fmt.Errorf("environment variable %s not set", sumoAccessIDEnvVar)
	}

	accessKeyStr := os.Getenv(sumoAccessKeyEnvVar)
	if accessKeyStr == "" {
		return nil, fmt.Errorf("environment variable %s not set", sumoAccessKeyEnvVar)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", accessIDStr, accessKeyStr)))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	DebugLogger.Println(blue(fmt.Sprintf("Response status: %s", resp.Status)))
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	DebugLogger.Println(blue(fmt.Sprintf("Response body: %s", string(respBody))))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: status %s, %s", resp.Status, string(respBody))
	}

	return respBody, nil
}
