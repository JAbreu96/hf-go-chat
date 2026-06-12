// Package config loads all runtime configuration from environment variables.
// Go has no built-in .env loader — we use os.Getenv directly here.
// To auto-load a .env file, see: https://github.com/joho/godotenv
package config

import (
	"errors"
	"os"
	"strconv"
)

// DefaultModel is a constant — value fixed at compile time, stored in read-only memory.
// Convention: package-level constants use MixedCaps, not SCREAMING_SNAKE_CASE.
// https://go.dev/ref/spec#Constant_declarations
const (
	DefaultModel      = "Qwen/Qwen2.5-7B-Instruct-1M"
	DefaultPort       = "8080"
	DefaultDataset    = "rajpurkar/squad"
	DefaultDatasetCfg = "plain_text"
	DefaultLimit      = 1000
)

// Config holds all values loaded from the environment.
// Unexported fields would be hidden from other packages — here we export all
// fields so handlers can read them directly.
// https://go.dev/ref/spec#Exported_identifiers
type Config struct {
	HFToken      string
	BraveKey     string
	Port         string
	Model        string
	DatasetName  string
	DatasetCfg   string
	DatasetLimit int
}

// Load reads environment variables and returns a populated Config.
// Multiple return values — Go functions return (value, error) instead of throwing.
// The caller must check the error before using cfg.
// https://go.dev/tour/basics/6
func Load() (Config, error) {
	token := os.Getenv("HF_TOKEN")
	// if err != nil — the canonical Go error check. There is no try/catch.
	// https://go.dev/blog/error-handling-and-go
	if token == "" {
		return Config{}, errors.New("HF_TOKEN environment variable is required")
	}

	brave := os.Getenv("BRAVE_API_KEY")
	if brave == "" {
		return Config{}, errors.New("BRAVE_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	model := os.Getenv("HF_MODEL")
	if model == "" {
		model = DefaultModel
	}

	dataset := os.Getenv("DATASET_NAME")
	if dataset == "" {
		dataset = DefaultDataset
	}

	limit := DefaultLimit
	if raw := os.Getenv("DATASET_LIMIT"); raw != "" {
		// strconv.Atoi returns (int, error) — two return values in one assignment.
		// The := operator declares both n and err as new variables in this scope.
		n, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, errors.New("DATASET_LIMIT must be an integer")
		}
		limit = n
	}

	return Config{
		HFToken:      token,
		BraveKey:     brave,
		Port:         port,
		Model:        model,
		DatasetName:  dataset,
		DatasetCfg:   DefaultDatasetCfg,
		DatasetLimit: limit,
	}, nil
}
