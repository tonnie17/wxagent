package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

type GoogleSearch struct {
}

func NewGoogleSearch() Tool {
	return &GoogleSearch{}
}

func (w *GoogleSearch) Name() string {
	return "google_search"
}

func (w *GoogleSearch) Description() string {
	return "Make a query to the Google search engine to receive a list of results"
}

func (w *GoogleSearch) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The query to be passed to Google search",
			},
		},
		"required": []string{"query"},
	}
}

func (w *GoogleSearch) Execute(ctx context.Context, input string) (string, error) {
	var arguments struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(input), &arguments); err != nil {
		slog.Error("unmarshal failed", slog.Any("err", err))
		return "", err
	}

	apiKey := os.Getenv("GOOGLE_SEARCH_API_KEY")
	engine := os.Getenv("GOOGLE_SEARCH_ENGINE")
	apiURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%v&cx=%v&q=%v&num=3", apiKey, engine, url.QueryEscape(arguments.Query))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	sr := &searchResult{}
	if err := json.Unmarshal(content, &sr); err != nil {
		return "", err
	}

	itemsJSON, _ := json.Marshal(sr.Items)
	return string(itemsJSON), nil
}

type searchResult struct {
	Items []*searchResultItem `json:"items"`
}

type searchResultItem struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}
