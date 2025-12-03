package config

import (
	"d2tool/steam"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const configFileName = "d2tool_config.json"

// FileConfig represents a single config file entry
type FileConfig struct {
	FilePath                  string            `json:"filePath"`
	Enabled                   bool              `json:"enabled"`
	Attributes                map[string]string `json:"attributes"`
	LastUpdateTimestampMillis int64             `json:"lastUpdateTimestampMillis"`
	LastUpdateErrorMessage    string            `json:"lastUpdateErrorMessage"`
}

// PositionConfig represents a position entry
type PositionConfig struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

// HeroesLayoutConfig contains heroes layout related settings
type HeroesLayoutConfig struct {
	Files     []FileConfig     `json:"files"`
	Positions []PositionConfig `json:"positions"`
}

// Config is the main configuration structure
type Config struct {
	mu sync.RWMutex

	HeroesLayout HeroesLayoutConfig `json:"heroesLayout"`
}

func getConfigPath() string {
	execPath, err := os.Executable()
	if err != nil {
		slog.Warn("Error getting executable path", "error", err)
		return configFileName
	}
	return filepath.Join(filepath.Dir(execPath), configFileName)
}

func defaultPositions() []PositionConfig {
	return []PositionConfig{
		{ID: "1", Enabled: true},
		{ID: "2", Enabled: true},
		{ID: "3", Enabled: true},
		{ID: "4", Enabled: true},
		{ID: "5", Enabled: true},
	}
}

func LoadConfig() *Config {
	config := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:     []FileConfig{},
			Positions: defaultPositions(),
		},
	}

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Info("Config file not found, using defaults", "path", configPath)
		// Try to auto-discover Steam paths
		config.setSteamLayoutFiles()
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		slog.Warn("Error parsing config file, using defaults", "error", err)
		config.setSteamLayoutFiles()
		return config
	}

	// Ensure positions exist
	if len(config.HeroesLayout.Positions) == 0 {
		config.HeroesLayout.Positions = defaultPositions()
	}

	config.updateSteamLayoutFileAttributes()

	return config
}

func (c *Config) setSteamLayoutFiles() {
	steamPath, err := steam.FindSteamPath()
	if err != nil {
		slog.Warn("Error finding Steam path", "error", err)
		return
	}

	configFiles, err := steam.FindSteamHeroesLayoutConfigFiles(steamPath)
	if err != nil {
		slog.Warn("Error finding hero grid config files", "error", err)
		return
	}

	for _, configFile := range configFiles {
		c.HeroesLayout.Files = append(c.HeroesLayout.Files,
			FileConfig{
				FilePath:                  configFile.Path,
				Enabled:                   true,
				Attributes:                configFile.ToAttributesMap(),
				LastUpdateTimestampMillis: 0,
				LastUpdateErrorMessage:    "",
			},
		)
	}
}

func (c *Config) updateSteamLayoutFileAttributes() {
	steamPath, err := steam.FindSteamPath()
	if err != nil {
		slog.Warn("Error finding Steam path", "error", err)
		return
	}

	configFiles, err := steam.FindSteamHeroesLayoutConfigFiles(steamPath)
	if err != nil {
		slog.Warn("Error finding hero grid config files", "error", err)
		return
	}

	pathToConfigFile := map[string]steam.SteamHeroesLayoutConfigFileInfo{}
	for _, configFile := range configFiles {
		pathToConfigFile[configFile.Path] = configFile
	}

	for i := range c.HeroesLayout.Files {
		if configFile, ok := pathToConfigFile[c.HeroesLayout.Files[i].FilePath]; ok {
			c.HeroesLayout.Files[i].Attributes = configFile.ToAttributesMap()
		}
	}
}

func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0644)
}

// --- Heroes Layout File Methods ---

func (c *Config) GetHeroesLayoutFiles() []FileConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]FileConfig, len(c.HeroesLayout.Files))
	copy(result, c.HeroesLayout.Files)
	return result
}

func (c *Config) SetHeroesLayoutFiles(files []FileConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.HeroesLayout.Files = files
	go c.Save()
}

func (c *Config) AddHeroesLayoutFile(filePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check for duplicates
	for _, f := range c.HeroesLayout.Files {
		if f.FilePath == filePath {
			return
		}
	}

	c.HeroesLayout.Files = append(c.HeroesLayout.Files, FileConfig{
		FilePath:   filePath,
		Enabled:    true,
		Attributes: map[string]string{},
	})
	go c.Save()
}

func (c *Config) RemoveHeroesLayoutFile(index int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index < 0 || index >= len(c.HeroesLayout.Files) {
		return
	}

	c.HeroesLayout.Files = append(c.HeroesLayout.Files[:index], c.HeroesLayout.Files[index+1:]...)
	go c.Save()
}

func (c *Config) SetHeroesLayoutFileEnabled(index int, enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index < 0 || index >= len(c.HeroesLayout.Files) {
		return
	}

	c.HeroesLayout.Files[index].Enabled = enabled
	go c.Save()
}

func (c *Config) UpdateHeroesLayoutFileStatus(filePath string, timestampMillis int64, errorMessage string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.HeroesLayout.Files {
		if c.HeroesLayout.Files[i].FilePath == filePath {
			c.HeroesLayout.Files[i].LastUpdateTimestampMillis = timestampMillis
			c.HeroesLayout.Files[i].LastUpdateErrorMessage = errorMessage
			break
		}
	}
	go c.Save()
}

// --- Heroes Layout Position Methods ---

func (c *Config) GetPositions() []PositionConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]PositionConfig, len(c.HeroesLayout.Positions))
	copy(result, c.HeroesLayout.Positions)
	return result
}

func (c *Config) SetPositions(positions []PositionConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.HeroesLayout.Positions = positions
	go c.Save()
}

func (c *Config) SetPositionEnabled(id string, enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.HeroesLayout.Positions {
		if c.HeroesLayout.Positions[i].ID == id {
			c.HeroesLayout.Positions[i].Enabled = enabled
			break
		}
	}
	go c.Save()
}

// --- Helper Methods for Update Logic ---

// GetEnabledFilePaths returns only enabled file paths
func (c *Config) GetEnabledFilePaths() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var paths []string
	for _, f := range c.HeroesLayout.Files {
		if f.Enabled {
			paths = append(paths, f.FilePath)
		}
	}
	return paths
}

// GetEnabledPositionIDs returns only enabled position IDs in order
func (c *Config) GetEnabledPositionIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var ids []string
	for _, p := range c.HeroesLayout.Positions {
		if p.Enabled {
			ids = append(ids, p.ID)
		}
	}
	return ids
}
