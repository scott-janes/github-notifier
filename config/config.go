package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	GitHubToken       string   `json:"github_token"`
	GitHubUsername    string   `json:"github_username"`
	PollIntervalMin   int      `json:"poll_interval_minutes"`
	ScheduleDays      []string `json:"schedule_days"`
	ScheduleStartHour int      `json:"schedule_start_hour"`
	ScheduleEndHour   int      `json:"schedule_end_hour"`
	SoundEnabled      bool     `json:"sound_enabled"`
	AutoHideSeconds   int      `json:"auto_hide_seconds"`
	DisableDrag       bool     `json:"disable_drag"`
	DNDEnabled        bool     `json:"dnd_enabled"`
	DNDHours          float64  `json:"dnd_hours"`
	PanelOpacity      float64  `json:"panel_opacity"`
	WindowX           int      `json:"window_x"`
	WindowY           int      `json:"window_y"`
}

func DefaultConfigDir() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "github-notifications")
}

func Default() *Config {
	return &Config{
		PollIntervalMin:   5,
		ScheduleDays:      []string{"Monday", "Tuesday", "Wednesday", "Thursday"},
		ScheduleStartHour: 9,
		ScheduleEndHour:   17,
		SoundEnabled:      true,
		AutoHideSeconds:   10,
		DNDHours:          2,
		PanelOpacity:      0.95,
	}
}

func Load() (*Config, error) {
	cfg := Default()
	path := filepath.Join(DefaultConfigDir(), "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil
	}
	err = json.Unmarshal(data, cfg)
	return cfg, err
}

func (c *Config) Save() error {
	dir := DefaultConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
