package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	apiGithubUrl = "https://api.github.com"
	repoOwner    = "MillQK"
	repoName     = "d2tool"
)

type Client interface {
	// GetLatestRelease fetches the latest release
	GetLatestRelease() (*Release, error)
}

type HttpClient struct {
	httpClient *http.Client
	apiUrl     string
}

// NewHttpClient creates a new GitHub API client
// If apiUrl is empty, the default GitHub API URL will be used
func NewHttpClient(apiUrl string) *HttpClient {
	apiUrl = strings.TrimRight(apiUrl, "/")
	if apiUrl == "" {
		apiUrl = apiGithubUrl
	}

	return &HttpClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiUrl: apiUrl,
	}
}

func (c *HttpClient) GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.apiUrl, repoOwner, repoName)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/vnd.github+json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", response.StatusCode, response.Status)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var release Release
	err = json.Unmarshal(responseBody, &release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}
