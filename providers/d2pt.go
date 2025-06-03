package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
)

// Hero represents a Dota 2 hero with its statistics
type Hero struct {
	HeroID      int     `json:"hero_id"`
	Position    string  `json:"position"`
	MMR         string  `json:"mmr"`
	Period      string  `json:"period"`
	ContestRate float64 `json:"contest_rate"`
	Matches     int     `json:"matches"`
	Wins        int     `json:"wins"`
	UpdatedAt   string  `json:"updated_at"`
	HeroName    string  `json:"hero_name"`
	NPC         string  `json:"npc"`
	D2PTRating  int     `json:"d2pt_rating"`
}

// FetchHeroes fetches heroes data from the API for a specific position
func FetchHeroes(position string) ([]Hero, error) {
	d2ptUrl := fmt.Sprintf("https://dota2protracker.com/api/heroes/stats?mmr=7000&order_by=matches&min_matches=20&period=8&position=%s", url.QueryEscape(position))

	// Create a new request
	req, err := http.NewRequest("GET", d2ptUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add required headers
	req.Header.Add("Accept", "application/json")

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var heroes []Hero
	if err := json.Unmarshal(body, &heroes); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return heroes, nil
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
