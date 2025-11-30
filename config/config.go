package config

import (
	"d2tool/heroesGrid"
	"d2tool/steam"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const configFileName = "d2tool_config.json"

type Config struct {
	mu sync.RWMutex

	HeroesGridFilePaths []string `json:"heroesGridFilePaths"`
	PositionsOrder      []string `json:"positionsOrder"`

	LastUpdateTimestampMillis         int64  `json:"lastUpdateTimestampMillis"`
	LastUpdateErrorMessage            string `json:"lastUpdateErrorMessage"`
	AppLastUpdateCheckTimestampMillis int64  `json:"appLastUpdateCheckTimestampMillis"`
	AutoUpdateEnabled                 *bool  `json:"autoUpdateEnabled"`
}

func getConfigPath() string {
	execPath, err := os.Executable()
	if err != nil {
		slog.Warn("Error getting executable path", "error", err)
		return configFileName
	}
	return filepath.Join(filepath.Dir(execPath), configFileName)
}

func LoadConfig() *Config {
	config := &Config{
		PositionsOrder: []string{"1", "2", "3", "4", "5"},
	}

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Info("Config file not found, using defaults", "path", configPath)
		// Try to auto-discover Steam paths
		config.discoverDefaultPaths()
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		slog.Warn("Error parsing config file, using defaults", "error", err)
		config.discoverDefaultPaths()
		return config
	}

	// If no paths configured, try to discover them
	if len(config.HeroesGridFilePaths) == 0 {
		config.discoverDefaultPaths()
	}

	return config
}

func (c *Config) discoverDefaultPaths() {
	steamPath, err := steam.FindSteamPath()
	if err != nil {
		slog.Warn("Error finding Steam path", "error", err)
		return
	}

	paths, err := heroesGrid.FindHeroGridConfigFiles(steamPath)
	if err != nil {
		slog.Warn("Error finding hero grid config files", "error", err)
		return
	}

	c.HeroesGridFilePaths = paths
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

// Getters and setters with mutex protection

func (c *Config) GetHeroesGridFilePaths() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.HeroesGridFilePaths))
	copy(result, c.HeroesGridFilePaths)
	return result
}

func (c *Config) SetHeroesGridFilePaths(paths []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.HeroesGridFilePaths = paths
	go c.Save()
}

func (c *Config) GetPositionsOrder() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.PositionsOrder))
	copy(result, c.PositionsOrder)
	return result
}

func (c *Config) SetPositionsOrder(positions []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PositionsOrder = positions
	go c.Save()
}

func (c *Config) GetLastUpdateTimestampMillis() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastUpdateTimestampMillis
}

func (c *Config) SetLastUpdateTimestampMillis(millis int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastUpdateTimestampMillis = millis
	go c.Save()
}

func (c *Config) GetLastUpdateErrorMessage() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastUpdateErrorMessage
}

func (c *Config) SetLastUpdateErrorMessage(msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastUpdateErrorMessage = msg
	go c.Save()
}

func (c *Config) GetAppLastUpdateCheckTimestampMillis() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AppLastUpdateCheckTimestampMillis
}

func (c *Config) SetAppLastUpdateCheckTimestampMillis(millis int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AppLastUpdateCheckTimestampMillis = millis
	go c.Save()
}

func (c *Config) GetAutoUpdateEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Default to true if not set
	if c.AutoUpdateEnabled == nil {
		return true
	}
	return *c.AutoUpdateEnabled
}

func (c *Config) SetAutoUpdateEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AutoUpdateEnabled = &enabled
	go c.Save()
}
