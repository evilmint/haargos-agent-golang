package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/evilmint/haargos-agent-golang/types"
	websocketclient "github.com/evilmint/haargos-agent-golang/websocket-client"
	"github.com/sirupsen/logrus"
)

type HaargosClient struct {
	BaseURL    string
	AgentToken string
	Logger     *logrus.Logger
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

func (c *HaargosClient) sendRequest(method, url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}

	c.Logger.Debugf("Sending %s", string(jsonData))

	hasPayload := strings.ToLower(method) == "put" || strings.ToLower(method) == "post"
	var body io.Reader = nil // Initialize body as nil

	if hasPayload {
		buf := new(bytes.Buffer)

		g := gzip.NewWriter(buf)
		if _, err = g.Write(jsonData); err != nil {
			c.Logger.Error(err)
			return nil, fmt.Errorf("error compressing JSON: %v", err)
		}
		if err = g.Close(); err != nil {
			c.Logger.Error(err)
			return nil, fmt.Errorf("error compressing JSON: %v", err)
		}
		body = buf
	}

	req, err := http.NewRequest(method, c.BaseURL+url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	if hasPayload {
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("x-agent-token", c.AgentToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	c.Logger.Debugf("Response status: %s", resp.Status)
	return resp, nil
}

func (c *HaargosClient) FetchText(url string) (string, error) {
	resp, err := c.sendRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	buf := new(strings.Builder)
	n, err := io.Copy(buf, resp.Body)

	if err != nil {
		return "", err
	}

	return buf.String(), nil
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

type NotificationRequest struct {
	Notifications []websocketclient.WSAPINotificationDetails `json:"notifications"`
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
