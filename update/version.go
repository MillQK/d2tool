package update

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

func normalizeVersion(value string) (string, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "v") {
		value = "v" + value
	}
	if !semver.IsValid(value) {
		return "", fmt.Errorf("invalid semantic version %q", value)
	}
	return value, nil
}

func compareVersions(latest, current string) (int, error) {
	latestNormalized, err := normalizeVersion(latest)
	if err != nil {
		return 0, err
	}
	currentNormalized, err := normalizeVersion(current)
	if err != nil {
		return 0, err
	}
	return semver.Compare(latestNormalized, currentNormalized), nil
}

func isUpdateAvailable(latest, current string) bool {
	comparison, err := compareVersions(latest, current)
	return err == nil && comparison > 0
}
