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

// Packages available for your implementation — the blank identifier keeps them importable
// without a "declared and not used" compile error while the body is unimplemented.
var (
	_ = os.Getenv    // reads an env var by name; returns "" if unset — https://pkg.go.dev/os#Getenv
	_ = errors.New   // creates a plain error value from a string — https://pkg.go.dev/errors#New
	_ = strconv.Atoi // parses a decimal string as int, returns (int, error) — https://pkg.go.dev/strconv#Atoi
)

// Config holds all values loaded from the environment.
// Exported fields (uppercase) are readable by any package that imports config.
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
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - Multiple return values: Go functions return (value, error) instead of throwing.
//     https://go.dev/tour/basics/6
//   - Error handling: check every error explicitly with `if err != nil`.
//     https://go.dev/blog/error-handling-and-go
//   - os.Getenv: reads a named environment variable; returns "" if unset.
//     https://pkg.go.dev/os#Getenv
//   - errors.New: creates a new error value with a message.
//     https://pkg.go.dev/errors#New
//   - strconv.Atoi: converts a string to int, returns (int, error).
//     https://pkg.go.dev/strconv#Atoi
//   - Struct literal: Config{Field: value, OtherField: otherValue}
//     https://go.dev/tour/basics/5
//
// Steps:
//  1. Read "HF_TOKEN" with os.Getenv. If empty, return Config{} and an error:
//     errors.New("HF_TOKEN environment variable is required")
//  2. Read "BRAVE_API_KEY". Return an error if empty (same pattern as step 1).
//  3. Read "PORT". If empty, use the DefaultPort constant.
//  4. Read "HF_MODEL". If empty, use DefaultModel.
//  5. Read "DATASET_NAME". If empty, use DefaultDataset.
//  6. Read "DATASET_LIMIT". If non-empty, convert it to int with strconv.Atoi.
//     If Atoi returns an error, return Config{} and errors.New("DATASET_LIMIT must be an integer").
//     If empty, use DefaultLimit.
//  7. Return a fully populated Config{...} and nil as the error.
func Load() (Config, error) {
	panic("not implemented")
}
