package config

import (
	"os"
)

type Config struct {
	ListenAddr string
	Debug      bool
}

func FromEnv() Config {
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = "127.0.0.1:6571"
	}

	return Config{
		ListenAddr: address,
		Debug:      Debug == "true",
	}
}

// Debug is set to "true" by the development build's linker flags.
var Debug = "false"
