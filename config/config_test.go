package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Helper to create a test config with custom save path
func newTestConfig(t *testing.T, configPath string) *Config {
	t.Helper()
	return &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:     []FileConfig{},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond, // Short delay for tests
	}
}

func TestDefaultPositions(t *testing.T) {
	positions := defaultPositions()

	if len(positions) != 5 {
		t.Errorf("expected 5 default positions, got %d", len(positions))
	}

	for i, pos := range positions {
		expectedID := string(rune('1' + i))
		if pos.ID != expectedID {
			t.Errorf("position %d: expected ID %q, got %q", i, expectedID, pos.ID)
		}
		if !pos.Enabled {
			t.Errorf("position %d: expected Enabled=true", i)
		}
	}
}

func TestConfig_AddHeroesLayoutFile(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:     []FileConfig{},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	cfg.AddHeroesLayoutFile("/path/to/file1.json")

	files := cfg.GetHeroesLayoutFiles()
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].FilePath != "/path/to/file1.json" {
		t.Errorf("expected path '/path/to/file1.json', got %q", files[0].FilePath)
	}
	if !files[0].Enabled {
		t.Error("new file should be enabled by default")
	}
}

func TestConfig_AddHeroesLayoutFile_PreventsDuplicates(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:     []FileConfig{},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	cfg.AddHeroesLayoutFile("/path/to/file.json")
	cfg.AddHeroesLayoutFile("/path/to/file.json") // Duplicate

	files := cfg.GetHeroesLayoutFiles()
	if len(files) != 1 {
		t.Errorf("expected 1 file (no duplicates), got %d", len(files))
	}
}

func TestConfig_RemoveHeroesLayoutFile(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{FilePath: "/file1.json", Enabled: true},
				{FilePath: "/file2.json", Enabled: true},
				{FilePath: "/file3.json", Enabled: true},
			},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	cfg.RemoveHeroesLayoutFile(1) // Remove middle file

	files := cfg.GetHeroesLayoutFiles()
	if len(files) != 2 {
		t.Fatalf("expected 2 files after removal, got %d", len(files))
	}
	if files[0].FilePath != "/file1.json" || files[1].FilePath != "/file3.json" {
		t.Error("wrong files remaining after removal")
	}
}

func TestConfig_RemoveHeroesLayoutFile_InvalidIndex(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{FilePath: "/file1.json", Enabled: true},
			},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	// Should not panic with invalid indices
	cfg.RemoveHeroesLayoutFile(-1)
	cfg.RemoveHeroesLayoutFile(100)

	files := cfg.GetHeroesLayoutFiles()
	if len(files) != 1 {
		t.Error("file should not be removed with invalid index")
	}
}

func TestConfig_SetHeroesLayoutFileEnabled(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{FilePath: "/file1.json", Enabled: true},
			},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	cfg.SetHeroesLayoutFileEnabled(0, false)

	files := cfg.GetHeroesLayoutFiles()
	if files[0].Enabled {
		t.Error("file should be disabled")
	}

	cfg.SetHeroesLayoutFileEnabled(0, true)

	files = cfg.GetHeroesLayoutFiles()
	if !files[0].Enabled {
		t.Error("file should be enabled")
	}
}

func TestConfig_SetPositionEnabled(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:     []FileConfig{},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	cfg.SetPositionEnabled("1", false)

	positions := cfg.GetPositions()
	for _, pos := range positions {
		if pos.ID == "1" && pos.Enabled {
			t.Error("position 1 should be disabled")
		}
	}
}

func TestConfig_SetPositions(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:     []FileConfig{},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	newPositions := []PositionConfig{
		{ID: "5", Enabled: true},
		{ID: "4", Enabled: false},
		{ID: "3", Enabled: true},
	}

	cfg.SetPositions(newPositions)

	positions := cfg.GetPositions()
	if len(positions) != 3 {
		t.Fatalf("expected 3 positions, got %d", len(positions))
	}
	if positions[0].ID != "5" {
		t.Error("position order not preserved")
	}
}

func TestConfig_GetEnabledFilePaths(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{FilePath: "/enabled1.json", Enabled: true},
				{FilePath: "/disabled.json", Enabled: false},
				{FilePath: "/enabled2.json", Enabled: true},
			},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	paths := cfg.GetEnabledFilePaths()

	if len(paths) != 2 {
		t.Fatalf("expected 2 enabled paths, got %d", len(paths))
	}
	if paths[0] != "/enabled1.json" || paths[1] != "/enabled2.json" {
		t.Error("wrong enabled paths returned")
	}
}

func TestConfig_GetEnabledPositionIDs(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{},
			Positions: []PositionConfig{
				{ID: "1", Enabled: true},
				{ID: "2", Enabled: false},
				{ID: "3", Enabled: true},
			},
		},
		saveDelay: 50 * time.Millisecond,
	}

	ids := cfg.GetEnabledPositionIDs()

	if len(ids) != 2 {
		t.Fatalf("expected 2 enabled IDs, got %d", len(ids))
	}
	if ids[0] != "1" || ids[1] != "3" {
		t.Error("wrong enabled IDs returned")
	}
}

func TestConfig_UpdateHeroesLayoutFileStatus(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{FilePath: "/file1.json", Enabled: true},
				{FilePath: "/file2.json", Enabled: true},
			},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	timestamp := time.Now().UnixMilli()
	cfg.UpdateHeroesLayoutFileStatus([]string{"/file1.json"}, timestamp, "")

	files := cfg.GetHeroesLayoutFiles()
	if files[0].LastUpdateTimestampMillis != timestamp {
		t.Error("timestamp not updated")
	}
	if files[1].LastUpdateTimestampMillis != 0 {
		t.Error("other file's timestamp should not be updated")
	}
}

func TestConfig_UpdateHeroesLayoutFileStatus_WithError(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{FilePath: "/file1.json", Enabled: true},
			},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	errorMsg := "failed to update"
	cfg.UpdateHeroesLayoutFileStatus([]string{"/file1.json"}, 0, errorMsg)

	files := cfg.GetHeroesLayoutFiles()
	if files[0].LastUpdateErrorMessage != errorMsg {
		t.Errorf("expected error message %q, got %q", errorMsg, files[0].LastUpdateErrorMessage)
	}
}

func TestConfig_GetHeroesLayoutFiles_ReturnsCopy(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{FilePath: "/file1.json", Enabled: true},
			},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	files1 := cfg.GetHeroesLayoutFiles()
	files1[0].FilePath = "/modified.json"

	files2 := cfg.GetHeroesLayoutFiles()
	if files2[0].FilePath == "/modified.json" {
		t.Error("GetHeroesLayoutFiles should return a copy, not the original")
	}
}

func TestConfig_GetPositions_ReturnsCopy(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:     []FileConfig{},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	positions1 := cfg.GetPositions()
	positions1[0].ID = "modified"

	positions2 := cfg.GetPositions()
	if positions2[0].ID == "modified" {
		t.Error("GetPositions should return a copy, not the original")
	}
}

func TestConfig_SaveNow(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{FilePath: "/test.json", Enabled: true},
			},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	// We can't easily test SaveNow without modifying getConfigPath
	// This test documents the expected interface
	_ = configPath
	_ = cfg
}

func TestConfig_ConcurrentAccess(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:     []FileConfig{},
			Positions: defaultPositions(),
		},
		saveDelay: 50 * time.Millisecond,
	}

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			cfg.AddHeroesLayoutFile("/file" + string(rune(i)) + ".json")
		}
	}()

	// Concurrent readers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = cfg.GetHeroesLayoutFiles()
		}
	}()

	// Concurrent position updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			cfg.SetPositionEnabled("1", i%2 == 0)
		}
	}()

	wg.Wait()
	// Test passes if no race conditions or panics occur
}

func TestConfig_JSONMarshal(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{
					FilePath:                  "/test.json",
					Enabled:                   true,
					Attributes:                map[string]string{"key": "value"},
					LastUpdateTimestampMillis: 1234567890,
					LastUpdateErrorMessage:    "",
				},
			},
			Positions: []PositionConfig{
				{ID: "1", Enabled: true},
				{ID: "2", Enabled: false},
			},
		},
		saveDelay: 50 * time.Millisecond,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if len(decoded.HeroesLayout.Files) != 1 {
		t.Error("files not preserved in JSON roundtrip")
	}
	if decoded.HeroesLayout.Files[0].FilePath != "/test.json" {
		t.Error("file path not preserved in JSON roundtrip")
	}
	if len(decoded.HeroesLayout.Positions) != 2 {
		t.Error("positions not preserved in JSON roundtrip")
	}
}

func TestConfig_JSONMarshal_ExcludesPrivateFields(t *testing.T) {
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:     []FileConfig{},
			Positions: defaultPositions(),
		},
		saveDelay: 500 * time.Millisecond,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Private fields should not appear in JSON
	jsonStr := string(data)
	if contains(jsonStr, "saveDelay") {
		t.Error("saveDelay should not be in JSON output")
	}
	if contains(jsonStr, "saveTimer") {
		t.Error("saveTimer should not be in JSON output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Integration test for save file format
func TestConfig_SaveFileFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create config data directly
	cfg := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files: []FileConfig{
				{FilePath: "/steam/config.json", Enabled: true},
			},
			Positions: []PositionConfig{
				{ID: "1", Enabled: true},
				{ID: "2", Enabled: false},
			},
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Read back and verify structure
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var loaded Config
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatal(err)
	}

	if len(loaded.HeroesLayout.Files) != 1 {
		t.Error("files not loaded correctly")
	}
	if len(loaded.HeroesLayout.Positions) != 2 {
		t.Error("positions not loaded correctly")
	}
	if loaded.HeroesLayout.Positions[1].Enabled {
		t.Error("position 2 should be disabled")
	}
}
