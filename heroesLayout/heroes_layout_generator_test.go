package heroesLayout

import (
	"d2tool/providers"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHeroesLayoutConfigs_CreatesD2TPrefix(t *testing.T) {
	positions := []string{"1", "2"}
	positionToHeroes := map[string][]providers.Hero{
		"1": {
			{HeroID: 1, HeroName: "Anti-Mage", D2PTRating: 100, Matches: 500, Wins: 275},
			{HeroID: 2, HeroName: "Juggernaut", D2PTRating: 90, Matches: 400, Wins: 200},
		},
		"2": {
			{HeroID: 10, HeroName: "Shadow Fiend", D2PTRating: 110, Matches: 600, Wins: 330},
		},
	}

	configs := generateHeroesLayoutConfigs(positions, positionToHeroes)

	if len(configs) == 0 {
		t.Fatal("expected at least one config")
	}

	// Check that config name has [D2T] prefix
	if !strings.HasPrefix(configs[0].ConfigName, d2tPrefix) {
		t.Errorf("expected config name to start with %q, got %q", d2tPrefix, configs[0].ConfigName)
	}
}

func TestGenerateHeroesLayoutConfigs_CreatesCategories(t *testing.T) {
	positions := []string{"1"}
	positionToHeroes := map[string][]providers.Hero{
		"1": {
			{HeroID: 1, D2PTRating: 100, Matches: 500, Wins: 250},
			{HeroID: 2, D2PTRating: 90, Matches: 400, Wins: 200},
		},
	}

	configs := generateHeroesLayoutConfigs(positions, positionToHeroes)

	if len(configs[0].Categories) == 0 {
		t.Fatal("expected categories to be created")
	}

	// Should have categories for "Top Rating Heroes" and "Most Matches Heroes"
	hasTopRating := false
	hasMostMatches := false
	for _, cat := range configs[0].Categories {
		if strings.Contains(cat.CategoryName, "Top Rating Heroes") {
			hasTopRating = true
		}
		if strings.Contains(cat.CategoryName, "Most Matches Heroes") {
			hasMostMatches = true
		}
	}

	if !hasTopRating {
		t.Error("expected 'Top Rating Heroes' category")
	}
	if !hasMostMatches {
		t.Error("expected 'Most Matches Heroes' category")
	}
}

func TestGenerateHeroesLayoutConfigs_CalculatesWinrate(t *testing.T) {
	positions := []string{"1"}
	positionToHeroes := map[string][]providers.Hero{
		"1": {
			{HeroID: 1, D2PTRating: 100, Matches: 100, Wins: 55}, // 55% winrate
		},
	}

	configs := generateHeroesLayoutConfigs(positions, positionToHeroes)

	// Find a category with winrate percentage
	foundWinrate := false
	for _, cat := range configs[0].Categories {
		if strings.Contains(cat.CategoryName, "55.0%") {
			foundWinrate = true
			break
		}
	}

	if !foundWinrate {
		t.Error("expected to find 55.0% winrate in categories")
	}
}

func TestProcessHeroesLayoutConfig_PreservesUserConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "hero_grid_config.json")

	// Create existing config with user-defined layout
	existingConfig := heroGridConfig{
		Version: 1,
		Configs: []heroGridCategory{
			{ConfigName: "My Custom Grid", Categories: []heroGridPosition{}},
			{ConfigName: "Another User Grid", Categories: []heroGridPosition{}},
		},
	}
	data, _ := json.MarshalIndent(existingConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	// Process with new heroes data
	positions := []string{"1"}
	positionToHeroes := map[string][]providers.Hero{
		"1": {{HeroID: 1, D2PTRating: 100, Matches: 100, Wins: 50}},
	}

	err := processHeroesLayoutConfig(configPath, positions, positionToHeroes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back and verify user configs are preserved
	resultData, _ := os.ReadFile(configPath)
	var resultConfig heroGridConfig
	json.Unmarshal(resultData, &resultConfig)

	userConfigCount := 0
	for _, cfg := range resultConfig.Configs {
		if !strings.HasPrefix(cfg.ConfigName, d2tPrefix) {
			userConfigCount++
		}
	}

	if userConfigCount != 2 {
		t.Errorf("expected 2 user configs to be preserved, got %d", userConfigCount)
	}
}

func TestProcessHeroesLayoutConfig_RemovesOldD2TConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "hero_grid_config.json")

	// Create existing config with old D2T entry
	existingConfig := heroGridConfig{
		Version: 1,
		Configs: []heroGridCategory{
			{ConfigName: "[D2T] Heroes Meta 2024-01-01", Categories: []heroGridPosition{}},
			{ConfigName: "[D2T] Heroes Meta 2024-01-02", Categories: []heroGridPosition{}},
			{ConfigName: "My Custom Grid", Categories: []heroGridPosition{}},
		},
	}
	data, _ := json.MarshalIndent(existingConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	// Process with new heroes data
	positions := []string{"1"}
	positionToHeroes := map[string][]providers.Hero{
		"1": {{HeroID: 1, D2PTRating: 100, Matches: 100, Wins: 50}},
	}

	err := processHeroesLayoutConfig(configPath, positions, positionToHeroes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back and verify old D2T configs were removed
	resultData, _ := os.ReadFile(configPath)
	var resultConfig heroGridConfig
	json.Unmarshal(resultData, &resultConfig)

	d2tConfigCount := 0
	for _, cfg := range resultConfig.Configs {
		if strings.HasPrefix(cfg.ConfigName, d2tPrefix) {
			d2tConfigCount++
			// Should not be the old dates
			if strings.Contains(cfg.ConfigName, "2024-01-01") || strings.Contains(cfg.ConfigName, "2024-01-02") {
				t.Error("old D2T config should have been removed")
			}
		}
	}

	if d2tConfigCount != 1 {
		t.Errorf("expected exactly 1 D2T config after processing, got %d", d2tConfigCount)
	}
}

func TestProcessHeroesLayoutConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "hero_grid_config.json")

	// Write invalid JSON
	os.WriteFile(configPath, []byte("not valid json"), 0644)

	positions := []string{"1"}
	positionToHeroes := map[string][]providers.Hero{
		"1": {{HeroID: 1}},
	}

	err := processHeroesLayoutConfig(configPath, positions, positionToHeroes)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestProcessHeroesLayoutConfig_FileNotFound(t *testing.T) {
	positions := []string{"1"}
	positionToHeroes := map[string][]providers.Hero{
		"1": {{HeroID: 1}},
	}

	err := processHeroesLayoutConfig("/nonexistent/path/config.json", positions, positionToHeroes)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestProcessHeroesLayoutConfig_PreservesVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "hero_grid_config.json")

	existingConfig := heroGridConfig{
		Version: 42, // Custom version number
		Configs: []heroGridCategory{},
	}
	data, _ := json.MarshalIndent(existingConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	positions := []string{"1"}
	positionToHeroes := map[string][]providers.Hero{
		"1": {{HeroID: 1, D2PTRating: 100, Matches: 100, Wins: 50}},
	}

	err := processHeroesLayoutConfig(configPath, positions, positionToHeroes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultData, _ := os.ReadFile(configPath)
	var resultConfig heroGridConfig
	json.Unmarshal(resultData, &resultConfig)

	if resultConfig.Version != 42 {
		t.Errorf("expected version 42 to be preserved, got %d", resultConfig.Version)
	}
}

func TestGenerateHeroesLayoutConfigs_EmptyPositions(t *testing.T) {
	positions := []string{}
	positionToHeroes := map[string][]providers.Hero{}

	configs := generateHeroesLayoutConfigs(positions, positionToHeroes)

	if len(configs) != 1 {
		t.Fatalf("expected 1 config even with empty positions, got %d", len(configs))
	}

	// Should have config with just the name, no hero categories
	if !strings.HasPrefix(configs[0].ConfigName, d2tPrefix) {
		t.Error("config should still have D2T prefix")
	}
}

func TestGenerateHeroesLayoutConfigs_MultiplePositions(t *testing.T) {
	positions := []string{"1", "2", "3", "4", "5"}
	positionToHeroes := map[string][]providers.Hero{}
	for _, pos := range positions {
		positionToHeroes[pos] = []providers.Hero{
			{HeroID: 1, D2PTRating: 100, Matches: 100, Wins: 50},
		}
	}

	configs := generateHeroesLayoutConfigs(positions, positionToHeroes)

	// Count position headers
	positionHeaders := 0
	for _, cat := range configs[0].Categories {
		for _, pos := range positions {
			if strings.HasPrefix(cat.CategoryName, pos+" -") {
				positionHeaders++
				break
			}
		}
	}

	// Should have 2 headers per position (Top Rating + Most Matches)
	expectedHeaders := len(positions) * 2
	if positionHeaders != expectedHeaders {
		t.Errorf("expected %d position headers, got %d", expectedHeaders, positionHeaders)
	}
}
