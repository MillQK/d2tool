package providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// --- Fetch tests (migrated from d2pt_test.go) ---

func TestD2PTHeroesProvider_FetchHeroes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/heroes/stats" {
			t.Errorf("expected path /heroes/stats, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("position") != "1" {
			t.Errorf("expected position=1, got %s", r.URL.Query().Get("position"))
		}
		if r.URL.Query().Get("mmr") != "7000" {
			t.Errorf("expected mmr=7000, got %s", r.URL.Query().Get("mmr"))
		}

		if r.Header.Get("Accept") != "application/json" {
			t.Error("expected Accept: application/json header")
		}
		if r.Header.Get("Referer") != "https://dota2protracker.com" {
			t.Error("expected Referer header")
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("expected User-Agent header")
		}

		heroes := []Hero{
			{HeroID: 1, HeroName: "Anti-Mage", D2PTRating: 100, Matches: 500},
			{HeroID: 2, HeroName: "Axe", D2PTRating: 150, Matches: 300},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(heroes)
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	heroes, err := provider.FetchHeroes("1", "8")
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

func TestD2PTHeroesProvider_FetchHeroes_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	_, err := provider.FetchHeroes("1", "8")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestD2PTHeroesProvider_FetchHeroes_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	_, err := provider.FetchHeroes("1", "8")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestD2PTHeroesProvider_FetchHeroes_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	_, err := provider.FetchHeroes("1", "8")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestD2PTHeroesProvider_FetchHeroes_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	heroes, err := provider.FetchHeroes("1", "8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(heroes) != 0 {
		t.Errorf("expected 0 heroes, got %d", len(heroes))
	}
}

// --- Cache tests ---

func newTestServer(t *testing.T, hitCounter *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCounter.Add(1)
		heroes := []Hero{
			{HeroID: 1, HeroName: "Anti-Mage", D2PTRating: 100, Matches: 500},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(heroes)
	}))
}

func TestD2PTHeroesProvider_CacheHit(t *testing.T) {
	var hits atomic.Int32
	server := newTestServer(t, &hits)
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 10*time.Minute)

	_, err := provider.FetchHeroes("1", "8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = provider.FetchHeroes("1", "8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hits.Load() != 1 {
		t.Errorf("expected 1 server hit (cache should serve second call), got %d", hits.Load())
	}
}

func TestD2PTHeroesProvider_CacheDifferentKeys(t *testing.T) {
	var hits atomic.Int32
	server := newTestServer(t, &hits)
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 10*time.Minute)

	_, _ = provider.FetchHeroes("1", "8")
	_, _ = provider.FetchHeroes("2", "patch")
	_, _ = provider.FetchHeroes("1", "8")
	_, _ = provider.FetchHeroes("2", "patch")

	if hits.Load() != 2 {
		t.Errorf("expected 2 server hits (one per unique key), got %d", hits.Load())
	}
}

func TestD2PTHeroesProvider_CacheExpiry(t *testing.T) {
	var hits atomic.Int32
	server := newTestServer(t, &hits)
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 1*time.Millisecond)

	_, _ = provider.FetchHeroes("1", "8")
	time.Sleep(5 * time.Millisecond)
	_, _ = provider.FetchHeroes("1", "8")

	if hits.Load() != 2 {
		t.Errorf("expected 2 server hits (cache should have expired), got %d", hits.Load())
	}
}

func TestD2PTHeroesProvider_NoCacheWhenTTLZero(t *testing.T) {
	var hits atomic.Int32
	server := newTestServer(t, &hits)
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	_, _ = provider.FetchHeroes("1", "8")
	_, _ = provider.FetchHeroes("1", "8")
	_, _ = provider.FetchHeroes("1", "8")

	if hits.Load() != 3 {
		t.Errorf("expected 3 server hits (no caching with ttl=0), got %d", hits.Load())
	}
}

func TestD2PTHeroesProvider_FetchErrorNotCached(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 10*time.Minute)

	_, _ = provider.FetchHeroes("1", "8")
	_, _ = provider.FetchHeroes("1", "8")

	if hits.Load() != 2 {
		t.Errorf("expected 2 server hits (errors should not be cached), got %d", hits.Load())
	}
}

// --- Helper function tests (migrated from d2pt_test.go) ---

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

	originalFirstID := heroes[0].HeroID

	_ = GetTopHeroesByRating(heroes, 2)

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

func TestAggregateHeroesByID_MultipleFacets(t *testing.T) {
	heroes := []Hero{
		{HeroID: 1, HeroName: "Anti-Mage", Matches: 100, Wins: 50, D2PTRating: 60, FacetName: "Facet A", FacetNumber: 1},
		{HeroID: 1, HeroName: "Anti-Mage", Matches: 50, Wins: 30, D2PTRating: 90, FacetName: "Facet B", FacetNumber: 2},
		{HeroID: 2, HeroName: "Axe", Matches: 200, Wins: 110, D2PTRating: 100, FacetName: "Facet C", FacetNumber: 1},
	}

	aggregated := AggregateHeroesByID(heroes)

	if len(aggregated) != 2 {
		t.Fatalf("expected 2 heroes after aggregation, got %d", len(aggregated))
	}

	var antiMage *Hero
	for i := range aggregated {
		if aggregated[i].HeroID == 1 {
			antiMage = &aggregated[i]
			break
		}
	}

	if antiMage == nil {
		t.Fatal("Anti-Mage not found in aggregated results")
	}

	if antiMage.Matches != 150 {
		t.Errorf("expected 150 matches (100+50), got %d", antiMage.Matches)
	}
	if antiMage.Wins != 80 {
		t.Errorf("expected 80 wins (50+30), got %d", antiMage.Wins)
	}
	if antiMage.D2PTRating != 70 {
		t.Errorf("expected D2PT rating 70, got %d", antiMage.D2PTRating)
	}
	if antiMage.FacetName != "" {
		t.Errorf("expected empty FacetName after aggregation, got %s", antiMage.FacetName)
	}
	if antiMage.FacetNumber >= 0 {
		t.Errorf("expected invalid FacetNumber after aggregation, got %d", antiMage.FacetNumber)
	}
}

func TestAggregateHeroesByID_EmptySlice(t *testing.T) {
	heroes := []Hero{}
	aggregated := AggregateHeroesByID(heroes)

	if len(aggregated) != 0 {
		t.Errorf("expected 0 heroes for empty input, got %d", len(aggregated))
	}
}

func TestAggregateHeroesByID_SingleHero(t *testing.T) {
	heroes := []Hero{
		{HeroID: 1, HeroName: "Anti-Mage", Matches: 100, Wins: 50, FacetName: "Facet A", FacetNumber: 1},
	}

	aggregated := AggregateHeroesByID(heroes)

	if len(aggregated) != 1 {
		t.Fatalf("expected 1 hero, got %d", len(aggregated))
	}
	if aggregated[0].Matches != 100 {
		t.Errorf("expected 100 matches, got %d", aggregated[0].Matches)
	}
	if aggregated[0].FacetName != "" {
		t.Errorf("expected empty FacetName, got %s", aggregated[0].FacetName)
	}
}

func TestAggregateHeroesByID_DoesNotModifyOriginal(t *testing.T) {
	heroes := []Hero{
		{HeroID: 1, HeroName: "Anti-Mage", Matches: 100, FacetName: "Facet A"},
		{HeroID: 1, HeroName: "Anti-Mage", Matches: 50, FacetName: "Facet B"},
	}

	originalFacetName := heroes[0].FacetName
	originalMatches := heroes[0].Matches

	_ = AggregateHeroesByID(heroes)

	if heroes[0].FacetName != originalFacetName {
		t.Error("AggregateHeroesByID modified the original slice FacetName")
	}
	if heroes[0].Matches != originalMatches {
		t.Error("AggregateHeroesByID modified the original slice Matches")
	}
}

// --- Benchmarks ---

func BenchmarkGetTopHeroesByRating(b *testing.B) {
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
