package heroesGrid

import (
	"d2tool/providers"
	"d2tool/utils"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	positionPrefix = "pos "
	d2tPrefix      = "[D2T]"
)

type UpdateHeroGridConfig struct {
	ConfigFilePaths []string
	Positions       []string
}

// heroGridConfig represents the structure of the hero_grid_config.json file
type heroGridConfig struct {
	Version int                `json:"version"`
	Configs []heroGridCategory `json:"configs"`
}

// heroGridCategory represents a category in the hero grid config
type heroGridCategory struct {
	ConfigName string             `json:"config_name"`
	Categories []heroGridPosition `json:"categories"`
}

// heroGridPosition represents a position category in the hero grid
type heroGridPosition struct {
	CategoryName string `json:"category_name"`
	XPosition    int    `json:"x_position"`
	YPosition    int    `json:"y_position"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	HeroIDs      []int  `json:"hero_ids"`
}

// UpdateHeroesGrid updates all hero grid config files with new hero data
func UpdateHeroesGrid(config UpdateHeroGridConfig) error {
	if len(config.ConfigFilePaths) == 0 {
		slog.Info("No config files provided, skipping update")
		return nil
	}

	// Fetch heroes data for all positions
	positions := utils.Map(
		config.Positions,
		func(position string) string {
			return fmt.Sprintf("%s%s", positionPrefix, position)
		},
	)

	positionToHeroes := make(map[string][]providers.Hero)

	for _, position := range positions {
		heroes, err := providers.FetchHeroes(position)
		if err != nil {
			slog.Error(fmt.Sprintf("Error fetching heroes for position %s", position), "error", err)
			continue
		}
		positionToHeroes[position] = heroes
	}

	// Process each config file
	for _, configFile := range config.ConfigFilePaths {
		slog.Info(fmt.Sprintf("Processing config file %s", configFile))
		if err := processHeroGridConfig(configFile, positions, positionToHeroes); err != nil {
			slog.Error(fmt.Sprintf("Error processing config file %s", configFile), "error", err)
			continue
		}
		slog.Info(fmt.Sprintf("Successfully updated config file %s", configFile))
	}

	return nil
}

// generateHeroGridConfigs generates new hero grid configs for each role
func generateHeroGridConfigs(positions []string, positionToHero map[string][]providers.Hero) []heroGridCategory {
	var configs []heroGridCategory

	// Create a single merged config
	mergedConfig := heroGridCategory{
		ConfigName: fmt.Sprintf("%s Heroes Meta %s", d2tPrefix, time.Now().Format("2006-01-02")),
		Categories: []heroGridPosition{},
	}

	// Set vertical layout with fixed width
	const heroWidth = 70
	const heroHeight = 110
	const categorySpacing = 50
	const infoHeight = 30
	const heroesPerRow = 15
	const heroWinrateSpacing = 20
	const heroRowSpacing = 30

	// Current Y position for vertical layout
	currentY := 0

	generateCategoryFunc := func(categoryName string, heroes []providers.Hero) {
		infoCategory := heroGridPosition{
			CategoryName: categoryName,
			XPosition:    0,
			YPosition:    currentY,
			Width:        0,
			Height:       0,
			HeroIDs:      []int{},
		}
		mergedConfig.Categories = append(mergedConfig.Categories, infoCategory)
		currentY += infoHeight

		// Add each hero as a separate category with winrate and match count in the category name
		for i, hero := range heroes {
			// Calculate position in the grid
			row := i / heroesPerRow
			col := i % heroesPerRow

			// Calculate x and y position
			xPos := col * heroWidth
			yPos := currentY + row*(heroHeight+heroWinrateSpacing+heroRowSpacing)

			// Calculate winrate
			winrate := float64(hero.Wins) / float64(hero.Matches) * 100

			heroWinrateCategory := heroGridPosition{
				CategoryName: fmt.Sprintf("  %.1f%%", winrate),
				XPosition:    xPos,
				YPosition:    yPos,
				Width:        0,
				Height:       0,
				HeroIDs:      []int{},
			}

			mergedConfig.Categories = append(mergedConfig.Categories, heroWinrateCategory)

			// Create category for the hero
			heroCategory := heroGridPosition{
				CategoryName: fmt.Sprintf("  %d", hero.Matches),
				XPosition:    xPos,
				YPosition:    yPos + heroWinrateSpacing,
				Width:        heroWidth,
				Height:       heroHeight,
				HeroIDs:      []int{hero.HeroID},
			}
			mergedConfig.Categories = append(mergedConfig.Categories, heroCategory)
		}

		// Update currentY to account for all rows of heroes
		rows := (len(heroes) + heroesPerRow - 1) / heroesPerRow // Ceiling division
		currentY += rows*(heroHeight+heroWinrateSpacing+heroRowSpacing) + categorySpacing
	}

	// Generate categories for each position in the specified order
	for _, position := range positions {
		heroes := positionToHero[position]

		// Get top 10 heroes by rating
		topRated := providers.GetTopHeroesByRating(heroes, 10)

		// Add list header for top rated heroes
		generateCategoryFunc(
			fmt.Sprintf("%s - Top Rating Heroes", position),
			topRated,
		)

		// Get top 30 heroes by matches
		topMatches := providers.GetHeroesSortedByMatches(heroes, 30)

		// Add list header for top matches heroes
		generateCategoryFunc(
			fmt.Sprintf("%s - Most Matches Heroes", position),
			topMatches,
		)
	}

	configs = append(configs, mergedConfig)
	return configs
}

// FindHeroGridConfigFiles finds all hero_grid_config.json files for all Steam users
func FindHeroGridConfigFiles(steamPath string) ([]string, error) {
	userdataPath := filepath.Join(steamPath, "userdata")

	// Check if userdata directory exists
	if _, err := os.Stat(userdataPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("userdata directory not found at %s", userdataPath)
	}

	var configFiles []string

	// List all directories in userdata (each is a Steam user ID)
	userDirs, err := os.ReadDir(userdataPath)
	if err != nil {
		return nil, fmt.Errorf("error reading userdata directory: %w", err)
	}

	for _, userDir := range userDirs {
		if !userDir.IsDir() {
			continue
		}

		// Construct path to hero_grid_config.json
		configPath := filepath.Join(userdataPath, userDir.Name(), "570", "remote", "cfg", "hero_grid_config.json")

		// Check if file exists
		if _, err := os.Stat(configPath); err == nil {
			configFiles = append(configFiles, configPath)
		}
	}

	if len(configFiles) == 0 {
		return nil, fmt.Errorf("no hero_grid_config.json files found")
	}

	return configFiles, nil
}

// processHeroGridConfig processes a hero_grid_config.json file
func processHeroGridConfig(configPath string, positions []string, positionToHeroes map[string][]providers.Hero) error {
	// Read the existing config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var config heroGridConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// Filter out configs with {d2tPrefix} prefix
	var filteredConfigs []heroGridCategory
	for _, cfg := range config.Configs {
		if !strings.HasPrefix(cfg.ConfigName, d2tPrefix) {
			filteredConfigs = append(filteredConfigs, cfg)
		}
	}

	// Generate new configs for each role
	newConfigs := generateHeroGridConfigs(positions, positionToHeroes)

	// Merge configs
	config.Configs = append(filteredConfigs, newConfigs...)

	// Write the updated config back to file
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling updated config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("error writing updated config: %w", err)
	}

	return nil
}
