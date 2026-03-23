package steam

import (
	"d2tool/steamid"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/andygrunwald/vdf"
)

type DiscoveredAccount struct {
	SteamID64    string
	SteamID3     string
	AccountName  string
	PersonaName  string
	AvatarBase64 string
}

func ScanAccounts(steamPath string) ([]DiscoveredAccount, error) {
	vdfPath := filepath.Join(steamPath, "config", "loginusers.vdf")
	vdfFile, err := os.Open(vdfPath)
	if err != nil {
		return nil, fmt.Errorf("error opening loginusers.vdf: %w", err)
	}
	defer vdfFile.Close()

	parser := vdf.NewParser(vdfFile)
	vdfContent, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("error parsing loginusers.vdf: %w", err)
	}

	usersMapInterface, ok := vdfContent["users"]
	if !ok {
		return nil, fmt.Errorf("no 'users' key in loginusers.vdf")
	}

	usersMap, ok := usersMapInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("'users' is not a map in loginusers.vdf")
	}

	var accounts []DiscoveredAccount

	for steamId64Str, userDataInterface := range usersMap {
		steamId64, err := strconv.ParseUint(steamId64Str, 10, 64)
		if err != nil {
			slog.Warn("Error parsing SteamID64", "steamId64", steamId64Str, "error", err)
			continue
		}

		steamId3 := steamid.ID64toID3(steamId64)
		steamId3Str := strconv.FormatUint(steamId3, 10)

		configPath := HeroGridConfigPath(steamPath, steamId3Str)
		if _, err := os.Stat(configPath); err != nil {
			continue
		}

		account := DiscoveredAccount{
			SteamID64: steamId64Str,
			SteamID3:  steamId3Str,
		}

		if userMap, ok := userDataInterface.(map[string]interface{}); ok {
			if v, ok := userMap["AccountName"].(string); ok {
				account.AccountName = v
			}
			if v, ok := userMap["PersonaName"].(string); ok {
				account.PersonaName = v
			}
		}

		avatarPath := filepath.Join(steamPath, "config", "avatarcache", steamId64Str+".png")
		if avatarData, err := os.ReadFile(avatarPath); err == nil {
			account.AvatarBase64 = base64.StdEncoding.EncodeToString(avatarData)
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func HeroGridConfigPath(steamPath string, steamId3 string) string {
	return filepath.Join(steamPath, "userdata", steamId3, "570", "remote", "cfg", "hero_grid_config.json")
}
