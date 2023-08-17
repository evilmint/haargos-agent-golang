package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/evilmint/haargos-agent-golang/types"
)

var API_URL = "https://api.haargos.smartrezydencja.pl/"

func SendObservation(observation types.Observation, userID, token string) (*http.Response, error) {
	url := API_URL + "observations?installation_id=f2687b3e-d6f7-4cbd-a58b-48000752c2a9"

	jsonData, err := json.Marshal(observation)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", userID)
	req.Header.Set("x-token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	fmt.Printf("Response status: %s\n", resp.Status)
	return resp, nil
}
