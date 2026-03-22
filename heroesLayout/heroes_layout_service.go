package heroesLayout

import (
	"d2tool/config"
	"d2tool/providers"
	"d2tool/steam"
	"d2tool/utils"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type HeroesLayoutService interface {
	UpdateHeroesLayout() error
}

type HeroesLayoutServiceImpl struct {
	mu           sync.Mutex
	config       *config.Config
	steamService *steam.SteamService
	httpClient   *http.Client
}

func NewHeroesLayoutService(config *config.Config, steamService *steam.SteamService) *HeroesLayoutServiceImpl {
	return &HeroesLayoutServiceImpl{
		config:       config,
		steamService: steamService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

	d2ptConfig := s.config.GetD2PTConfig()
	period := d2ptConfig.Period
	heroesPerRow := s.config.GetHeroesPerRow()

	positionToAggregatedHeroes := make(map[string][]providers.Hero)
	positionToFacetedHeroes := make(map[string][]providers.Hero)

	var positionsFetchErr error
	for _, position := range positions {
		heroes, err := providers.FetchHeroes(position, period, s.httpClient, "")
		if err != nil {
			slog.Error("Error fetching heroes for position", "position", position, "error", err)
			positionsFetchErr = fmt.Errorf("error fetching heroes for position %s: %w", position, err)
			break
		}
		positionToAggregatedHeroes[position] = providers.AggregateHeroesByID(heroes)
		positionToFacetedHeroes[position] = providers.NormalizeFacetNumbers(heroes)
	}

	now := time.Now()

	if positionsFetchErr != nil {
		for steamId64 := range steamAccountPaths {
			s.steamService.UpdateAccountStatus(steamId64, now.UnixMilli(), positionsFetchErr.Error())
		}
		s.config.UpdateHeroesLayoutFileStatus(enabledFilePaths, now.UnixMilli(), positionsFetchErr.Error())
		return positionsFetchErr
	}

	for _, configFile := range allPaths {
		slog.Info("Processing config file", "path", configFile)

		errorMsg := ""
		if err := processHeroesLayoutConfig(configFile, positions, positionToAggregatedHeroes, positionToFacetedHeroes, heroesPerRow); err != nil {
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
