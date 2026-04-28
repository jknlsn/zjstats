package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config drives how zjstat formats the status line.
type Config struct {
	Theme   Theme            `toml:"theme"`
	Metrics []MetricConfig   `toml:"metrics"`
	Context ContextConfig    `toml:"context"`
}

// Theme holds zjstatus color directives.
type Theme struct {
	Background string `toml:"background"`
	Label      string `toml:"label"`
	Text       string `toml:"text"`
	Alert      string `toml:"alert"`
	Warn       string `toml:"warn"`
}

// MetricConfig describes one metric slot in the status line.
type MetricConfig struct {
	Name          string  `toml:"name"`
	Label         string  `toml:"label"`
	Format        string  `toml:"format"`
	Mount         string  `toml:"mount,omitempty"`          // for disk metrics
	HideIfMissing bool    `toml:"hide_if_missing,omitempty"` // skip if disk not mounted / gpu unavailable
	WarnAt        float64 `toml:"warn,omitempty"`           // threshold to switch label to warn color
	AlertAt       float64 `toml:"alert,omitempty"`          // threshold to switch label to alert color
}

// ContextConfig holds formatting rules for the four context states.
type ContextConfig struct {
	SSH     ContextRule `toml:"ssh"`
	SSHUser ContextRule `toml:"ssh_user"`
	User    ContextRule `toml:"user"`
	Local   ContextRule `toml:"local"`
}

// ContextRule is one context state formatter.
type ContextRule struct {
	Format string `toml:"format"`
	Color  string `toml:"color"`
}

// Default returns the built-in configuration that matches the original layout.
func Default() *Config {
	return &Config{
		Theme: Theme{
			Background: "$surface0",
			Label:      "$blue,bold",
			Text:       "$text",
			Alert:      "$red,bold",
			Warn:       "$yellow,bold",
		},
		Metrics: []MetricConfig{
			{Name: "cpu",    Label: "cpu", Format: "%2.0f%%"},
			{Name: "gpu",    Label: "gpu", Format: "%2.0f%%", HideIfMissing: true},
			{Name: "memory", Label: "mem", Format: "%2.0f%%"},
			{Name: "disk",   Label: "ssd", Format: "%2.0f%%", Mount: "/"},
			{Name: "disk",   Label: "ext", Format: "%2.0f%%", Mount: "/Volumes/OWC", HideIfMissing: true},
		},
		Context: ContextConfig{
			SSH:     ContextRule{Format: "@{hostname}", Color: "$yellow,bold"},
			SSHUser: ContextRule{Format: "{user}@{hostname}", Color: "$red,bold"},
			User:    ContextRule{Format: "{user}@local", Color: "$red,bold"},
			Local:   ContextRule{Format: "@local", Color: "$blue,bold"},
		},
	}
}

// Load reads a TOML file from the standard config path or returns defaults.
func Load() (*Config, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return Default(), nil
	}
	path := filepath.Join(cfgDir, "zjstat", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}
