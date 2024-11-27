package tool

import (
	"context"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
)

type WebPageSummary struct {
}

func NewWebPageSummary() Tool {
	return &WebPageSummary{}
}

func (w *WebPageSummary) Name() string {
	return "webpage_summary"
}

func (w *WebPageSummary) Description() string {
	return "Summaries the content of a web page"
}

func (w *WebPageSummary) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "the URL of the web page",
			},
		},
		"required": []string{"url"},
	}
}

func (w *WebPageSummary) Execute(ctx context.Context, input string) (string, error) {
	var arguments struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(input), &arguments); err != nil {
		slog.Error("unmarshal failed", slog.Any("err", err))
		return "", err
	}

	req, err := http.NewRequest("GET", arguments.URL, nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)

	client := http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(r.Body)

	document, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	document.Find("script, style, pre, code").Each(func(index int, item *goquery.Selection) {
		item.Remove()
	})
	text := document.Text()

	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return text, nil
}
