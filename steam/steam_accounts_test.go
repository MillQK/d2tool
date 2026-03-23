package steam

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func setupTestSteamDir(t *testing.T) string {
	t.Helper()
	steamDir := t.TempDir()

	vdfContent := `"users"
{
	"76561197960278073"
	{
		"AccountName"		"testuser"
		"PersonaName"		"TestPlayer"
	}
	"76561197960278074"
	{
		"AccountName"		"testuser2"
		"PersonaName"		"TestPlayer2"
	}
}
`
	configDir := filepath.Join(steamDir, "config")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "loginusers.vdf"), []byte(vdfContent), 0644)

	// SteamID3 for 76561197960278073 = 12345
	userDataDir := filepath.Join(steamDir, "userdata", "12345", "570", "remote", "cfg")
	os.MkdirAll(userDataDir, 0755)
	os.WriteFile(filepath.Join(userDataDir, "hero_grid_config.json"), []byte("{}"), 0644)

	// Second user (SteamID3 = 12346) has NO hero_grid_config.json

	return steamDir
}

func TestScanAccounts_DiscoverFromVDFAndFiles(t *testing.T) {
	steamDir := setupTestSteamDir(t)
	accounts, err := ScanAccounts(steamDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected 1 discovered account, got %d", len(accounts))
	}
	acc := accounts[0]
	if acc.SteamID64 != "76561197960278073" {
		t.Errorf("expected SteamID64 76561197960278073, got %s", acc.SteamID64)
	}
	if acc.SteamID3 != "12345" {
		t.Errorf("expected SteamID3 12345, got %s", acc.SteamID3)
	}
	if acc.AccountName != "testuser" {
		t.Errorf("expected AccountName testuser, got %s", acc.AccountName)
	}
	if acc.PersonaName != "TestPlayer" {
		t.Errorf("expected PersonaName TestPlayer, got %s", acc.PersonaName)
	}
}

func TestScanAccounts_LoadAvatar(t *testing.T) {
	steamDir := setupTestSteamDir(t)
	avatarDir := filepath.Join(steamDir, "config", "avatarcache")
	os.MkdirAll(avatarDir, 0755)
	fakeAvatar := []byte("fake-png-data")
	os.WriteFile(filepath.Join(avatarDir, "76561197960278073.png"), fakeAvatar, 0644)

	accounts, err := ScanAccounts(steamDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatal("expected 1 account")
	}
	expectedBase64 := base64.StdEncoding.EncodeToString(fakeAvatar)
	if accounts[0].AvatarBase64 != expectedBase64 {
		t.Errorf("avatar base64 mismatch")
	}
}

func TestScanAccounts_NoAvatar(t *testing.T) {
	steamDir := setupTestSteamDir(t)
	accounts, err := ScanAccounts(steamDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accounts[0].AvatarBase64 != "" {
		t.Error("expected empty avatar when file does not exist")
	}
}

func TestScanAccounts_InvalidSteamPath(t *testing.T) {
	_, err := ScanAccounts("/nonexistent/path")
	if err == nil {
		t.Error("expected error for invalid steam path")
	}
}

func TestScanAccounts_MissingVDF(t *testing.T) {
	steamDir := t.TempDir()
	os.MkdirAll(filepath.Join(steamDir, "userdata"), 0755)
	_, err := ScanAccounts(steamDir)
	if err == nil {
		t.Error("expected error when loginusers.vdf is missing")
	}
}
