package util

import "flag"

type Config struct {
	Discover *DiscoverConfig
	Mesh     *MeshConfig
	Streamer *StreamerConfig
}

type DiscoverConfig struct {
	Port int
}

type MeshConfig struct {
	AutoAccept bool
}

type StreamerConfig struct {
	AutoStart bool
}

func InitConfig() *Config {
	discoverPort := flag.Int("port", 19416, "Server port")
	autoAccept := flag.Bool("auto-accept", false, "Auto accept discovered devices")
	autoStartStream := flag.Bool("auto-start-stream", false, "Auto start audio stream")

	flag.Parse()

	discoverConfig := &DiscoverConfig{Port: *discoverPort}
	meshConfig := &MeshConfig{AutoAccept: *autoAccept}
	streamerConfig := &StreamerConfig{AutoStart: *autoStartStream}

	config := &Config{Discover: discoverConfig, Mesh: meshConfig, Streamer: streamerConfig}

	return config
}
