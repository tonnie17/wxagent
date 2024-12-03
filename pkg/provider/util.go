package provider

import (
	"fmt"
	"os"
	"strings"
)

func GetAPIKey(provider string) string {
	return os.Getenv(fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider)))
}

func GetAPIBaseURL(provider string) string {
	return os.Getenv(fmt.Sprintf("%s_BASE_URL", strings.ToUpper(provider)))
}
