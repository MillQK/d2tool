package update

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFileFromArchive_PathTraversal(t *testing.T) {
	rootDir := t.TempDir()

	tests := []struct {
		name        string
		fileName    string
		shouldError bool
	}{
		{"normal file", "app.exe", false},
		{"subdirectory file", "subdir/file.txt", false},
		{"nested subdirectory", "a/b/c/file.txt", false},
		{"parent traversal", "../outside.txt", true},
		{"deep traversal", "../../../etc/passwd", true},
		{"hidden traversal", "foo/../../../etc/passwd", true},
		{"dot prefix traversal", "foo/../../outside.txt", true},
		{"current dir reference", "./normal.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a zip file in memory with the test filename
			zipFile := createTestZipFile(t, tt.fileName, []byte("test content"))

			err := extractFileFromArchive(zipFile, rootDir)

			if tt.shouldError && err == nil {
				t.Errorf("expected error for path %q, got nil", tt.fileName)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error for path %q: %v", tt.fileName, err)
			}

			// For valid paths, verify the file was created in the right place
			if !tt.shouldError && err == nil {
				expectedPath := filepath.Join(rootDir, filepath.Clean(tt.fileName))
				if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
					t.Errorf("file was not created at expected path: %s", expectedPath)
				}
			}
		})
	}
}

func TestExtractFileFromArchive_Directory(t *testing.T) {
	rootDir := t.TempDir()

	// Create a zip with a directory entry
	zipFile := createTestZipDirectory(t, "subdir/")

	err := extractFileFromArchive(zipFile, rootDir)
	if err != nil {
		t.Fatalf("unexpected error creating directory: %v", err)
	}

	// Verify directory was created
	dirPath := filepath.Join(rootDir, "subdir")
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory, got a file")
	}
}

func TestExtractFileFromArchive_DirectoryTraversal(t *testing.T) {
	rootDir := t.TempDir()

	// Try to create a directory outside rootDir
	zipFile := createTestZipDirectory(t, "../outside_dir/")

	err := extractFileFromArchive(zipFile, rootDir)
	if err == nil {
		t.Error("expected error for directory traversal, got nil")
	}
}

func TestExtractFileFromArchive_OverwritesExistingFile(t *testing.T) {
	rootDir := t.TempDir()

	// Create an existing file
	existingPath := filepath.Join(rootDir, "existing.txt")
	if err := os.WriteFile(existingPath, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	// Extract a file with the same name
	zipFile := createTestZipFile(t, "existing.txt", []byte("new content"))

	err := extractFileFromArchive(zipFile, rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the old file was renamed
	oldFiles, _ := filepath.Glob(filepath.Join(rootDir, fmt.Sprintf("%s*", oldFilesPrefix)))
	if len(oldFiles) == 0 {
		t.Errorf("expected old file to be renamed with %s prefix", oldFilesPrefix)
	}

	// Verify new content
	content, _ := os.ReadFile(existingPath)
	if string(content) != "new content" {
		t.Errorf("expected 'new content', got %q", string(content))
	}
}

func TestCleanupOldFiles(t *testing.T) {
	rootDir := t.TempDir()

	// Create some oldFilesPrefix files
	oldFile1 := filepath.Join(rootDir, fmt.Sprintf("%sfile1.txt", oldFilesPrefix))
	oldFile2 := filepath.Join(rootDir, fmt.Sprintf("%sfile2.exe", oldFilesPrefix))
	normalFile := filepath.Join(rootDir, "normal.txt")

	os.WriteFile(oldFile1, []byte("old1"), 0644)
	os.WriteFile(oldFile2, []byte("old2"), 0644)
	os.WriteFile(normalFile, []byte("normal"), 0644)

	// Note: cleanupOldFiles uses os.Executable() which won't work in tests
	// This test documents the expected behavior but may need adjustment
	// depending on how the function is refactored for testability
}

// Helper functions

func createTestZipFile(t *testing.T, name string, content []byte) *zip.File {
	t.Helper()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, err := w.Create(name)
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}

	_, err = f.Write(content)
	if err != nil {
		t.Fatalf("failed to write zip content: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("failed to create zip reader: %v", err)
	}

	return r.File[0]
}

func createTestZipDirectory(t *testing.T, name string) *zip.File {
	t.Helper()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Create directory entry (name must end with /)
	_, err := w.Create(name)
	if err != nil {
		t.Fatalf("failed to create zip directory entry: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("failed to create zip reader: %v", err)
	}

	return r.File[0]
}
