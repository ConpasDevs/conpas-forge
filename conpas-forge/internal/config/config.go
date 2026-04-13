package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/conpasDEVS/conpas-forge/internal/models"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version      int               `yaml:"version"`
	Persona      string            `yaml:"persona"`
	Models       map[string]string `yaml:"models"`
	Modules      ModulesConfig     `yaml:"modules"`
	Engram       EngramConfig      `yaml:"engram"`
	SddStrictTDD string            `yaml:"sdd_strict_tdd,omitempty"` // "enabled" | "disabled" | "" (empty = auto-detect)
}

type ModulesConfig struct {
	Engram     ModuleStatus `yaml:"engram"`
	GentleAI   ModuleStatus `yaml:"gentle-ai"`
	ZohoDeluge ModuleStatus `yaml:"zoho-deluge"`
}

type ModuleStatus struct {
	Installed      bool   `yaml:"installed"`
	Version        string `yaml:"version,omitempty"`
	BinPath        string `yaml:"bin_path,omitempty"`
	SkillsDeployed int    `yaml:"skills_deployed,omitempty"`
}

type EngramConfig struct {
	DataDir string `yaml:"data_dir,omitempty"`
	Port    int    `yaml:"port,omitempty"`
}

func DefaultConfig() Config {
	modelDefaults := make(map[string]string, len(models.CanonicalRoles))
	for role, model := range models.Defaults {
		modelDefaults[role] = model
	}
	return Config{
		Version: 1,
		Persona: "asturiano",
		Models:  modelDefaults,
		Engram:  EngramConfig{Port: 7437},
	}
}

func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := DefaultConfig()
			return &cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config.yaml is corrupted: %w — delete %s to reset to defaults", err, ConfigPath())
	}

	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported config version %d (expected 1)", cfg.Version)
	}

	if cfg.Models == nil {
		cfg.Models = make(map[string]string)
	}
	for role, defaultModel := range models.Defaults {
		if cfg.Models[role] == "" {
			cfg.Models[role] = defaultModel
		}
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	// Refresh rollback backup before each write.
	if _, err := os.Stat(ConfigPath()); err == nil {
		existing, err := os.ReadFile(ConfigPath())
		if err != nil {
			return fmt.Errorf("read config for backup: %w", err)
		}
		if err := AtomicWrite(ConfigBak(), existing, 0644); err != nil {
			return fmt.Errorf("refresh config backup: %w", err)
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return AtomicWrite(ConfigPath(), data, 0644)
}

func Exists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}
