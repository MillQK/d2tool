package providers

type HeroesProvider interface {
	FetchHeroes(position string, period string) ([]Hero, error)
}

// Hero represents a Dota 2 hero with its statistics
type Hero struct {
	HeroID     int    `json:"hero_id"`
	Matches    int    `json:"matches"`
	Wins       int    `json:"wins"`
	HeroName   string `json:"hero_name"`
	D2PTRating int    `json:"d2pt_rating"`
}
