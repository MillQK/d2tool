package providers

import (
	"cmp"
	"slices"
)

// The site's Tier & Rating column sorts on the unrounded score, with unrated rows last.
func (h Hero) rating() float64 {
	score := cmp.Or(h.D2PTRatingV2, h.MetaScore)
	if score == nil || *score < 0 || *score > 100 {
		return -1
	}
	return *score
}

func (stats PositionStats) TopRated(n int) []Hero {
	rows := make([]Hero, 0, len(stats.Heroes))
	for _, hero := range stats.Heroes {
		if hero.Matches >= stats.minRatingMatches {
			rows = append(rows, hero)
		}
	}
	slices.SortStableFunc(rows, func(a, b Hero) int {
		return cmp.Compare(b.rating(), a.rating())
	})
	return rows[:min(n, len(rows))]
}

func (stats PositionStats) MostPlayed(n int) []Hero {
	// The grid counts facets with at least 20 matches toward each hero's total.
	var heroes []Hero
	indexByID := make(map[int]int)
	for _, row := range stats.Heroes {
		if row.Matches < 20 {
			continue
		}
		if i, ok := indexByID[row.HeroID]; ok {
			heroes[i].Matches += row.Matches
			heroes[i].Wins += row.Wins
		} else {
			indexByID[row.HeroID] = len(heroes)
			heroes = append(heroes, row)
		}
	}
	slices.SortStableFunc(heroes, func(a, b Hero) int {
		return cmp.Compare(b.Matches, a.Matches)
	})
	return heroes[:min(n, len(heroes))]
}
