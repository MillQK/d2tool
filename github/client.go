package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	url       = "https://api.github.com"
	repoOwner = "MillQK"
	repoName  = "d2tool"
)

type Client interface {
	// GetLatestRelease fetches the latest release
	GetLatestRelease() (*Release, error)
}

type HttpClient struct {
	httpClient *http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		httpClient: http.DefaultClient,
	}
}

func (c *HttpClient) GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", url, repoOwner, repoName)

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
