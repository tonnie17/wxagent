package ha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetEntityStates(ctx context.Context, domains []string) ([]EntityState, error) {
	req, err := newHARequest(ctx, "GET", "/api/states", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	entityStates := make([]EntityState, 0)
	if err := json.Unmarshal(content, &entityStates); err != nil {
		return nil, err
	}

	res := make([]EntityState, 0)
	for _, entityState := range entityStates {
		if entityState.State == "unknown" {
			continue
		}
		split := strings.Split(entityState.EntityID, ".")
		if len(domains) > 0 && (len(split) < 2 || !containString(domains, split[0])) {
			continue
		}
		res = append(res, entityState)
	}

	return res, nil
}

func ExecuteService(ctx context.Context, domain string, service string, entityID string) ([]EntityState, error) {
	body := map[string]interface{}{
		"entity_id": entityID,
	}
	bodyJSON, _ := json.Marshal(body)
	req, err := newHARequest(ctx, "POST", fmt.Sprintf("/api/services/%v/%v", domain, service), bytes.NewReader(bodyJSON))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	entityStates := make([]EntityState, 0)
	if err := json.Unmarshal(content, &entityStates); err != nil {
		return nil, err
	}

	return entityStates, nil
}

func newHARequest(ctx context.Context, method string, apiPath string, body io.Reader) (*http.Request, error) {
	haBaseURL := os.Getenv("HA_BASE_URL")
	haToken := os.Getenv("HA_BEARER_TOKEN")

	apiURL := fmt.Sprintf("%v%v", haBaseURL, apiPath)
	req, err := http.NewRequest(method, apiURL, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", haToken))
	return req, nil
}

func containString(target []string, s string) bool {
	for _, t := range target {
		if s == t {
			return true
		}
	}
	return false
}
