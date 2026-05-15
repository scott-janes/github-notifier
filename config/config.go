package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	GitHubToken       string   `json:"github_token"`
	GitHubUsername    string   `json:"github_username"`
	PollIntervalMin   int      `json:"poll_interval_minutes"`
	ScheduleDays      []string `json:"schedule_days"`
	ScheduleStartHour int      `json:"schedule_start_hour"`
	ScheduleEndHour   int      `json:"schedule_end_hour"`
	SoundEnabled      bool     `json:"sound_enabled"`
	SoundPath         string   `json:"sound_path"`
	AutoHideSeconds   int      `json:"auto_hide_seconds"`
	DisableDrag       bool     `json:"disable_drag"`
	DNDEnabled        bool     `json:"dnd_enabled"`
	DNDHours          float64  `json:"dnd_hours"`
	PanelOpacity      float64  `json:"panel_opacity"`
	WindowX           int      `json:"window_x"`
	WindowY           int      `json:"window_y"`
	MockMode          bool     `json:"mock_mode"`
}

func DefaultConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("config: os.UserConfigDir() failed: %v, falling back to HOME", err)
		dir = os.Getenv("HOME")
		if dir == "" {
			dir = "/tmp"
		}
	}
	return filepath.Join(dir, "github-notifier")
}

func Default() *Config {
	return &Config{
		PollIntervalMin:   5,
		ScheduleDays:      []string{"Monday", "Tuesday", "Wednesday", "Thursday"},
		ScheduleStartHour: 9,
		ScheduleEndHour:   17,
		SoundEnabled:      true,
		SoundPath:         "/System/Library/Sounds/Ping.aiff",
		AutoHideSeconds:   10,
		DNDHours:          2,
		PanelOpacity:      0.95,
	}
}

func saveTokenToKeychain(token string) error {
	cmd := exec.Command("security", "add-generic-password",
		"-a", "github-notifier",
		"-s", "github-notifier-token",
		"-w", token,
		"-U",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("keychain save failed: %w\n%s", err, out)
	}
	return nil
}

func getTokenFromKeychain() (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-a", "github-notifier",
		"-s", "github-notifier-token",
		"-w",
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keychain find failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func deleteTokenFromKeychain() error {
	cmd := exec.Command("security", "delete-generic-password",
		"-a", "github-notifier",
		"-s", "github-notifier-token",
	)
	return cmd.Run()
}

// Load reads config from file and loads the GitHub token from the macOS Keychain.
// If a token exists in the config file but not in Keychain (migration), it is
// automatically moved to Keychain on first load.
func Load() (*Config, error) {
	cfg := Default()
	path := filepath.Join(DefaultConfigDir(), "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("config: no config file at %s: %v", path, err)
	} else if err := json.Unmarshal(data, cfg); err != nil {
		log.Printf("config: error parsing config file: %v", err)
	}

	// Keychain token takes precedence over config file token
	token, keychainErr := getTokenFromKeychain()
	if keychainErr != nil {
		// No token in keychain yet. If one exists in the config file, migrate it.
		if cfg.GitHubToken != "" {
			log.Printf("config: migrating token from config file to Keychain")
			if migrateErr := saveTokenToKeychain(cfg.GitHubToken); migrateErr != nil {
				log.Printf("config: failed to migrate token to Keychain: %v", migrateErr)
			}
		}
	} else {
		cfg.GitHubToken = token
	}

	return cfg, nil
}

// Save writes config to disk without the token, and stores the token in the
// macOS Keychain. If the token is empty, any existing Keychain entry is removed.
func (c *Config) Save() error {
	if c.GitHubToken != "" {
		if err := saveTokenToKeychain(c.GitHubToken); err != nil {
			log.Printf("config: failed to save token to Keychain: %v", err)
		}
	} else {
		// Token cleared — remove from Keychain (ignore error if not present)
		_ = deleteTokenFromKeychain()
	}

	dir := DefaultConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write config to disk WITHOUT the token (it lives in Keychain)
	token := c.GitHubToken
	c.GitHubToken = ""
	defer func() { c.GitHubToken = token }()

	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
