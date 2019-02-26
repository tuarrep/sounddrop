package util

import "flag"

// Config store the application
type Config struct {
	Discover *DiscoverConfig
	Mesh     *MeshConfig
	Streamer *StreamerConfig
}

// DiscoverConfig peer discovering config
type DiscoverConfig struct {
	Port int
}

// MeshConfig mesh network config
type MeshConfig struct {
	AutoAccept bool
}

// StreamerConfig streamer config
type StreamerConfig struct {
	AutoStart   bool
	PlaylistDir string
}

// InitConfig load config from flags
func InitConfig() *Config {
	discoverPort := flag.Int("port", 19416, "Server port")

	autoAccept := flag.Bool("auto-accept", false, "Auto accept discovered devices")

	autoStartStream := flag.Bool("auto-start-stream", false, "Auto start audio stream")
	playlistDir := flag.String("playlist-dir", ".", "Directory containing audio files to play")

	flag.Parse()

	discoverConfig := &DiscoverConfig{Port: *discoverPort}
	meshConfig := &MeshConfig{AutoAccept: *autoAccept}
	streamerConfig := &StreamerConfig{AutoStart: *autoStartStream, PlaylistDir: *playlistDir}

	config := &Config{Discover: discoverConfig, Mesh: meshConfig, Streamer: streamerConfig}

	return config
}
