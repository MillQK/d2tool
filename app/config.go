package app

import (
	"d2tool/heroesGrid"
	"d2tool/steam"
	"fyne.io/fyne/v2"
	"log/slog"
	"sync"
)

const (
	heroesGridFilePathsConfigKey      = "heroesGridFilePaths"
	heroesGridPositionsOrderConfigKey = "heroesGridPositionsOrder"

	lastUpdateTimestampMillisKey = "lastUpdateTimestampMillis"
	lastUpdateErrorMessageKey    = "lastUpdateErrorMessage"
)

type ConfigBindings struct {
	HeroesGridFilePaths ValueProvider[[]string]
	PositionsOrder      ValueProvider[[]string]

	LastUpdateTimestampMillis ValueProvider[int]
	LastUpdateErrorMessage    ValueProvider[string]
}

type ValueProvider[T any] interface {
	Get() T
	Set(T)
}

type preferencesValueProviderImpl[T any] struct {
	getter func() T
	setter func(T)

	lock  sync.RWMutex
	cache *T
}

func (v *preferencesValueProviderImpl[T]) Get() T {
	v.lock.RLock()
	defer v.lock.RUnlock()
	if v.cache == nil {
		val := v.getter()
		v.cache = &val
	}
	return *v.cache
}

func (v *preferencesValueProviderImpl[T]) Set(val T) {
	v.lock.RLock()
	defer v.lock.RUnlock()
	v.cache = &val
	v.setter(val)
}

func GetConfigBindings(preferences fyne.Preferences) *ConfigBindings {
	var defaultHeroesGridFilePaths []string
	steamPath, err := steam.FindSteamPath()
	if err == nil {
		paths, err := heroesGrid.FindHeroGridConfigFiles(steamPath)
		if err == nil {
			defaultHeroesGridFilePaths = paths
		} else {
			slog.Warn("Error finding hero grid config files", "error", err)
		}
	} else {
		slog.Warn("Error finding Steam path", "error", err)
	}

	return &ConfigBindings{
		HeroesGridFilePaths: &preferencesValueProviderImpl[[]string]{
			getter: func() []string {
				return preferences.StringListWithFallback(heroesGridFilePathsConfigKey, defaultHeroesGridFilePaths)
			},
			setter: func(paths []string) {
				preferences.SetStringList(heroesGridFilePathsConfigKey, paths)
			},
		},
		PositionsOrder: &preferencesValueProviderImpl[[]string]{
			getter: func() []string {
				return preferences.StringListWithFallback(heroesGridPositionsOrderConfigKey, []string{"1", "2", "3", "4", "5"})
			},
			setter: func(positions []string) {
				preferences.SetStringList(heroesGridPositionsOrderConfigKey, positions)
			},
		},
		LastUpdateTimestampMillis: &preferencesValueProviderImpl[int]{
			getter: func() int {
				return preferences.IntWithFallback(lastUpdateTimestampMillisKey, 0)
			},
			setter: func(millis int) {
				preferences.SetInt(lastUpdateTimestampMillisKey, millis)
			},
		},
		LastUpdateErrorMessage: &preferencesValueProviderImpl[string]{
			getter: func() string {
				return preferences.StringWithFallback(lastUpdateErrorMessageKey, "")
			},
			setter: func(msg string) {
				preferences.SetString(lastUpdateErrorMessageKey, msg)
			},
		},
	}
}
