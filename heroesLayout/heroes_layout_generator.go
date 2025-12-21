package heroesLayout

import (
	"d2tool/providers"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	positionPrefix = "pos "
	d2tPrefix      = "[D2T]"
	facetsPrefix   = "[Facets]"
)

type UpdateHeroesLayoutConfig struct {
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
	CategoryName string  `json:"category_name"`
	XPosition    float64 `json:"x_position"`
	YPosition    float64 `json:"y_position"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	HeroIDs      []int   `json:"hero_ids"`
}

// generateHeroesLayoutConfigs generates new hero grid configs for each role
func generateHeroesLayoutConfigs(configNamePrefix string, positions []string, positionToHero map[string][]providers.Hero) []heroGridCategory {
	var configs []heroGridCategory

	// Create a single merged config
	mergedConfig := heroGridCategory{
		ConfigName: fmt.Sprintf("%s Heroes Meta %s", configNamePrefix, time.Now().Format("2006-01-02")),
		Categories: []heroGridPosition{},
	}

	// Set vertical layout with fixed width
	const heroWidth = 70
	const heroHeight = 110
	const categorySpacing = 50
	const infoHeight = 30
	const heroesPerRow = 15
	const heroWinrateSpacing = 20
	const heroFacetSpacing = 20
	const heroRowSpacing = 30
	const headerWidth = 100 // Width for row headers

	// Current Y position for vertical layout
	currentY := 0

	// Check if any hero has facet info (used to determine row height)
	hasFacetInfo := func(heroes []providers.Hero) bool {
		for _, hero := range heroes {
			if hero.FacetNumber > 0 {
				return true
			}
		}
		return false
	}

	generateCategoryFunc := func(categoryName string, heroes []providers.Hero) {
		infoCategory := heroGridPosition{
			CategoryName: categoryName,
			XPosition:    0,
			YPosition:    float64(currentY),
			Width:        0,
			Height:       0,
			HeroIDs:      []int{},
		}
		mergedConfig.Categories = append(mergedConfig.Categories, infoCategory)
		currentY += infoHeight

		// Determine if we need extra spacing for facet labels
		showFacets := hasFacetInfo(heroes)
		facetOffset := 0
		if showFacets {
			facetOffset = heroFacetSpacing
		}

		// Add each hero as a separate category with winrate and match count in the category name
		for i, hero := range heroes {
			// Calculate position in the grid
			row := i / heroesPerRow
			col := i % heroesPerRow

			// Calculate x and y position (offset by headerWidth to make room for row headers)
			xPos := headerWidth + col*heroWidth
			yPos := currentY + row*(heroHeight+heroWinrateSpacing+facetOffset+heroRowSpacing)

			labelYPos := yPos

			// Add row headers at the start of each row
			if col == 0 {
				if showFacets {
					mergedConfig.Categories = append(mergedConfig.Categories, heroGridPosition{
						CategoryName: "Facet",
						XPosition:    0,
						YPosition:    float64(labelYPos),
						Width:        0,
						Height:       0,
						HeroIDs:      []int{},
					})
				}
				mergedConfig.Categories = append(mergedConfig.Categories, heroGridPosition{
					CategoryName: "Winrate",
					XPosition:    0,
					YPosition:    float64(labelYPos + facetOffset),
					Width:        0,
					Height:       0,
					HeroIDs:      []int{},
				})
				mergedConfig.Categories = append(mergedConfig.Categories, heroGridPosition{
					CategoryName: "Matches",
					XPosition:    0,
					YPosition:    float64(labelYPos + facetOffset + heroWinrateSpacing),
					Width:        0,
					Height:       0,
					HeroIDs:      []int{},
				})
			}

			// Add facet label if present (above winrate)
			if hero.FacetNumber > 0 {
				heroFacetCategory := heroGridPosition{
					CategoryName: fmt.Sprintf("  %d", hero.FacetNumber),
					XPosition:    float64(xPos),
					YPosition:    float64(labelYPos),
					Width:        0,
					Height:       0,
					HeroIDs:      []int{},
				}
				mergedConfig.Categories = append(mergedConfig.Categories, heroFacetCategory)
				labelYPos += heroFacetSpacing
			}

			// Calculate winrate
			winrate := 0.0
			if hero.Matches > 0 {
				winrate = float64(hero.Wins) / float64(hero.Matches) * 100
			}

			heroWinrateCategory := heroGridPosition{
				CategoryName: fmt.Sprintf("  %.1f%%", winrate),
				XPosition:    float64(xPos),
				YPosition:    float64(labelYPos),
				Width:        0,
				Height:       0,
				HeroIDs:      []int{},
			}

			mergedConfig.Categories = append(mergedConfig.Categories, heroWinrateCategory)

			// Create category for the hero
			heroCategory := heroGridPosition{
				CategoryName: fmt.Sprintf("  %d", hero.Matches),
				XPosition:    float64(xPos),
				YPosition:    float64(labelYPos + heroWinrateSpacing),
				Width:        heroWidth,
				Height:       heroHeight,
				HeroIDs:      []int{hero.HeroID},
			}
			mergedConfig.Categories = append(mergedConfig.Categories, heroCategory)
		}

		// Update currentY to account for all rows of heroes
		rows := (len(heroes) + heroesPerRow - 1) / heroesPerRow // Ceiling division
		currentY += rows*(heroHeight+heroWinrateSpacing+facetOffset+heroRowSpacing) + categorySpacing
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

// processHeroesLayoutConfig processes a hero_grid_config.json file
func processHeroesLayoutConfig(configPath string, positions []string, positionToAggregatedHeroes map[string][]providers.Hero, positionToFacetedHeroes map[string][]providers.Hero) error {
	// Read the existing config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	config := heroGridConfig{
		Version: 3,
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// Filter out configs with [D2T] prefix (removes both aggregated and faceted)
	var filteredConfigs []heroGridCategory
	for _, cfg := range config.Configs {
		if !strings.HasPrefix(cfg.ConfigName, d2tPrefix) {
			filteredConfigs = append(filteredConfigs, cfg)
		}
	}

	config.Configs = filteredConfigs

	// Generate aggregated config (no facets)
	aggregatedConfigs := generateHeroesLayoutConfigs(d2tPrefix, positions, positionToAggregatedHeroes)

	// Generate faceted config (split by facets)
	facetedConfigs := generateHeroesLayoutConfigs(fmt.Sprintf("%s %s", d2tPrefix, facetsPrefix), positions, positionToFacetedHeroes)

	// Merge all configs
	config.Configs = append(config.Configs, aggregatedConfigs...)
	config.Configs = append(config.Configs, facetedConfigs...)

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
