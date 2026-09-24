package config

import (
	"os"
	"strconv"
)

type Config struct {
	ListenAddr      string
	RateLimitBurst  uint64
	RateLimitRefill float64
	StateDir        string
	Debug           bool
}

func FromEnv() Config {
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = "127.0.0.1:6571"
	}

	rateLimitBurst, err := strconv.ParseUint(os.Getenv("RATE_LIMIT_BURST"), 10, 32)
	if err != nil {
		rateLimitBurst = 500
	}

	rateLimitRefill, err := strconv.ParseFloat(os.Getenv("RATE_LIMIT_REFILL"), 32)
	if err != nil {
		rateLimitRefill = 10 // per second
	}

	stateDir := os.Getenv("STATE_DIR")
	if stateDir == "" {
		stateDir = "."
	}

	return Config{
		ListenAddr:      address,
		RateLimitBurst:  rateLimitBurst,
		RateLimitRefill: rateLimitRefill,
		StateDir:        stateDir,
		Debug:           Debug == "true",
	}
}

// Debug is set to "true" by the development build's linker flags.
var Debug = "false"
