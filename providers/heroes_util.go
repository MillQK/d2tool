package providers

import "sort"

// NormalizeFacetNumbers normalizes facet numbers per hero_id to sequential 1, 2, 3, etc.
// The API may return non-sequential numbers (e.g., 2, 3 instead of 1, 2), so we normalize them.
func NormalizeFacetNumbers(heroes []Hero) []Hero {
	// Group heroes by hero_id and collect their facet numbers
	heroFacets := make(map[int][]int)
	for _, hero := range heroes {
		heroFacets[hero.HeroID] = append(heroFacets[hero.HeroID], hero.FacetNumber)
	}

	// For each hero_id, sort facet numbers and create mapping to normalized values
	facetMapping := make(map[int]map[int]int) // hero_id -> original_facet -> normalized_facet
	for heroID, facets := range heroFacets {
		sort.Ints(facets)

		// Create mapping: original -> normalized (1-based)
		facetMapping[heroID] = make(map[int]int)
		for i, f := range facets {
			facetMapping[heroID][f] = i + 1
		}
	}

	// Apply normalization to all heroes
	result := make([]Hero, len(heroes))
	for i, hero := range heroes {
		result[i] = hero
		if mapping, ok := facetMapping[hero.HeroID]; ok {
			if normalized, ok := mapping[hero.FacetNumber]; ok {
				result[i].FacetNumber = normalized
			}
		}
	}

	return result
}

// AggregateHeroesByID merges heroes with the same hero_id by summing wins and matches.
// This is used when facet grouping is disabled to combine stats across all facets.
func AggregateHeroesByID(heroes []Hero) []Hero {
	heroIdToAllInstances := make(map[int][]Hero)
	aggregatedHeroMap := make(map[int]Hero)

	for _, hero := range heroes {
		heroIdToAllInstances[hero.HeroID] = append(heroIdToAllInstances[hero.HeroID], hero)
	}

	for heroId, mapHeroes := range heroIdToAllInstances {
		aggregatedHero := Hero{
			HeroID:      heroId,
			Matches:     0,
			Wins:        0,
			HeroName:    mapHeroes[0].HeroName,
			D2PTRating:  0,
			FacetName:   "",
			FacetNumber: -1,
		}

		for _, hero := range mapHeroes {
			aggregatedHero.Matches += hero.Matches
			aggregatedHero.Wins += hero.Wins
		}

		if aggregatedHero.Matches == 0 {
			continue
		}

		for _, hero := range mapHeroes {
			weight := float64(hero.Matches) / float64(aggregatedHero.Matches)
			aggregatedHero.D2PTRating += int(float64(hero.D2PTRating) * weight)
		}

		aggregatedHeroMap[heroId] = aggregatedHero
	}

	result := make([]Hero, 0, len(aggregatedHeroMap))
	for _, hero := range aggregatedHeroMap {
		result = append(result, hero)
	}

	return result
}

// GetTopHeroesByRating returns the top N heroes by d2pt_rating
func GetTopHeroesByRating(heroes []Hero, n int) []Hero {
	// Create a copy to avoid modifying the original slice
	result := make([]Hero, len(heroes))
	copy(result, heroes)

	// Sort by d2pt_rating in descending order
	sort.Slice(result, func(i, j int) bool {
		return result[i].D2PTRating > result[j].D2PTRating
	})

	// Return top N heroes or all if less than N
	if len(result) > n {
		return result[:n]
	}
	return result
}

// GetHeroesSortedByMatches returns the top N heroes by matches
func GetHeroesSortedByMatches(heroes []Hero, n int) []Hero {
	// Create a copy to avoid modifying the original slice
	result := make([]Hero, len(heroes))
	copy(result, heroes)

	// Sort by matches in descending order
	sort.Slice(result, func(i, j int) bool {
		return result[i].Matches > result[j].Matches
	})

	// Return top N heroes or all if less than N
	if len(result) > n {
		return result[:n]
	}
	return result
}
