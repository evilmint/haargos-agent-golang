package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/evilmint/haargos-agent-golang/types"
	websocketclient "github.com/evilmint/haargos-agent-golang/websocket-client"
	"github.com/sirupsen/logrus"
)

type HaargosClient struct {
	BaseURL    string
	AgentToken string
	Logger     *logrus.Logger
}

type NotificationRequest struct {
	Notifications []websocketclient.WSAPINotificationDetails `json:"notifications"`
}
type AgentConfigResponse struct {
	Body AgentConfig `json:"body"`
}

type AgentConfig struct {
	CycleInterval int `json:"cycle_interval"`
}

func NewClient(apiURL string, agentToken string) *HaargosClient {
	return &HaargosClient{
		BaseURL:    apiURL,
		AgentToken: agentToken,
		Logger:     logrus.New(),
	}
}

func (c *HaargosClient) FetchAgentConfig() (*AgentConfig, error) {
	resp, err := c.sendRequest("GET", "agent-config", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	var config AgentConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &config.Body, nil
}

func (c *HaargosClient) SendNotifications(notifications []websocketclient.WSAPINotificationDetails) (*http.Response, error) {
	requestData := NotificationRequest{Notifications: notifications}
	return c.sendRequest("PUT", "installations/notifications", requestData)
}

func (c *HaargosClient) SendLogs(logs types.Logs) (*http.Response, error) {
	return c.sendRequest("PUT", "installations/logs", logs)
}

func (c *HaargosClient) SendObservation(observation types.Observation) (*http.Response, error) {
	return c.sendRequest("POST", "observations", observation)
}

func (c *HaargosClient) sendRequest(method, url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
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

	req, err := http.NewRequest(method, c.BaseURL+url, &buf)
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

	c.Logger.Debugf("Response status: %s", resp.Status)
	return resp, nil
}
