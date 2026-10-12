package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	PublicURL       string
	ListenAddr      string
	RateLimitBurst  uint64
	RateLimitRefill float64
	StateDir        string
	AdminPassword   string
	DraftPassword   string
	Debug           bool
}

func FromEnv() (Config, error) {
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = "127.0.0.1:6571"
	}

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		publicURL = "http://" + address
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

	adminPassword := os.Getenv("ADMIN_PASS")
	if adminPassword == "" {
		return Config{}, fmt.Errorf("missing admin password")
	}

	draftPassword := os.Getenv("DRAFT_PASS")
	if draftPassword == "" {
		return Config{}, fmt.Errorf("missing draft password")
	}

	return Config{
		PublicURL:       publicURL,
		ListenAddr:      address,
		RateLimitBurst:  rateLimitBurst,
		RateLimitRefill: rateLimitRefill,
		StateDir:        stateDir,
		AdminPassword:   adminPassword,
		DraftPassword:   draftPassword,
		Debug:           Debug == "true",
	}, nil
}

// Debug is set to "true" by the development build's linker flags.
var Debug = "false"

// set during release build
var Version = "dev"
