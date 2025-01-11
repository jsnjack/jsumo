package cmd

// Collector is the Collector instance in SumoLogic
type Collector struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	Timezone      string `json:"timezone"`
	CollectorType string `json:"collectorType"`
}

// Source is the Source instance in SumoLogic
type Source struct {
	ID                         int    `json:"id"`
	Name                       string `json:"name"`
	Description                string `json:"description"`
	Category                   string `json:"category"`
	HostName                   string `json:"hostName"`
	Timezone                   string `json:"timezone"`
	SourceType                 string `json:"sourceType"`
	Encoding                   string `json:"encoding"`
	ForceTimeZone              bool   `json:"forceTimeZone"`
	ContentType                string `json:"contentType"`
	MultilineProcessingEnabled bool   `json:"multilineProcessingEnabled"`
	AutomaticDateParsing       bool   `json:"automaticDateParsing"`
	URL                        string `json:"url"`
}

// CollectorResponse is the response from the SumoLogic API when creating a collector
// Ref: https://help.sumologic.com/docs/api/collector-management/collector-api-methods-examples/#create-hosted-collector
type CollectorResponse struct {
	Collector Collector `json:"collector"`
}

type CollectorsListResponse struct {
	Collectors []Collector `json:"collectors"`
}

type SourceResponse struct {
	Source Source `json:"source"`
}

type SourcesListResponse struct {
	Sources []Source `json:"sources"`
}
