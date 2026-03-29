package providers

import "sort"

// AggregateHeroesByID merges heroes with the same hero_id by summing wins and matches.
func AggregateHeroesByID(heroes []Hero) []Hero {
	heroIdToAllInstances := make(map[int][]Hero)
	aggregatedHeroMap := make(map[int]Hero)

	for _, hero := range heroes {
		heroIdToAllInstances[hero.HeroID] = append(heroIdToAllInstances[hero.HeroID], hero)
	}

	for heroId, mapHeroes := range heroIdToAllInstances {
		aggregatedHero := Hero{
			HeroID:     heroId,
			Matches:    0,
			Wins:       0,
			HeroName:   mapHeroes[0].HeroName,
			D2PTRating: 0,
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
