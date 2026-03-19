package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all user settings loaded from YAML.
type Config struct {
	// Threshold is legacy single-threshold support for older configs.
	Threshold float64        `yaml:"threshold"`
	Thresholds ThresholdsConfig `yaml:"thresholds"`
	Interval   int           `yaml:"interval"`
	Duration   int           `yaml:"duration"`
	Cooldown   int           `yaml:"cooldown"`
	Email      EmailConfig   `yaml:"email"`
}

// ThresholdsConfig defines separate limits for each metric.
type ThresholdsConfig struct {
	CPU    float64 `yaml:"cpu"`
	Memory float64 `yaml:"memory"`
}

// EmailConfig keeps SMTP settings used for sending alerts.
type EmailConfig struct {
	To       string `yaml:"to"`
	From     string `yaml:"from"`
	SMTP     string `yaml:"smtp"`
	Password string `yaml:"password"`
}

func (c Config) IntervalDuration() time.Duration {
	return time.Duration(c.Interval) * time.Second
}

func (c Config) DurationDuration() time.Duration {
	return time.Duration(c.Duration) * time.Second
}

func (c Config) CooldownDuration() time.Duration {
	return time.Duration(c.Cooldown) * time.Second
}

// Load reads and validates the YAML config file.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", err)
	}

	cfg.normalize()

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// normalize fills defaults and supports older config formats.
func (c *Config) normalize() {
	if c.Thresholds.CPU == 0 && c.Threshold > 0 {
		c.Thresholds.CPU = c.Threshold
	}
	if c.Thresholds.Memory == 0 {
		c.Thresholds.Memory = 75
	}
}

// Validate checks that required fields are set to safe values.
func (c Config) Validate() error {
	if c.Thresholds.CPU <= 0 || c.Thresholds.CPU > 100 {
		return fmt.Errorf("thresholds.cpu must be between 0 and 100")
	}
	if c.Thresholds.Memory <= 0 || c.Thresholds.Memory > 100 {
		return fmt.Errorf("thresholds.memory must be between 0 and 100")
	}
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be greater than 0")
	}
	if c.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}
	if c.Cooldown < 0 {
		return fmt.Errorf("cooldown must be 0 or greater")
	}
	if c.Email.To == "" {
		return fmt.Errorf("email.to is required")
	}
	if c.Email.From == "" {
		return fmt.Errorf("email.from is required")
	}
	if c.Email.SMTP == "" {
		return fmt.Errorf("email.smtp is required")
	}
	if c.Email.Password == "" {
		return fmt.Errorf("email.password is required")
	}

	return nil
}
