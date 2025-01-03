package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"unicode"
)

type GetWeather struct {
}

func NewGetWeather() Tool {
	return &GetWeather{}
}

func (w *GetWeather) Name() string {
	return "get_weather"
}

func (w *GetWeather) Description() string {
	return "Retrieve the current weather information for a specified city and return the city name in English"
}

func (w *GetWeather) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"city": map[string]interface{}{
				"type":        "string",
				"description": "city name",
			},
		},
		"required": []string{"city"},
	}
}

func (w *GetWeather) Execute(ctx context.Context, input string) (string, error) {
	var arguments struct {
		City string `json:"city"`
	}
	if err := json.Unmarshal([]byte(input), &arguments); err != nil {
		slog.Error("unmarshal failed", slog.Any("err", err))
		return "", err
	}

	if arguments.City == "" {
		return "", fmt.Errorf("city name is empty")
	}

	city := w.completeCNCity(arguments.City)

	appID := os.Getenv("OPENWEATHERMAP_API_KEY")
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%v&units=metric&appid=%v&lang=zh_cn&cnt=5", city, appID)

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

	var weatherData struct {
		List []struct {
			Dt   int `json:"dt"`
			Main struct {
				Temp float64 `json:"temp"`
			} `json:"main"`
			Weather []struct {
				Description string `json:"description"`
			} `json:"weather"`
			DtTxt string `json:"dt_txt"`
		} `json:"list"`
		City struct {
			Name  string `json:"name"`
			Coord struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"coord"`
			Sunrise int `json:"sunrise"`
			Sunset  int `json:"sunset"`
		} `json:"city"`
	}

	if err := json.Unmarshal(content, &weatherData); err != nil {
		return "", err
	}

	mainContent, _ := json.Marshal(weatherData)

	return string(mainContent), nil
}

func (w *GetWeather) isCNCity(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return true
}

func (w *GetWeather) completeCNCity(city string) string {
	if w.isCNCity(city) && !strings.HasSuffix(city, "市") {
		return city + "市"
	}
	return city
}
