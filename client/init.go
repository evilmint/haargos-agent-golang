package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

const apiURL = "https://api.haargos.smartrezydencja.pl/"

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
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

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
