package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"

	"io"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

const apiURL = "https://api.haargos.com/"

type HaargosClient struct {
	BaseURL    string
	AgentToken string
	Logger     *logrus.Logger
}

type AgentConfig struct {
	CycleInterval int `json:"cycle_interval"`
}

func NewClient(agentToken string) *HaargosClient {
	return &HaargosClient{
		BaseURL:    apiURL,
		AgentToken: agentToken,
		Logger:     logrus.New(),
	}
}

func (c *HaargosClient) FetchAgentConfig() (*AgentConfig, error) {
	url := c.BaseURL + "agent-config"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("x-agent-token", c.AgentToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(io.Reader(resp.Body))
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var config AgentConfig
	if err := json.Unmarshal(bodyBytes, &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &config, nil
}

func (c *HaargosClient) SendObservation(observation types.Observation) (*http.Response, error) {
	url := c.BaseURL + "observations"

	jsonData, err := json.Marshal(observation)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}

	c.Logger.Debugf("Sending %s", string(jsonData))

	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err = g.Write(jsonData); err != nil {
		c.Logger.Error(err)
		return nil, fmt.Errorf("error compressing JSON: %v", err)
	}
	if err = g.Close(); err != nil {
		c.Logger.Error(err)
		return nil, fmt.Errorf("error compressing JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-agent-token", c.AgentToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	fmt.Printf("Response status: %s\n", resp.Status)
	return resp, nil
}
