package steam

import (
	"cmp"
	"d2tool/config"
	"log/slog"
	"os"
	"slices"
	"sync"
)

type SteamAccountView struct {
	SteamID64                 string `json:"steamId64"`
	SteamID3                  string `json:"steamId3"`
	AccountName               string `json:"accountName"`
	PersonaName               string `json:"personaName"`
	AvatarBase64              string `json:"avatarBase64"`
	Enabled                   bool   `json:"enabled"`
	LastUpdateTimestampMillis int64  `json:"lastUpdateTimestampMillis"`
	LastUpdateErrorMessage    string `json:"lastUpdateErrorMessage"`
}

type SteamService struct {
	mu     sync.RWMutex
	config *config.Config
	cache  []SteamAccountView
}

func NewSteamService(cfg *config.Config) *SteamService {
	return &SteamService{
		config: cfg,
		cache:  []SteamAccountView{},
	}
}

// Init initializes the service: auto-discovers Steam path if needed, runs initial scan
func (s *SteamService) Init() {
	steamPath := s.config.GetSteamPath()
	if steamPath == "" {
		discovered, err := FindSteamPath()
		if err != nil {
			slog.Warn("Could not auto-discover Steam path", "error", err)
		} else {
			slog.Info("Auto-discovered Steam path", "path", discovered)
			s.config.SetSteamPath(discovered)
		}
	}

	if err := s.Scan(); err != nil {
		slog.Warn("Error scanning Steam accounts during init", "error", err)
	}
}

// Scan discovers accounts, syncs with config, rebuilds cache.
func (s *SteamService) Scan() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	steamPath := s.config.GetSteamPath()
	if steamPath == "" {
		slog.Warn("Steam path not set, skipping scan")
		return nil
	}

	discovered, err := ScanAccounts(steamPath)
	if err != nil {
		slog.Warn("Steam account scan failed, keeping existing state", "error", err)
		return err
	}

	discoveredMap := make(map[string]DiscoveredAccount)
	for _, d := range discovered {
		discoveredMap[d.SteamID64] = d
	}

	existingAccounts := s.config.GetSteamAccounts()
	existingMap := make(map[string]config.SteamAccountConfig)
	for _, a := range existingAccounts {
		existingMap[a.SteamID64] = a
	}

	autoEnable := s.config.GetSteamConfig().AutoEnableNewAccounts

	var updatedAccounts []config.SteamAccountConfig
	for steamId64 := range discoveredMap {
		if existing, ok := existingMap[steamId64]; ok {
			updatedAccounts = append(updatedAccounts, existing)
		} else {
			d := discoveredMap[steamId64]
			slog.Info("Discovered new Steam account", "steamId64", steamId64, "personaName", d.PersonaName)
			updatedAccounts = append(updatedAccounts, config.SteamAccountConfig{
				SteamID64: steamId64,
				Enabled:   autoEnable,
			})
		}
	}

	s.config.SetSteamAccounts(updatedAccounts)
	s.cache = s.buildCache(updatedAccounts, discoveredMap)

	return nil
}

// GetAccounts returns the cached account views
func (s *SteamService) GetAccounts() []SteamAccountView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]SteamAccountView, len(s.cache))
	copy(result, s.cache)
	return result
}

// SetAccountEnabled updates the enabled state in both config and cache
func (s *SteamService) SetAccountEnabled(steamId64 string, enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.SetSteamAccountEnabled(steamId64, enabled)
	for i := range s.cache {
		if s.cache[i].SteamID64 == steamId64 {
			s.cache[i].Enabled = enabled
			break
		}
	}
}

// UpdateAccountStatus updates the status in both config and cache
func (s *SteamService) UpdateAccountStatus(steamId64 string, timestampMillis int64, errorMessage string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.UpdateSteamAccountStatus(steamId64, timestampMillis, errorMessage)
	for i := range s.cache {
		if s.cache[i].SteamID64 == steamId64 {
			s.cache[i].LastUpdateTimestampMillis = timestampMillis
			s.cache[i].LastUpdateErrorMessage = errorMessage
			break
		}
	}
}

func (s *SteamService) IsPathValid() bool {
	steamPath := s.config.GetSteamPath()
	if steamPath == "" {
		return false
	}
	info, err := os.Stat(steamPath)
	return err == nil && info.IsDir()
}

// GetEnabledAccountPaths returns hero_grid_config.json paths for enabled accounts
func (s *SteamService) GetEnabledAccountPaths() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	steamPath := s.config.GetSteamPath()
	if steamPath == "" {
		return nil
	}

	paths := make(map[string]string)
	for _, acc := range s.cache {
		if acc.Enabled {
			paths[acc.SteamID64] = HeroGridConfigPath(steamPath, acc.SteamID3)
		}
	}
	return paths
}

func (s *SteamService) buildCache(accounts []config.SteamAccountConfig, discovered map[string]DiscoveredAccount) []SteamAccountView {
	views := make([]SteamAccountView, 0, len(accounts))
	for _, acc := range accounts {
		view := SteamAccountView{
			SteamID64:                 acc.SteamID64,
			Enabled:                   acc.Enabled,
			LastUpdateTimestampMillis: acc.LastUpdateTimestampMillis,
			LastUpdateErrorMessage:    acc.LastUpdateErrorMessage,
		}
		if d, ok := discovered[acc.SteamID64]; ok {
			view.SteamID3 = d.SteamID3
			view.AccountName = d.AccountName
			view.PersonaName = d.PersonaName
			view.AvatarBase64 = d.AvatarBase64
		}
		views = append(views, view)
	}
	slices.SortFunc(views, func(a, b SteamAccountView) int {
		return cmp.Compare(a.AccountName, b.AccountName)
	})

	return views
}
