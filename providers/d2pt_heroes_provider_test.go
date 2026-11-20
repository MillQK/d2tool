package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func score(value float64) *float64 { return &value }

func TestD2PTHeroesProvider_FetchPositions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/meta/api/section" {
			t.Errorf("expected path /meta/api/section, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("meta_position") != "pos 1" {
			t.Errorf("expected meta_position=pos 1, got %s", r.URL.Query().Get("meta_position"))
		}
		if r.URL.Query().Get("meta_mmr") != "7000" {
			t.Errorf("expected mmr=7000, got %s", r.URL.Query().Get("meta_mmr"))
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
			{HeroID: 1, HeroName: "Anti-Mage", MetaScore: score(80), Matches: 500},
			{HeroID: 2, HeroName: "Axe", MetaScore: score(90.5), Matches: 300},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"meta": heroes})
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	positions, err := provider.FetchPositions([]string{"pos 1"}, "8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	heroes := positions["pos 1"].Heroes
	if len(heroes) != 2 {
		t.Fatalf("expected 2 heroes, got %d", len(heroes))
	}
	if heroes[0].HeroID != 1 {
		t.Errorf("expected first hero ID 1, got %d", heroes[0].HeroID)
	}
	if heroes[0].HeroName != "Anti-Mage" {
		t.Errorf("expected first hero name 'Anti-Mage', got %s", heroes[0].HeroName)
	}
	if heroes[1].rating() != 90.5 {
		t.Errorf("expected second hero rating 90.5, got %v", heroes[1].rating())
	}
}

func TestD2PTHeroesProvider_FetchPositions_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	_, err := provider.FetchPositions([]string{"pos 1"}, "8")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestD2PTHeroesProvider_FetchPositions_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	_, err := provider.FetchPositions([]string{"pos 1"}, "8")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestD2PTHeroesProvider_FetchPositions_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	_, err := provider.FetchPositions([]string{"pos 1"}, "8")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestD2PTHeroesProvider_FetchPositions_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"meta":[]}`))
	}))
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	positions, err := provider.FetchPositions([]string{"pos 1"}, "8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	heroes := positions["pos 1"].Heroes
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
			{HeroID: 1, HeroName: "Anti-Mage", MetaScore: score(80), Matches: 500},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"meta": heroes})
	}))
}

func TestD2PTHeroesProvider_CacheHit(t *testing.T) {
	var hits atomic.Int32
	server := newTestServer(t, &hits)
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 10*time.Minute)

	_, err := provider.FetchPositions([]string{"pos 1"}, "8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = provider.FetchPositions([]string{"pos 1"}, "8")
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

	_, _ = provider.FetchPositions([]string{"pos 1"}, "8")
	_, _ = provider.FetchPositions([]string{"pos 2"}, "patch")
	_, _ = provider.FetchPositions([]string{"pos 1"}, "8")
	_, _ = provider.FetchPositions([]string{"pos 2"}, "patch")

	if hits.Load() != 2 {
		t.Errorf("expected 2 server hits (one per unique key), got %d", hits.Load())
	}
}

func TestD2PTHeroesProvider_CacheExpiry(t *testing.T) {
	var hits atomic.Int32
	server := newTestServer(t, &hits)
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 1*time.Millisecond)

	_, _ = provider.FetchPositions([]string{"pos 1"}, "8")
	time.Sleep(5 * time.Millisecond)
	_, _ = provider.FetchPositions([]string{"pos 1"}, "8")

	if hits.Load() != 2 {
		t.Errorf("expected 2 server hits (cache should have expired), got %d", hits.Load())
	}
}

func TestD2PTHeroesProvider_NoCacheWhenTTLZero(t *testing.T) {
	var hits atomic.Int32
	server := newTestServer(t, &hits)
	defer server.Close()

	provider := NewD2PTHeroesProvider(nil, server.URL, 0)

	_, _ = provider.FetchPositions([]string{"pos 1"}, "8")
	_, _ = provider.FetchPositions([]string{"pos 1"}, "8")
	_, _ = provider.FetchPositions([]string{"pos 1"}, "8")

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

	_, _ = provider.FetchPositions([]string{"pos 1"}, "8")
	_, _ = provider.FetchPositions([]string{"pos 1"}, "8")

	if hits.Load() != 2 {
		t.Errorf("expected 2 server hits (errors should not be cached), got %d", hits.Load())
	}
}

func TestTopRatedUsesMetaTableOrder(t *testing.T) {
	stats := PositionStats{Heroes: []Hero{
		{HeroID: 1, Matches: 400},
		{HeroID: 2, Matches: 250, MetaScore: score(0)},
		{HeroID: 3, Matches: 200, MetaScore: score(75.2)},
		{HeroID: 4, Matches: 300, MetaScore: score(75.2)},
		{HeroID: 5, Matches: 199, MetaScore: score(75.3)},
		{HeroID: 6, Matches: 250, MetaScore: score(10), D2PTRatingV2: score(80)},
	}}
	original := slices.Clone(stats.Heroes)
	for _, tt := range []struct {
		minMatches, limit int
		want              []int
	}{
		{0, 3, []int{6, 5, 3}},
		{200, 10, []int{6, 3, 4, 2, 1}},
		{500, 10, nil},
	} {
		stats.minRatingMatches = tt.minMatches
		var got []int
		for _, hero := range stats.TopRated(tt.limit) {
			got = append(got, hero.HeroID)
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("minMatches=%d, limit=%d: got %v, want %v", tt.minMatches, tt.limit, got, tt.want)
		}
	}
	if !slices.Equal(stats.Heroes, original) {
		t.Error("sorting modified the cached rows")
	}
}

func TestMostPlayedAggregatesMatchesAndWins(t *testing.T) {
	stats := PositionStats{Heroes: []Hero{
		{HeroID: 1, Matches: 100, Wins: 50},
		{HeroID: 2, Matches: 120, Wins: 60},
		{HeroID: 1, Matches: 20, Wins: 12},
		{HeroID: 1, Matches: 19, Wins: 19},
		{HeroID: 3, Matches: 200, Wins: 130},
	}}
	original := slices.Clone(stats.Heroes)
	want := []Hero{{HeroID: 3, Matches: 200, Wins: 130}, {HeroID: 1, Matches: 120, Wins: 62}}
	if got := stats.MostPlayed(2); !slices.Equal(got, want) {
		t.Errorf("MostPlayed(2) = %v, want %v", got, want)
	}
	if !slices.Equal(stats.Heroes, original) {
		t.Error("aggregation modified the cached rows")
	}
	if got := (PositionStats{}).MostPlayed(30); len(got) != 0 {
		t.Errorf("expected empty ranking, got %v", got)
	}
}

func TestAutoPeriodBoundary(t *testing.T) {
	released := time.Date(2026, time.November, 18, 21, 0, 19, 0, time.UTC)
	timeline := []patchRelease{
		{Timestamp: released.AddDate(0, 1, 0).Unix()},
		{Timestamp: released.AddDate(0, 0, -25).Unix()},
		{Timestamp: released.Unix()},
	}
	for _, tt := range []struct {
		age  time.Duration
		want string
	}{
		{0, "patch"},
		{8*24*time.Hour - time.Nanosecond, "patch"},
		{8 * 24 * time.Hour, "8"},
		{9 * 24 * time.Hour, "8"},
	} {
		got := autoPeriod(timeline, released.Add(tt.age))
		if got != tt.want {
			t.Errorf("patch age %v: got %q, want %q", tt.age, got, tt.want)
		}
	}
}

func TestFetchPositionsResolvesAutoOncePerUpdate(t *testing.T) {
	now := time.Now()
	var releaseTS atomic.Int64
	var timelineHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/patches/timeline/stats":
			timelineHits.Add(1)
			fmt.Fprintf(w, `[{"release_date_ts":%d}]`, releaseTS.Load())
		case "/meta/api/section":
			period := r.URL.Query().Get("meta_period")
			switch period {
			case "patch":
				fmt.Fprint(w, `{"patch":"7.42","meta":[{"hero_id":1,"matches":19,"meta_score":50.5}]}`)
			case "8":
				fmt.Fprint(w, `{"patch":false,"meta":[{"hero_id":2,"matches":200},{"hero_id":3,"matches":199,"meta_score":99}]}`)
			default:
				t.Errorf("unexpected effective period %q", period)
				http.Error(w, "unknown period", http.StatusBadRequest)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	provider := NewD2PTHeroesProvider(server.Client(), server.URL, time.Hour)
	positions := []string{"pos 1", "pos 5"}
	for i, step := range []struct {
		age    time.Duration
		wantID int
	}{
		{9 * 24 * time.Hour, 2},
		{time.Hour, 1},
	} {
		releaseTS.Store(now.Add(-step.age).Unix())
		stats, err := provider.FetchPositions(positions, "auto")
		if err != nil {
			t.Fatal(err)
		}
		for _, position := range positions {
			if got := stats[position].TopRated(10); len(got) != 1 || got[0].HeroID != step.wantID {
				t.Errorf("update %d, %s: unexpected ranking %v", i, position, got)
			}
		}
		if got := timelineHits.Load(); got != int32(i+1) {
			t.Errorf("after update %d: got %d timeline requests", i, got)
		}
	}
}

func TestFetchPositionsAutoFallsBackTo8Days(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		body   string
	}{
		{"empty timeline", http.StatusOK, `[]`},
		{"missing timestamp", http.StatusOK, `[{}]`},
		{"future patch only", http.StatusOK, fmt.Sprintf(`[{"release_date_ts":%d}]`, time.Now().Add(24*time.Hour).Unix())},
		{"invalid JSON", http.StatusOK, `invalid`},
		{"unavailable timeline", http.StatusServiceUnavailable, `unavailable`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/patches/timeline/stats":
					w.WriteHeader(tt.status)
					fmt.Fprint(w, tt.body)
				case "/meta/api/section":
					if got := r.URL.Query().Get("meta_period"); got != "8" {
						t.Errorf("requested period %q, want 8", got)
					}
					fmt.Fprint(w, `{"meta":[{"hero_id":42}]}`)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			provider := NewD2PTHeroesProvider(server.Client(), server.URL, 0)
			stats, err := provider.FetchPositions([]string{"pos 1"}, "auto")
			if err != nil {
				t.Fatal(err)
			}
			if heroes := stats["pos 1"].Heroes; len(heroes) != 1 || heroes[0].HeroID != 42 {
				t.Errorf("unexpected fallback heroes %v", heroes)
			}
		})
	}
}

// --- Benchmarks ---

func BenchmarkTopRated(b *testing.B) {
	heroes := make([]Hero, 100)
	for i := range heroes {
		heroes[i] = Hero{
			HeroID:    i + 1,
			MetaScore: score(float64(i)),
			Matches:   i * 5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(PositionStats{Heroes: heroes}).TopRated(10)
	}
}

func BenchmarkMostPlayed(b *testing.B) {
	heroes := make([]Hero, 100)
	for i := range heroes {
		heroes[i] = Hero{
			HeroID:    i + 1,
			MetaScore: score(float64(i)),
			Matches:   i * 5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(PositionStats{Heroes: heroes}).MostPlayed(30)
	}
}
