package providers

import (
	"d2tool/config"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const d2ptURL = "https://dota2protracker.com"

type cacheKey struct {
	position string
	period   string
}

type cacheEntry struct {
	stats     PositionStats
	fetchedAt time.Time
}

type D2PTHeroesProvider struct {
	httpClient *http.Client
	baseURL    string
	ttl        time.Duration

	mu    sync.RWMutex
	cache map[cacheKey]cacheEntry
}

func NewD2PTHeroesProvider(httpClient *http.Client, baseURL string, ttl time.Duration) *D2PTHeroesProvider {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if baseURL == "" {
		baseURL = d2ptURL
	}
	return &D2PTHeroesProvider{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
		ttl:        ttl,
		cache:      make(map[cacheKey]cacheEntry),
	}
}

func (p *D2PTHeroesProvider) FetchPositions(positions []string, period string) (map[string]PositionStats, error) {
	switch period {
	case config.D2PTPeriodAuto:
		var timeline []patchRelease
		if err := p.getJSON("/api/patches/timeline/stats", &timeline); err != nil {
			slog.Warn("Could not load D2PT patch timeline, using last 8 days", "error", err)
			period = config.D2PTPeriod8Days
			break
		}
		period = autoPeriod(timeline, time.Now())
	case "":
		period = config.D2PTPeriod8Days
	case config.D2PTPeriodPatch, config.D2PTPeriod8Days:
	default:
		return nil, fmt.Errorf("invalid D2PT period: %q", period)
	}

	stats := make(map[string]PositionStats, len(positions))
	for _, position := range positions {
		rows, err := p.fetchHeroes(cacheKey{position: position, period: period})
		if err != nil {
			return nil, fmt.Errorf("fetching heroes for %s: %w", position, err)
		}
		stats[position] = rows
	}
	return stats, nil
}

type patchRelease struct {
	Timestamp int64 `json:"release_date_ts"`
}

func autoPeriod(timeline []patchRelease, now time.Time) string {
	var latest int64
	for _, patch := range timeline {
		if patch.Timestamp <= now.Unix() && patch.Timestamp > latest {
			latest = patch.Timestamp
		}
	}
	released := time.Unix(latest, 0)
	if latest > 0 && now.Sub(released) < 8*24*time.Hour {
		return config.D2PTPeriodPatch
	}
	return config.D2PTPeriod8Days
}

func (p *D2PTHeroesProvider) fetchHeroes(key cacheKey) (PositionStats, error) {
	if p.ttl > 0 {
		p.mu.RLock()
		entry, ok := p.cache[key]
		p.mu.RUnlock()
		if ok && time.Since(entry.fetchedAt) < p.ttl {
			return entry.stats, nil
		}
	}

	params := url.Values{
		"section":       {"meta"},
		"meta_mmr":      {"7000"},
		"meta_position": {key.position},
		"meta_period":   {key.period},
	}
	var section struct {
		Meta       []Hero          `json:"meta"`
		Patch      json.RawMessage `json:"patch"`
		ScoreError string          `json:"metaScoreDataError"`
	}
	if err := p.getJSON("/meta/api/section?"+params.Encode(), &section); err != nil {
		return PositionStats{}, err
	}
	if section.ScoreError != "" {
		return PositionStats{}, fmt.Errorf("D2PT meta scores unavailable: %s", section.ScoreError)
	}
	if section.Meta == nil {
		return PositionStats{}, fmt.Errorf("D2PT response is missing the meta table")
	}
	stats := PositionStats{Heroes: section.Meta}
	// Without a current patch label the website defaults to a 200-match minimum.
	switch string(section.Patch) {
	case "", "null", "false", `""`:
		stats.minRatingMatches = 200
	}
	if p.ttl > 0 {
		p.mu.Lock()
		p.cache[key] = cacheEntry{stats: stats, fetchedAt: time.Now()}
		p.mu.Unlock()
	}
	return stats, nil
}

func (p *D2PTHeroesProvider) getJSON(path string, result any) error {
	req, err := http.NewRequest(http.MethodGet, p.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", d2ptURL)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding D2PT response: %w", err)
	}
	return nil
}
