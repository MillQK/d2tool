package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRelease_JSONParsing(t *testing.T) {
	jsonData := `{
		"url": "https://api.github.com/repos/owner/repo/releases/1",
		"html_url": "https://github.com/owner/repo/releases/tag/1.0.0",
		"tag_name": "1.0.0",
		"name": "Release 1.0.0",
		"body": "Release notes here",
		"draft": false,
		"prerelease": false,
		"created_at": "2024-01-15T10:00:00Z",
		"published_at": "2024-01-15T12:00:00Z",
		"assets": [
			{
				"url": "https://api.github.com/repos/owner/repo/releases/assets/1",
				"name": "d2tool-windows-amd64.zip",
				"size": 12345678,
				"browser_download_url": "https://github.com/owner/repo/releases/download/v1.0.0/d2tool-windows-amd64.zip",
				"content_type": "application/zip",
				"state": "uploaded",
				"download_count": 100
			}
		]
	}`

	var release Release
	err := json.Unmarshal([]byte(jsonData), &release)
	if err != nil {
		t.Fatalf("failed to parse release JSON: %v", err)
	}

	if release.TagName != "1.0.0" {
		t.Errorf("expected tag_name '1.0.0', got %q", release.TagName)
	}
	if release.Name != "Release 1.0.0" {
		t.Errorf("expected name 'Release 1.0.0', got %q", release.Name)
	}
	if release.Draft {
		t.Error("expected draft to be false")
	}
	if release.Prerelease {
		t.Error("expected prerelease to be false")
	}
	if len(release.Assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(release.Assets))
	}
}

func TestReleaseAsset_JSONParsing(t *testing.T) {
	jsonData := `{
		"url": "https://api.github.com/repos/owner/repo/releases/assets/1",
		"browser_download_url": "https://github.com/download/file.zip",
		"id": 12345,
		"node_id": "RA_kwDOAbcdef",
		"name": "d2tool-windows-amd64.zip",
		"label": "Windows AMD64",
		"state": "uploaded",
		"content_type": "application/zip",
		"size": 15728640,
		"download_count": 500,
		"created_at": "2024-01-15T10:00:00Z",
		"updated_at": "2024-01-15T10:05:00Z"
	}`

	var asset ReleaseAsset
	err := json.Unmarshal([]byte(jsonData), &asset)
	if err != nil {
		t.Fatalf("failed to parse asset JSON: %v", err)
	}

	if asset.Name != "d2tool-windows-amd64.zip" {
		t.Errorf("expected name 'd2tool-windows-amd64.zip', got %q", asset.Name)
	}
	if asset.Size != 15728640 {
		t.Errorf("expected size 15728640, got %d", asset.Size)
	}
	if asset.ContentType != "application/zip" {
		t.Errorf("expected content_type 'application/zip', got %q", asset.ContentType)
	}
	if asset.DownloadCount != 500 {
		t.Errorf("expected download_count 500, got %d", asset.DownloadCount)
	}
	if asset.State != "uploaded" {
		t.Errorf("expected state 'uploaded', got %q", asset.State)
	}
}

func TestUser_JSONParsing(t *testing.T) {
	jsonData := `{
		"login": "testuser",
		"id": 12345,
		"node_id": "MDQ6VXNlcjEyMzQ1",
		"avatar_url": "https://avatars.githubusercontent.com/u/12345",
		"type": "User",
		"site_admin": false
	}`

	var user User
	err := json.Unmarshal([]byte(jsonData), &user)
	if err != nil {
		t.Fatalf("failed to parse user JSON: %v", err)
	}

	if user.Login != "testuser" {
		t.Errorf("expected login 'testuser', got %q", user.Login)
	}
	if user.ID != 12345 {
		t.Errorf("expected id 12345, got %d", user.ID)
	}
	if user.Type != "User" {
		t.Errorf("expected type 'User', got %q", user.Type)
	}
}

func TestRelease_WithAuthorAndUploader(t *testing.T) {
	jsonData := `{
		"tag_name": "v1.0.0",
		"author": {
			"login": "releaseauthor",
			"id": 111
		},
		"assets": [
			{
				"name": "file.zip",
				"size": 1000,
				"uploader": {
					"login": "uploader",
					"id": 222
				}
			}
		]
	}`

	var release Release
	err := json.Unmarshal([]byte(jsonData), &release)
	if err != nil {
		t.Fatalf("failed to parse release JSON: %v", err)
	}

	if release.Author == nil {
		t.Fatal("expected author to be set")
	}
	if release.Author.Login != "releaseauthor" {
		t.Errorf("expected author login 'releaseauthor', got %q", release.Author.Login)
	}

	if len(release.Assets) == 0 || release.Assets[0].Uploader == nil {
		t.Fatal("expected asset with uploader")
	}
	if release.Assets[0].Uploader.Login != "uploader" {
		t.Errorf("expected uploader login 'uploader', got %q", release.Assets[0].Uploader.Login)
	}
}

func TestRelease_EmptyAssets(t *testing.T) {
	jsonData := `{
		"tag_name": "v1.0.0",
		"assets": []
	}`

	var release Release
	err := json.Unmarshal([]byte(jsonData), &release)
	if err != nil {
		t.Fatalf("failed to parse release JSON: %v", err)
	}

	if release.Assets == nil {
		t.Error("assets should be empty slice, not nil")
	}
	if len(release.Assets) != 0 {
		t.Errorf("expected 0 assets, got %d", len(release.Assets))
	}
}

func TestRelease_DateParsing(t *testing.T) {
	jsonData := `{
		"tag_name": "v1.0.0",
		"created_at": "2024-06-15T14:30:00Z",
		"published_at": "2024-06-15T15:00:00Z",
		"assets": []
	}`

	var release Release
	err := json.Unmarshal([]byte(jsonData), &release)
	if err != nil {
		t.Fatalf("failed to parse release JSON: %v", err)
	}

	expectedCreated := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	if !release.CreatedAt.Equal(expectedCreated) {
		t.Errorf("expected created_at %v, got %v", expectedCreated, release.CreatedAt)
	}

	expectedPublished := time.Date(2024, 6, 15, 15, 0, 0, 0, time.UTC)
	if !release.PublishedAt.Equal(expectedPublished) {
		t.Errorf("expected published_at %v, got %v", expectedPublished, release.PublishedAt)
	}
}

func TestGetLatestRelease_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/repos/MillQK/d2tool/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Error("missing or incorrect Accept header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tag_name": "0.0.6",
			"name": "Release 0.0.6",
			"assets": [
				{
					"name": "d2tool-windows-amd64.zip",
					"size": 12345678,
					"browser_download_url": "https://example.com/download.zip"
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewHttpClient(server.URL)
	release, err := client.GetLatestRelease()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release.TagName != "0.0.6" {
		t.Errorf("expected tag '0.0.6', got %q", release.TagName)
	}
	if len(release.Assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(release.Assets))
	}
	if release.Assets[0].Name != "d2tool-windows-amd64.zip" {
		t.Errorf("expected asset name 'd2tool-windows-amd64.zip', got %q", release.Assets[0].Name)
	}
}

func TestGetLatestRelease_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	client := NewHttpClient(server.URL)
	_, err := client.GetLatestRelease()

	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestGetLatestRelease_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := NewHttpClient(server.URL)
	_, err := client.GetLatestRelease()

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGetLatestRelease_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := &HttpClient{
		httpClient: &http.Client{Timeout: 10 * time.Millisecond},
		apiUrl:     server.URL,
	}

	_, err := client.GetLatestRelease()
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestGetLatestRelease_MultipleAssets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tag_name": "v0.0.6",
			"assets": [
				{"name": "d2tool-windows-amd64.zip", "size": 1000},
				{"name": "d2tool-darwin-amd64.zip", "size": 2000},
				{"name": "d2tool-darwin-arm64.zip", "size": 1500},
				{"name": "d2tool-linux-amd64.zip", "size": 1800}
			]
		}`))
	}))
	defer server.Close()

	client := NewHttpClient(server.URL)
	release, err := client.GetLatestRelease()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(release.Assets) != 4 {
		t.Errorf("expected 4 assets, got %d", len(release.Assets))
	}

	// Verify we can find platform-specific assets
	foundWindows := false
	foundDarwinArm := false
	for _, asset := range release.Assets {
		if asset.Name == "d2tool-windows-amd64.zip" {
			foundWindows = true
		}
		if asset.Name == "d2tool-darwin-arm64.zip" {
			foundDarwinArm = true
		}
	}
	if !foundWindows {
		t.Error("missing windows asset")
	}
	if !foundDarwinArm {
		t.Error("missing darwin arm64 asset")
	}
}

func TestHttpClient_HasTimeout(t *testing.T) {
	client := NewHttpClient("https://example.com")

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", client.httpClient.Timeout)
	}
}

// Benchmark JSON parsing
func BenchmarkReleaseParsing(b *testing.B) {
	jsonData := []byte(`{
		"tag_name": "v0.0.6",
		"name": "Release v0.0.6",
		"body": "Release notes with lots of text...",
		"assets": [
			{"name": "d2tool-windows-amd64.zip", "size": 12345678},
			{"name": "d2tool-darwin-amd64.zip", "size": 11234567},
			{"name": "d2tool-linux-amd64.zip", "size": 10123456}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var release Release
		json.Unmarshal(jsonData, &release)
	}
}
