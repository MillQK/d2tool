package heroesLayout

import (
	"d2tool/config"
	"d2tool/providers"
	"d2tool/steam"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestD2PTGridMatchesMetaTable(t *testing.T) {
	// Position 1, 7000 MMR, patch 7.42. The API lists rows in hero ID order.
	const rows = `[
		{"hero_id":12, "matches":1450, "meta_score":85.7},
		{"hero_id":18, "matches":1422, "meta_score":80.3},
		{"hero_id":32, "matches":470,  "meta_score":57.4},
		{"hero_id":42, "matches":267,  "meta_score":42.1},
		{"hero_id":44, "matches":1301, "meta_score":66.5},
		{"hero_id":48, "matches":1107, "meta_score":58.4},
		{"hero_id":54, "matches":813,  "meta_score":48.4},
		{"hero_id":67, "matches":304,  "meta_score":44.8},
		{"hero_id":93, "matches":878,  "meta_score":55.7},
		{"hero_id":95, "matches":216,  "meta_score":46.8},
		{"hero_id":109,"matches":549,  "meta_score":46.2}
	]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/meta/api/section" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("section") != "meta" || q.Get("meta_position") != "pos 1" || q.Get("meta_mmr") != "7000" || q.Get("meta_period") != "patch" {
			t.Errorf("unexpected meta parameters %v", q)
		}
		w.Write([]byte(`{"patch":"7.42","meta":` + rows + `}`))
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "hero_grid_config.json")
	if err := os.WriteFile(path, []byte(`{"version":3,"configs":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		HeroesLayout: config.HeroesLayoutConfig{
			Files:        []config.FileConfig{{FilePath: path, Enabled: true}},
			Positions:    []config.PositionConfig{{ID: "1", Enabled: true}},
			HeroesPerRow: 10,
		},
		D2PT: config.D2PTConfig{Period: "patch"},
	}
	service := NewHeroesLayoutService(cfg, steam.NewSteamService(cfg), providers.NewD2PTHeroesProvider(server.Client(), server.URL, 0))
	if err := service.UpdateHeroesLayout(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var grid heroGridConfig
	if err := json.Unmarshal(data, &grid); err != nil {
		t.Fatal(err)
	}
	if len(grid.Configs) != 1 {
		t.Fatalf("expected one generated config, got %d", len(grid.Configs))
	}
	var got []int
	for _, category := range grid.Configs[0].Categories {
		if category.CategoryName == "pos 1 - Most Matches Heroes" {
			break
		}
		got = append(got, category.HeroIDs...)
	}
	if want := []int{12, 18, 44, 48, 32, 93, 54, 95, 109, 67}; !slices.Equal(got, want) {
		t.Errorf("top 10 hero IDs = %v, want %v", got, want)
	}
}
