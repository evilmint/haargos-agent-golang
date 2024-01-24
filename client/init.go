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
	BaseURL        string
	AgentToken     string
	Logger         *logrus.Logger
	OnDataSentInKb func(int)
}

type AgentConfigResponse struct {
	Body AgentConfig `json:"body"`
}

type AgentConfig struct {
	CycleInterval int `json:"cycle_interval"`
}

func NewClient(apiURL string, agentToken string, dataSentInKb func(int)) *HaargosClient {
	return &HaargosClient{
		BaseURL:        apiURL,
		AgentToken:     agentToken,
		Logger:         logrus.New(),
		OnDataSentInKb: dataSentInKb,
	}
}

func (c *HaargosClient) sendRequest(method, url string, data interface{}, headers map[string]string) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}

	c.Logger.Debugf("Sending %s", string(jsonData))

	hasPayload := data != nil && (strings.ToLower(method) == "put" || strings.ToLower(method) == "post")
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

		c.OnDataSentInKb(buf.Len())
	}

	req, err := http.NewRequest(method, c.BaseURL+url, body)

	for key, value := range headers {
		req.Header.Add(key, value)
	}

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
		return resp, fmt.Errorf("error sending request: %v", err)
	}

	c.Logger.Debugf("Response status: %s", resp.Status)
	return resp, nil
}

func (c *HaargosClient) FetchText(url string, headers map[string]string) (string, error) {
	resp, err := c.sendRequest("GET", url, nil, headers)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *HaargosClient) FetchAgentConfig() (*AgentConfig, error) {
	resp, err := c.sendRequest("GET", "agent-config", nil, make(map[string]string))
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

type Addon struct {
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Description     string `json:"description"`
	Advanced        bool   `json:"advanced"`
	Stage           string `json:"stage"`
	Version         string `json:"version"`
	VersionLatest   string `json:"version_latest"`
	UpdateAvailable bool   `json:"update_available"`
	Available       bool   `json:"available"`
	Detached        bool   `json:"detached"`
	Homeassistant   string `json:"homeassistant"`
	State           string `json:"state"`
	Repository      string `json:"repository"`
	Build           bool   `json:"build"`
	URL             string `json:"url"`
	Icon            bool   `json:"icon"`
	Logo            bool   `json:"logo"`
}

type SupervisorResponse struct {
	Data struct {
		Addons []Addon `json:"addons"`
	} `json:"data"`
}

func (c *HaargosClient) FetchAddons(headers map[string]string) (*[]Addon, error) {
	resp, err := c.sendRequest("GET", "addons", nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	var response SupervisorResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &response.Data.Addons, nil
}

func (c *HaargosClient) FetchSupervisor(headers map[string]string) (*types.SupervisorInfo, error) {
	resp, err := c.sendRequest("GET", "supervisor/info", nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	var response types.SupervisorInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &response.Data, nil
}

func (c *HaargosClient) FetchOS(headers map[string]string) (*types.OSInfo, error) {
	resp, err := c.sendRequest("GET", "os/info", nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	var response types.OSInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &response.Data, nil
}

func (c *HaargosClient) UpdateCore(headers map[string]string) (*http.Response, error) {
	resp, err := c.sendRequest("POST", "core/update", nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	return resp, nil
}

func (c *HaargosClient) UpdateAddon(headers map[string]string, slug string) (*http.Response, error) {
	resp, err := c.sendRequest("POST", fmt.Sprintf("store/addons/%s/update", slug), nil, headers)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	return resp, nil
}

func (c *HaargosClient) CompleteJob(job types.GenericJob) (*[]types.GenericJob, error) {
	resp, err := c.sendRequest("POST", fmt.Sprintf("installations/jobs/%s/complete", job.ID), nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	var response types.JobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &response.Body, nil
}

func (c *HaargosClient) FetchJobs() (*[]types.GenericJob, error) {
	resp, err := c.sendRequest("GET", "installations/jobs/pending", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	var response types.JobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &response.Body, nil
}

type NotificationRequest struct {
	Notifications []websocketclient.WSAPINotificationDetails `json:"notifications"`
}

func (c *HaargosClient) SendNotifications(notifications []websocketclient.WSAPINotificationDetails) (*http.Response, error) {
	requestData := NotificationRequest{Notifications: notifications}
	return c.sendRequest("PUT", "installations/notifications", requestData, make(map[string]string))
}

func (c *HaargosClient) SendLogs(logs types.Logs) (*http.Response, error) {
	return c.sendRequest("PUT", "installations/logs", logs, make(map[string]string))
}

func (c *HaargosClient) SendAddons(addons []Addon) (*http.Response, error) {
	return c.sendRequest("PUT", "installations/addons", addons, make(map[string]string))
}

func (c *HaargosClient) SendSupervisor(supervisor types.SupervisorInfo) (*http.Response, error) {
	return c.sendRequest("PUT", "installations/supervisor", supervisor, make(map[string]string))
}

func (c *HaargosClient) SendOS(os types.OSInfo) (*http.Response, error) {
	return c.sendRequest("PUT", "installations/os", os, make(map[string]string))
}

func (c *HaargosClient) SendObservation(observation types.Observation) (*http.Response, error) {
	return c.sendRequest("POST", "observations", observation, make(map[string]string))
}
