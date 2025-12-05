package providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTopHeroesByRating(t *testing.T) {
	heroes := []Hero{
		{HeroID: 1, D2PTRating: 100, Matches: 50},
		{HeroID: 2, D2PTRating: 200, Matches: 30},
		{HeroID: 3, D2PTRating: 150, Matches: 40},
		{HeroID: 4, D2PTRating: 180, Matches: 20},
	}

	top2 := GetTopHeroesByRating(heroes, 2)

	if len(top2) != 2 {
		t.Fatalf("expected 2 heroes, got %d", len(top2))
	}
	if top2[0].HeroID != 2 {
		t.Errorf("expected hero ID 2 first (rating 200), got %d", top2[0].HeroID)
	}
	if top2[1].HeroID != 4 {
		t.Errorf("expected hero ID 4 second (rating 180), got %d", top2[1].HeroID)
	}
}

func TestGetTopHeroesByRating_LessThanN(t *testing.T) {
	heroes := []Hero{
		{HeroID: 1, D2PTRating: 100},
	}

	top5 := GetTopHeroesByRating(heroes, 5)

	if len(top5) != 1 {
		t.Errorf("expected 1 hero when requesting more than available, got %d", len(top5))
	}
}

func TestGetTopHeroesByRating_EmptySlice(t *testing.T) {
	heroes := []Hero{}

	top5 := GetTopHeroesByRating(heroes, 5)

	if len(top5) != 0 {
		t.Errorf("expected 0 heroes for empty input, got %d", len(top5))
	}
}

func TestGetTopHeroesByRating_DoesNotModifyOriginal(t *testing.T) {
	heroes := []Hero{
		{HeroID: 1, D2PTRating: 100},
		{HeroID: 2, D2PTRating: 200},
		{HeroID: 3, D2PTRating: 150},
	}

	// Store original order
	originalFirstID := heroes[0].HeroID

	_ = GetTopHeroesByRating(heroes, 2)

	// Verify original slice is unchanged
	if heroes[0].HeroID != originalFirstID {
		t.Error("GetTopHeroesByRating modified the original slice")
	}
}

func TestGetHeroesSortedByMatches(t *testing.T) {
	heroes := []Hero{
		{HeroID: 1, Matches: 10},
		{HeroID: 2, Matches: 50},
		{HeroID: 3, Matches: 30},
	}

	sorted := GetHeroesSortedByMatches(heroes, 3)

	if sorted[0].HeroID != 2 {
		t.Errorf("expected hero with most matches (ID 2) first, got ID %d", sorted[0].HeroID)
	}
	if sorted[1].HeroID != 3 {
		t.Errorf("expected hero with second most matches (ID 3) second, got ID %d", sorted[1].HeroID)
	}
	if sorted[2].HeroID != 1 {
		t.Errorf("expected hero with least matches (ID 1) last, got ID %d", sorted[2].HeroID)
	}
}

func TestGetHeroesSortedByMatches_TopN(t *testing.T) {
	heroes := []Hero{
		{HeroID: 1, Matches: 10},
		{HeroID: 2, Matches: 50},
		{HeroID: 3, Matches: 30},
		{HeroID: 4, Matches: 40},
	}

	top2 := GetHeroesSortedByMatches(heroes, 2)

	if len(top2) != 2 {
		t.Fatalf("expected 2 heroes, got %d", len(top2))
	}
	if top2[0].HeroID != 2 {
		t.Errorf("expected hero ID 2 first (50 matches), got %d", top2[0].HeroID)
	}
	if top2[1].HeroID != 4 {
		t.Errorf("expected hero ID 4 second (40 matches), got %d", top2[1].HeroID)
	}
}

func TestGetHeroesSortedByMatches_DoesNotModifyOriginal(t *testing.T) {
	heroes := []Hero{
		{HeroID: 1, Matches: 100},
		{HeroID: 2, Matches: 200},
		{HeroID: 3, Matches: 150},
	}

	originalFirstID := heroes[0].HeroID

	_ = GetHeroesSortedByMatches(heroes, 2)

	if heroes[0].HeroID != originalFirstID {
		t.Error("GetHeroesSortedByMatches modified the original slice")
	}
}

func TestFetchHeroes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		if r.URL.Path != "/heroes/stats" {
			t.Errorf("expected path /heroes/stats, got %s", r.URL.Path)
		}

		// Verify request parameters
		if r.URL.Query().Get("position") != "1" {
			t.Errorf("expected position=1, got %s", r.URL.Query().Get("position"))
		}
		if r.URL.Query().Get("mmr") != "7000" {
			t.Errorf("expected mmr=7000, got %s", r.URL.Query().Get("mmr"))
		}

		// Verify headers
		if r.Header.Get("Accept") != "application/json" {
			t.Error("expected Accept: application/json header")
		}
		if r.Header.Get("Referer") != "https://dota2protracker.com" {
			t.Error("expected Referer header")
		}

		heroes := []Hero{
			{HeroID: 1, HeroName: "Anti-Mage", D2PTRating: 100, Matches: 500},
			{HeroID: 2, HeroName: "Axe", D2PTRating: 150, Matches: 300},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(heroes)
	}))
	defer server.Close()

	heroes, err := FetchHeroes("1", http.DefaultClient, server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(heroes) != 2 {
		t.Fatalf("expected 2 heroes, got %d", len(heroes))
	}
	if heroes[0].HeroID != 1 {
		t.Errorf("expected first hero ID 1, got %d", heroes[0].HeroID)
	}
	if heroes[0].HeroName != "Anti-Mage" {
		t.Errorf("expected first hero name 'Anti-Mage', got %s", heroes[0].HeroName)
	}
	if heroes[1].D2PTRating != 150 {
		t.Errorf("expected second hero rating 150, got %d", heroes[1].D2PTRating)
	}
}

func TestFetchHeroes_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := FetchHeroes("1", http.DefaultClient, server.URL)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestFetchHeroes_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := FetchHeroes("1", http.DefaultClient, server.URL)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestFetchHeroes_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	_, err := FetchHeroes("1", http.DefaultClient, server.URL)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFetchHeroes_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	heroes, err := FetchHeroes("1", http.DefaultClient, server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(heroes) != 0 {
		t.Errorf("expected 0 heroes, got %d", len(heroes))
	}
}

// Benchmark tests
func BenchmarkGetTopHeroesByRating(b *testing.B) {
	// Create a realistic dataset
	heroes := make([]Hero, 100)
	for i := range heroes {
		heroes[i] = Hero{
			HeroID:     i + 1,
			D2PTRating: i * 10,
			Matches:    i * 5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetTopHeroesByRating(heroes, 10)
	}
}

func BenchmarkGetHeroesSortedByMatches(b *testing.B) {
	heroes := make([]Hero, 100)
	for i := range heroes {
		heroes[i] = Hero{
			HeroID:     i + 1,
			D2PTRating: i * 10,
			Matches:    i * 5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetHeroesSortedByMatches(heroes, 30)
	}
}
