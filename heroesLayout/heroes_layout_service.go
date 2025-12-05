package heroesLayout

import (
	"d2tool/config"
	"d2tool/providers"
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
	mu         sync.Mutex
	config     *config.Config
	httpClient *http.Client
}

func NewHeroesLayoutService(config *config.Config) *HeroesLayoutServiceImpl {
	return &HeroesLayoutServiceImpl{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *HeroesLayoutServiceImpl) UpdateHeroesLayout() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get enabled files and positions
	enabledFilePaths := s.config.GetEnabledFilePaths()
	enabledPositions := s.config.GetEnabledPositionIDs()

	if len(enabledPositions) == 0 {
		slog.Info("No config files provided, skipping update")
		return nil
	}

	if len(enabledPositions) == 0 {
		slog.Info("No hero positions enabled, skipping update")
		return nil
	}

	// Fetch heroes data for all positions
	positions := utils.Map(
		enabledPositions,
		func(position string) string {
			return fmt.Sprintf("%s%s", positionPrefix, position)
		},
	)

	positionToHeroes := make(map[string][]providers.Hero)

	var positionsFetchErr error

	for _, position := range positions {
		heroes, err := providers.FetchHeroes(position, s.httpClient, "")
		if err != nil {
			slog.Error(fmt.Sprintf("Error fetching heroes for position %s", position), "error", err)
			positionsFetchErr = fmt.Errorf("error fetching heroes for position %s: %w", position, err)
			break
		}
		positionToHeroes[position] = heroes
	}

	now := time.Now()

	if positionsFetchErr != nil {
		s.config.UpdateHeroesLayoutFileStatus(enabledFilePaths, now.UnixMilli(), positionsFetchErr.Error())
		return positionsFetchErr
	}

	// Process each config file
	for _, configFile := range enabledFilePaths {
		slog.Info(fmt.Sprintf("Processing config file %s", configFile))

		errorMsg := ""

		if err := processHeroesLayoutConfig(configFile, positions, positionToHeroes); err != nil {
			slog.Error(fmt.Sprintf("Error processing config file %s", configFile), "error", err)
			errorMsg = fmt.Sprintf("error processing config file: %v", err)
		} else {
			slog.Info(fmt.Sprintf("Successfully updated config file %s", configFile))
		}

		s.config.UpdateHeroesLayoutFileStatus(
			[]string{configFile},
			now.UnixMilli(),
			errorMsg,
		)
	}

	return nil
}
