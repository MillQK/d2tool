package steam

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/andygrunwald/vdf"
)

const (
	steamId64Indent = 76561197960265728
)

type SteamHeroesLayoutConfigFileInfo struct {
	Path        string
	SteamID3    string
	SteamID64   string
	AccountName string
	PersonaName string
}

// FindSteamHeroesLayoutConfigFiles finds all hero_grid_config.json files for all Steam users
func FindSteamHeroesLayoutConfigFiles(steamPath string) ([]SteamHeroesLayoutConfigFileInfo, error) {
	userdataPath := filepath.Join(steamPath, "userdata")

	// Check if userdata directory exists
	if _, err := os.Stat(userdataPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("userdata directory not found at %s", userdataPath)
	}

	var loginUsersContent map[string]interface{}

	loginUsersFile, err := os.Open(filepath.Join(steamPath, "config", "loginusers.vdf"))
	if err != nil {
		slog.Warn("Error opening loginusers.vdf file", "error", err)
	} else {
		defer loginUsersFile.Close()
		parser := vdf.NewParser(loginUsersFile)
		loginUsersContent, err = parser.Parse()
		if err != nil {
			slog.Warn("Error parsing loginusers.vdf file", "error", err)
		} else {
			slog.Debug("Parsed loginusers.vdf content", "loginUsersContent", loginUsersContent)
		}
	}

	var configFiles []SteamHeroesLayoutConfigFileInfo

	// List all directories in userdata (each is a Steam user ID)
	userDirs, err := os.ReadDir(userdataPath)
	if err != nil {
		return nil, fmt.Errorf("error reading userdata directory: %w", err)
	}

	for _, userDir := range userDirs {
		if !userDir.IsDir() {
			continue
		}

		// Construct path to hero_grid_config.json
		configPath := filepath.Join(userdataPath, userDir.Name(), "570", "remote", "cfg", "hero_grid_config.json")

		// Check if file exists
		if _, err := os.Stat(configPath); err == nil {
			configFileInfo := SteamHeroesLayoutConfigFileInfo{
				Path:     configPath,
				SteamID3: userDir.Name(),
			}

			steamId3Num, err := strconv.ParseUint(userDir.Name(), 10, 64)
			if err != nil {
				slog.Warn("Error parsing Steam ID 3 number", "error", err, "steamId3", userDir.Name())
			} else {
				configFileInfo.SteamID64 = strconv.FormatUint(steamId3Num+steamId64Indent, 10)
				if usersMapInterface, ok := loginUsersContent["users"]; ok {
					if usersMap, ok := usersMapInterface.(map[string]interface{}); ok {
						if userMapInterface, ok := usersMap[configFileInfo.SteamID64]; ok {
							if userMap, ok := userMapInterface.(map[string]interface{}); ok {
								if accountNameInterface, ok := userMap["AccountName"]; ok {
									if accountName, ok := accountNameInterface.(string); ok {
										configFileInfo.AccountName = accountName
									}
								}
								if personaNameInterface, ok := userMap["PersonaName"]; ok {
									if personaName, ok := personaNameInterface.(string); ok {
										configFileInfo.PersonaName = personaName
									}
								}
							}
						}
					}
				}
			}

			configFiles = append(configFiles, configFileInfo)
		}
	}

	if len(configFiles) == 0 {
		return nil, fmt.Errorf("no hero_grid_config.json files found")
	}

	return configFiles, nil
}

func (c SteamHeroesLayoutConfigFileInfo) ToAttributesMap() map[string]string {
	attributes := map[string]string{}
	if c.PersonaName != "" {
		attributes["Alias"] = c.PersonaName
	}

	if c.AccountName != "" {
		attributes["Account Name"] = c.AccountName
	}

	if c.SteamID3 != "" {
		attributes["SteamID3"] = c.SteamID3
	}

	if c.SteamID64 != "" {
		attributes["SteamID64"] = c.SteamID64
	}

	return attributes
}
