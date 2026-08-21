package update

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"d2tool/github"
)

func TestSelectReleaseAsset(t *testing.T) {
	release := &github.Release{Assets: []github.ReleaseAsset{
		{Name: "d2tool-linux-amd64.zip"},
		{Name: "d2tool-windows-amd64.zip"},
	}}
	asset, err := selectReleaseAsset(release, "windows", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Name != "d2tool-windows-amd64.zip" {
		t.Fatalf("selected %q", asset.Name)
	}
	if _, err := selectReleaseAsset(release, "darwin", "amd64"); err == nil {
		t.Fatal("missing matching asset must fail")
	}

	release.Assets = append(release.Assets, github.ReleaseAsset{Name: "d2tool-windows-amd64.zip"})
	if _, err := selectReleaseAsset(release, "windows", "amd64"); err == nil {
		t.Fatal("duplicate matching assets must fail")
	}
}

func TestDownloadReleaseArchive_ValidatesSizeAndDigest(t *testing.T) {
	body := []byte("archive bytes")
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256(body))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/octet-stream" {
			t.Errorf("missing asset Accept header")
		}
		_, _ = w.Write(body)
	}))
	defer server.Close()

	asset := github.ReleaseAsset{URL: server.URL, Size: int64(len(body)), Digest: digest}
	path, err := downloadReleaseArchive(context.Background(), http.DefaultClient, asset, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(body) {
		t.Fatalf("downloaded %q, want %q", got, body)
	}

	asset.Size++
	sizeMismatchDir := t.TempDir()
	if _, err := downloadReleaseArchive(context.Background(), http.DefaultClient, asset, sizeMismatchDir); err == nil {
		t.Fatal("size mismatch must fail")
	}
	assertDirectoryEmpty(t, sizeMismatchDir)
	asset.Size--
	asset.Digest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	digestMismatchDir := t.TempDir()
	if _, err := downloadReleaseArchive(context.Background(), http.DefaultClient, asset, digestMismatchDir); err == nil {
		t.Fatal("digest mismatch must fail")
	}
	assertDirectoryEmpty(t, digestMismatchDir)
	asset.Digest = ""
	pathWithoutDigest, err := downloadReleaseArchive(context.Background(), http.DefaultClient, asset, t.TempDir())
	if err != nil {
		t.Fatalf("asset without digest must remain backward compatible: %v", err)
	}
	defer os.Remove(pathWithoutDigest)
}

func TestDownloadReleaseArchive_RejectsHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	asset := github.ReleaseAsset{URL: server.URL, Size: 1}
	if _, err := downloadReleaseArchive(context.Background(), http.DefaultClient, asset, t.TempDir()); err == nil {
		t.Fatal("non-200 response must fail")
	}
}

type httpDoerFunc func(*http.Request) (*http.Response, error)

func (f httpDoerFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestDownloadReleaseArchive_PropagatesTimeout(t *testing.T) {
	client := httpDoerFunc(func(*http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})
	_, err := downloadReleaseArchive(
		context.Background(),
		client,
		github.ReleaseAsset{URL: "https://example.invalid/release.zip", Size: 1},
		t.TempDir(),
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want context deadline exceeded", err)
	}
}

func TestDownloadReleaseArchive_HonorsCancellationAndCleansUp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	asset := github.ReleaseAsset{URL: server.URL, Size: 1}
	_, err := downloadReleaseArchive(ctx, http.DefaultClient, asset, dir)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	assertDirectoryEmpty(t, dir)
}

func TestDownloadReleaseArchive_RejectsNonPositiveSizeBeforeNetwork(t *testing.T) {
	for _, size := range []int64{0, -1} {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			calls := 0
			client := httpDoerFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("network must not be called")
			})
			dir := t.TempDir()

			_, err := downloadReleaseArchive(
				context.Background(),
				client,
				github.ReleaseAsset{URL: "https://example.invalid/release.zip", Size: size},
				dir,
			)
			if err == nil || !strings.Contains(err.Error(), "size") {
				t.Fatalf("downloadReleaseArchive() error = %v, want invalid size", err)
			}
			if calls != 0 {
				t.Fatalf("network calls = %d, want 0", calls)
			}
			assertDirectoryEmpty(t, dir)
		})
	}
}

type countingReadCloser struct {
	reader io.Reader
	read   int
}

func (r *countingReadCloser) Read(buffer []byte) (int, error) {
	n, err := r.reader.Read(buffer)
	r.read += n
	return n, err
}

func (*countingReadCloser) Close() error { return nil }

func TestDownloadReleaseArchive_BoundsOversizedResponse(t *testing.T) {
	const declaredSize = int64(8)
	body := &countingReadCloser{reader: strings.NewReader(strings.Repeat("x", 4096))}
	client := httpDoerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       body,
		}, nil
	})
	dir := t.TempDir()

	_, err := downloadReleaseArchive(
		context.Background(),
		client,
		github.ReleaseAsset{URL: "https://example.invalid/release.zip", Size: declaredSize},
		dir,
	)
	if err == nil || !strings.Contains(err.Error(), "size mismatch") {
		t.Fatalf("downloadReleaseArchive() error = %v, want size mismatch", err)
	}
	if body.read != int(declaredSize)+1 {
		t.Fatalf("response bytes read = %d, want %d", body.read, declaredSize+1)
	}
	assertDirectoryEmpty(t, dir)
}

func writeArchive(t *testing.T, entries map[string]string, modes map[string]os.FileMode) string {
	t.Helper()
	archivePath := filepath.Join(t.TempDir(), "release.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, contents := range entries {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(modes[name])
		entry, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return archivePath
}

func TestExtractReleaseExecutable(t *testing.T) {
	tests := []struct {
		name    string
		entries map[string]string
		modes   map[string]os.FileMode
		goos    string
		wantErr string
	}{
		{name: "linux executable", entries: map[string]string{"./d2tool": "new"}, modes: map[string]os.FileMode{"./d2tool": 0755}, goos: "linux"},
		{name: "windows executable", entries: map[string]string{"d2tool.exe": "new"}, modes: map[string]os.FileMode{"d2tool.exe": 0644}, goos: "windows"},
		{name: "traversal", entries: map[string]string{"../d2tool": "new"}, modes: map[string]os.FileMode{"../d2tool": 0755}, goos: "linux", wantErr: "illegal archive path"},
		{name: "unexpected regular file", entries: map[string]string{"d2tool": "new", "notes.txt": "no"}, modes: map[string]os.FileMode{"d2tool": 0755, "notes.txt": 0644}, goos: "linux", wantErr: "unexpected archive entry"},
		{name: "missing executable", entries: map[string]string{}, modes: map[string]os.FileMode{}, goos: "linux", wantErr: "expected executable"},
		{name: "linux mode is not executable", entries: map[string]string{"d2tool": "new"}, modes: map[string]os.FileMode{"d2tool": 0644}, goos: "linux", wantErr: "not executable"},
		{name: "symlink", entries: map[string]string{"d2tool": "target"}, modes: map[string]os.FileMode{"d2tool": os.ModeSymlink | 0777}, goos: "linux", wantErr: "symlink"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archivePath := writeArchive(t, tt.entries, tt.modes)
			dir := t.TempDir()
			stagedPath, err := extractReleaseExecutable(context.Background(), archivePath, dir, tt.goos)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
				}
				assertDirectoryEmpty(t, dir)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(stagedPath)
			contents, err := os.ReadFile(stagedPath)
			if err != nil || string(contents) != "new" {
				t.Fatalf("staged contents = %q, error = %v", contents, err)
			}
		})
	}
}

type cancelAfterReadChecksContext struct {
	context.Context
	checks int
	done   chan struct{}
}

func (c *cancelAfterReadChecksContext) Err() error {
	c.checks++
	if c.checks >= 2 {
		if c.checks == 2 {
			close(c.done)
		}
		return context.Canceled
	}
	return nil
}

func (c *cancelAfterReadChecksContext) Done() <-chan struct{} { return c.done }

func TestExtractReleaseExecutable_HonorsCancellationDuringStaging(t *testing.T) {
	contents := strings.Repeat("new executable bytes", 32*1024)
	archivePath := writeArchive(
		t,
		map[string]string{"d2tool": contents},
		map[string]os.FileMode{"d2tool": 0755},
	)
	dir := t.TempDir()
	ctx := &cancelAfterReadChecksContext{Context: context.Background(), done: make(chan struct{})}

	_, err := extractReleaseExecutable(ctx, archivePath, dir, "linux")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("extractReleaseExecutable() error = %v, want context.Canceled", err)
	}
	if ctx.checks < 2 {
		t.Fatalf("context checks = %d, want cancellation checked while copying", ctx.checks)
	}
	assertDirectoryEmpty(t, dir)
}

func assertDirectoryEmpty(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("temporary files remain: entries=%v error=%v", entries, err)
	}
}
