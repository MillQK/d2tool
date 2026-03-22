package steamid

// Id64Offset is the fixed offset between SteamID3 and SteamID64.
const Id64Offset uint64 = 76561197960265728

func ID3toID64(id3 uint64) uint64  { return id3 + Id64Offset }
func ID64toID3(id64 uint64) uint64 { return id64 - Id64Offset }
