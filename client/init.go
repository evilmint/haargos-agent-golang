package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

const apiURL = "https://api.haargos.com/"

type Client struct {
	BaseURL string
	Logger  *logrus.Logger
}

func NewClient() *Client {
	return &Client{
		BaseURL: apiURL,
		Logger:  logrus.New(),
	}
}

func (c *Client) SendObservation(observation types.Observation, agentToken string) (*http.Response, error) {
	url := c.BaseURL + "observations"

	jsonData, err := json.Marshal(observation)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}

	c.Logger.Infof("Sending %s", string(jsonData))

	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err = g.Write(b); err != nil {
		c.Logger.Error(err)
		return nil, fmt.Errorf("error compressin JSON: %v", err)
	}
	if err = g.Close(); err != nil {
		c.Logger.Error(err)
		return nil, fmt.Errorf("error compressin JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-agent-token", agentToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	fmt.Printf("Response status: %s\n", resp.Status)
	return resp, nil
}
