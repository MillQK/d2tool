package github

import (
	"context"
	"encoding/json"
	"fmt"
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
	GetLatestRelease(ctx context.Context) (*Release, error)
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

func (c *HttpClient) GetLatestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.apiUrl, repoOwner, repoName)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var release Release
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}
