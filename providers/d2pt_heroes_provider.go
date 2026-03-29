package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	apiD2ptUrl  = "https://dota2protracker.com/api"
	period8Days = "8"
	periodPatch = "patch"
)

type cacheKey struct {
	position string
	period   string
}

type cacheEntry struct {
	heroes    []Hero
	fetchedAt time.Time
}

type D2PTHeroesProvider struct {
	httpClient *http.Client
	apiUrl     string
	ttl        time.Duration

	mu    sync.RWMutex
	cache map[cacheKey]cacheEntry
}

func NewD2PTHeroesProvider(httpClient *http.Client, apiUrl string, ttl time.Duration) *D2PTHeroesProvider {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return &D2PTHeroesProvider{
		httpClient: httpClient,
		apiUrl:     apiUrl,
		ttl:        ttl,
		cache:      make(map[cacheKey]cacheEntry),
	}
}

func (p *D2PTHeroesProvider) FetchHeroes(position string, period string) ([]Hero, error) {
	key := cacheKey{position: position, period: period}

	if p.ttl > 0 {
		p.mu.RLock()
		if entry, ok := p.cache[key]; ok {
			if time.Since(entry.fetchedAt) < p.ttl {
				p.mu.RUnlock()
				return entry.heroes, nil
			}
		}
		p.mu.RUnlock()
	}

	heroes, err := p.fetchFromAPI(position, period)
	if err != nil {
		return nil, err
	}

	if p.ttl > 0 {
		p.mu.Lock()
		p.cache[key] = cacheEntry{
			heroes:    heroes,
			fetchedAt: time.Now(),
		}
		p.mu.Unlock()
	}

	return heroes, nil
}

func (p *D2PTHeroesProvider) fetchFromAPI(position string, period string) ([]Hero, error) {
	if period == "" {
		period = period8Days
	}

	if period != period8Days && period != periodPatch {
		return nil, fmt.Errorf("invalid period value: %s", period)
	}

	params := url.Values{
		"mmr":         {"7000"},
		"order_by":    {"matches"},
		"min_matches": {"20"},
		"period":      {period},
		"position":    {position},
		"legacy":      {"false"},
	}

	apiUrl := strings.TrimRight(p.apiUrl, "/")
	if apiUrl == "" {
		apiUrl = apiD2ptUrl
	}

	d2ptUrl := fmt.Sprintf("%s/heroes/stats?%s", apiUrl, params.Encode())

	req, err := http.NewRequest("GET", d2ptUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Referer", "https://dota2protracker.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36")

	resp, err := p.httpClient.Do(req)
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
