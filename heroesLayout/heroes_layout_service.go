package heroesLayout

import (
	"d2tool/config"
	"d2tool/providers"
	"d2tool/steam"
	"d2tool/utils"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type HeroesLayoutService interface {
	UpdateHeroesLayout() error
}

type HeroesLayoutServiceImpl struct {
	mu             sync.Mutex
	config         *config.Config
	steamService   *steam.SteamService
	heroesProvider providers.HeroesProvider
}

func NewHeroesLayoutService(config *config.Config, steamService *steam.SteamService, heroesProvider providers.HeroesProvider) *HeroesLayoutServiceImpl {
	return &HeroesLayoutServiceImpl{
		config:         config,
		steamService:   steamService,
		heroesProvider: heroesProvider,
	}
}

func (s *HeroesLayoutServiceImpl) UpdateHeroesLayout() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	steamAccountPaths := s.steamService.GetEnabledAccountPaths() // map[steamId64]path
	enabledFilePaths := s.config.GetEnabledFilePaths()
	enabledPositions := s.config.GetEnabledPositionIDs()

	pathToSteamId64 := make(map[string]string, len(steamAccountPaths))
	var allPaths []string
	for steamId64, path := range steamAccountPaths {
		pathToSteamId64[path] = steamId64
		allPaths = append(allPaths, path)
	}
	allPaths = append(allPaths, enabledFilePaths...)

	if len(allPaths) == 0 {
		slog.Info("No config files provided, skipping update")
		return nil
	}

	if len(enabledPositions) == 0 {
		slog.Info("No hero positions enabled, skipping update")
		return nil
	}

	positions := utils.Map(enabledPositions, func(position string) string {
		return fmt.Sprintf("%s%s", positionPrefix, position)
	})

	period := s.config.GetD2PTConfig().Period
	heroesPerRow := s.config.GetHeroesPerRow()

	positionStats, fetchErr := s.heroesProvider.FetchPositions(positions, period)
	now := time.Now()
	if fetchErr != nil {
		slog.Error("Error fetching hero statistics", "error", fetchErr)
		for steamId64 := range steamAccountPaths {
			s.steamService.UpdateAccountStatus(steamId64, now.UnixMilli(), fetchErr.Error())
		}
		s.config.UpdateHeroesLayoutFileStatus(enabledFilePaths, now.UnixMilli(), fetchErr.Error())
		return fetchErr
	}

	for _, configFile := range allPaths {
		slog.Info("Processing config file", "path", configFile)

		errorMsg := ""
		if err := processHeroesLayoutConfig(configFile, positions, positionStats, heroesPerRow); err != nil {
			slog.Error("Error processing config file", "path", configFile, "error", err)
			errorMsg = fmt.Sprintf("error processing config file: %v", err)
		} else {
			slog.Info("Successfully updated config file", "path", configFile)
		}

		if steamId64, ok := pathToSteamId64[configFile]; ok {
			s.steamService.UpdateAccountStatus(steamId64, now.UnixMilli(), errorMsg)
		} else {
			s.config.UpdateHeroesLayoutFileStatus([]string{configFile}, now.UnixMilli(), errorMsg)
		}
	}

	return nil
}
