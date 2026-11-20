package providers

type HeroesProvider interface {
	FetchPositions(positions []string, period string) (map[string]PositionStats, error)
}

type PositionStats struct {
	Heroes           []Hero
	minRatingMatches int
}

// Hero is one row of the Dota2ProTracker meta table for the selected position.
type Hero struct {
	HeroID       int      `json:"hero_id"`
	Matches      int      `json:"matches"`
	Wins         int      `json:"wins"`
	HeroName     string   `json:"hero_name"`
	MetaScore    *float64 `json:"meta_score"`
	D2PTRatingV2 *float64 `json:"d2pt_rating_v2"`
}
