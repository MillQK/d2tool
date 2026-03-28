package config

import (
	"d2tool/steamid"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const configFileName = "d2tool_config.json"

var heroGridConfigPathRegex = regexp.MustCompile(`userdata/(\d+)/570/remote/cfg/hero_grid_config\.json$`)

// FileConfig represents a single config file entry
type FileConfig struct {
	FilePath                  string `json:"filePath"`
	Enabled                   bool   `json:"enabled"`
	LastUpdateTimestampMillis int64  `json:"lastUpdateTimestampMillis"`
	LastUpdateErrorMessage    string `json:"lastUpdateErrorMessage"`
}

// PositionConfig represents a position entry
type PositionConfig struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

// HeroesLayoutConfig contains heroes layout related settings
type HeroesLayoutConfig struct {
	Files        []FileConfig     `json:"files"`
	Positions    []PositionConfig `json:"positions"`
	HeroesPerRow int              `json:"heroesPerRow"`
}

// D2PTConfig contains Dota2ProTracker provider settings
type D2PTConfig struct {
	Period string `json:"period"` // "8" for last 8 days, "patch" for current patch
}

// SteamConfig contains Steam-related settings
type SteamConfig struct {
	SteamPath             string               `json:"steamPath"`
	AutoEnableNewAccounts bool                 `json:"autoEnableNewAccounts"`
	Accounts              []SteamAccountConfig `json:"accounts"`
}

// SteamAccountConfig represents a single Steam account entry
type SteamAccountConfig struct {
	SteamID64                 string `json:"steamId64"`
	Enabled                   bool   `json:"enabled"`
	LastUpdateTimestampMillis int64  `json:"lastUpdateTimestampMillis"`
	LastUpdateErrorMessage    string `json:"lastUpdateErrorMessage"`
}

func defaultD2PTConfig() D2PTConfig {
	return D2PTConfig{
		Period: "8", // Default to last 8 days
	}
}

// Config is the main configuration structure
type Config struct {
	mu sync.RWMutex

	HeroesLayout HeroesLayoutConfig `json:"heroesLayout"`
	D2PT         D2PTConfig         `json:"d2pt"`
	Steam        SteamConfig        `json:"steam"`

	// Debounce state for save operations (not persisted)
	saveTimer *time.Timer
	saveMu    sync.Mutex
	saveDelay time.Duration
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

const (
	defaultHeroesPerRow = 15
	minHeroesPerRow     = 1
	maxHeroesPerRow     = 50
)

func LoadConfig() *Config {
	config := &Config{
		HeroesLayout: HeroesLayoutConfig{
			Files:        []FileConfig{},
			Positions:    defaultPositions(),
			HeroesPerRow: defaultHeroesPerRow,
		},
		D2PT: defaultD2PTConfig(),
		Steam: SteamConfig{
			AutoEnableNewAccounts: true,
			Accounts:              []SteamAccountConfig{},
		},
		saveDelay: 500 * time.Millisecond,
	}

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Info("Config file not found, using defaults", "path", configPath)
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		slog.Warn("Error parsing config file, using defaults", "error", err)
		return config
	}

	// Check if migration is needed (no "steam" section in the config file)
	var rawConfig map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawConfig); err == nil {
		if _, hasSteam := rawConfig["steam"]; !hasSteam {
			slog.Info("Migrating config: adding steam section from existing files")
			config.migrateToSteamAccounts()
		}
	}

	// Ensure positions exist
	if len(config.HeroesLayout.Positions) == 0 {
		config.HeroesLayout.Positions = defaultPositions()
	}

	// Ensure D2PT config has valid period
	if config.D2PT.Period != "8" && config.D2PT.Period != "patch" {
		config.D2PT.Period = "8"
	}

	// Ensure HeroesPerRow is within valid range
	if config.HeroesLayout.HeroesPerRow < minHeroesPerRow || config.HeroesLayout.HeroesPerRow > maxHeroesPerRow {
		config.HeroesLayout.HeroesPerRow = defaultHeroesPerRow
	}

	// Ensure Steam.Accounts is never nil
	if config.Steam.Accounts == nil {
		config.Steam.Accounts = []SteamAccountConfig{}
	}

	return config
}

func (c *Config) save() error {
	c.mu.RLock()
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0644)
}

// scheduleSave debounces save operations, coalescing rapid changes into a single save.
// Callers should use `go c.scheduleSave()` to avoid holding c.mu while acquiring
// c.saveMu, which would invert the lock order with SaveNow and cause a deadlock.
func (c *Config) scheduleSave() {
	c.saveMu.Lock()
	defer c.saveMu.Unlock()

	// Cancel any pending save
	if c.saveTimer != nil {
		c.saveTimer.Stop()
	}

	// Schedule new save
	c.saveTimer = time.AfterFunc(c.saveDelay, func() {
		if err := c.save(); err != nil {
			slog.Error("Failed to save config", "error", err)
		}
	})
}

// SaveNow cancels any pending debounced save and saves immediately.
// Use this on shutdown to ensure all changes are persisted.
func (c *Config) SaveNow() error {
	c.saveMu.Lock()
	if c.saveTimer != nil {
		c.saveTimer.Stop()
		c.saveTimer = nil
	}
	c.saveMu.Unlock()

	return c.save()
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
	go c.scheduleSave()
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
		FilePath: filePath,
		Enabled:  true,
	})
	go c.scheduleSave()
}

func (c *Config) RemoveHeroesLayoutFile(filePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, f := range c.HeroesLayout.Files {
		if f.FilePath == filePath {
			c.HeroesLayout.Files = append(c.HeroesLayout.Files[:i], c.HeroesLayout.Files[i+1:]...)
			go c.scheduleSave()
			return
		}
	}
}

func (c *Config) SetHeroesLayoutFileEnabled(filePath string, enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, f := range c.HeroesLayout.Files {
		if f.FilePath == filePath {
			c.HeroesLayout.Files[i].Enabled = enabled
			go c.scheduleSave()
			return
		}
	}
}

func (c *Config) UpdateHeroesLayoutFileStatus(filePaths []string, timestampMillis int64, errorMessage string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.HeroesLayout.Files {
		if slices.Contains(filePaths, c.HeroesLayout.Files[i].FilePath) {
			c.HeroesLayout.Files[i].LastUpdateTimestampMillis = timestampMillis
			c.HeroesLayout.Files[i].LastUpdateErrorMessage = errorMessage
		}
	}
	go c.scheduleSave()
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
	go c.scheduleSave()
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
	go c.scheduleSave()
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

// --- D2PT Config Methods ---

// GetD2PTConfig returns the D2PT provider configuration
func (c *Config) GetD2PTConfig() D2PTConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.D2PT
}

// SetD2PTPeriod sets the D2PT period parameter
func (c *Config) SetD2PTPeriod(period string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.D2PT.Period = period
	go c.scheduleSave()
}

// --- Heroes Layout Settings Methods ---

// GetHeroesPerRow returns the configured heroes per row value
func (c *Config) GetHeroesPerRow() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HeroesLayout.HeroesPerRow
}

// SetHeroesPerRow sets the heroes per row value with validation (1-50)
func (c *Config) SetHeroesPerRow(heroesPerRow int) error {
	if heroesPerRow < minHeroesPerRow || heroesPerRow > maxHeroesPerRow {
		return fmt.Errorf("heroesPerRow must be between %d and %d, got %d", minHeroesPerRow, maxHeroesPerRow, heroesPerRow)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.HeroesLayout.HeroesPerRow = heroesPerRow
	go c.scheduleSave()
	return nil
}

// --- Steam Config Methods ---

// GetSteamConfig returns a copy of the Steam configuration
func (c *Config) GetSteamConfig() SteamConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cfg := c.Steam
	cfg.Accounts = make([]SteamAccountConfig, len(c.Steam.Accounts))
	copy(cfg.Accounts, c.Steam.Accounts)
	return cfg
}

// SetSteamPath sets the Steam installation path
func (c *Config) SetSteamPath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Steam.SteamPath = path
	go c.scheduleSave()
}

// SetAutoEnableNewAccounts sets whether newly discovered accounts are auto-enabled
func (c *Config) SetAutoEnableNewAccounts(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Steam.AutoEnableNewAccounts = enabled
	go c.scheduleSave()
}

// GetSteamAccounts returns a copy of the Steam accounts list
func (c *Config) GetSteamAccounts() []SteamAccountConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]SteamAccountConfig, len(c.Steam.Accounts))
	copy(result, c.Steam.Accounts)
	return result
}

// SetSteamAccounts replaces the Steam accounts list
func (c *Config) SetSteamAccounts(accounts []SteamAccountConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Steam.Accounts = accounts
	go c.scheduleSave()
}

// SetSteamAccountEnabled enables or disables a specific Steam account by SteamID64
func (c *Config) SetSteamAccountEnabled(steamId64 string, enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.Steam.Accounts {
		if c.Steam.Accounts[i].SteamID64 == steamId64 {
			c.Steam.Accounts[i].Enabled = enabled
			go c.scheduleSave()
			return
		}
	}
}

// UpdateSteamAccountStatus updates the last update timestamp and error message for a Steam account
func (c *Config) UpdateSteamAccountStatus(steamId64 string, timestampMillis int64, errorMessage string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.Steam.Accounts {
		if c.Steam.Accounts[i].SteamID64 == steamId64 {
			c.Steam.Accounts[i].LastUpdateTimestampMillis = timestampMillis
			c.Steam.Accounts[i].LastUpdateErrorMessage = errorMessage
			go c.scheduleSave()
			return
		}
	}
}

// GetSteamPath returns the configured Steam installation path
func (c *Config) GetSteamPath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Steam.SteamPath
}

// --- Migration Logic ---

func (c *Config) migrateToSteamAccounts() {
	c.Steam.AutoEnableNewAccounts = true
	c.Steam.Accounts = []SteamAccountConfig{}

	var remainingFiles []FileConfig
	for _, f := range c.HeroesLayout.Files {
		steamId3 := extractSteamId3FromPath(f.FilePath)
		if steamId3 != "" {
			steamId3Num, err := strconv.ParseUint(steamId3, 10, 64)
			if err == nil {
				steamId64 := strconv.FormatUint(steamid.ID3toID64(steamId3Num), 10)
				c.Steam.Accounts = append(c.Steam.Accounts, SteamAccountConfig{
					SteamID64:                 steamId64,
					Enabled:                   f.Enabled,
					LastUpdateTimestampMillis: f.LastUpdateTimestampMillis,
					LastUpdateErrorMessage:    f.LastUpdateErrorMessage,
				})
				continue
			}
		}
		remainingFiles = append(remainingFiles, FileConfig{
			FilePath:                  f.FilePath,
			Enabled:                   f.Enabled,
			LastUpdateTimestampMillis: f.LastUpdateTimestampMillis,
			LastUpdateErrorMessage:    f.LastUpdateErrorMessage,
		})
	}
	if remainingFiles == nil {
		remainingFiles = []FileConfig{}
	}
	c.HeroesLayout.Files = remainingFiles
}

func extractSteamId3FromPath(filePath string) string {
	normalized := strings.ReplaceAll(filePath, "\\", "/")
	matches := heroGridConfigPathRegex.FindStringSubmatch(normalized)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}
