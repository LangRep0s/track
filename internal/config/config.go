package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	cfg  *Config
	once sync.Once
	mu   sync.Mutex
)

type Config struct {
	Global GlobalConfig     `json:"global"`
	Repos  map[string]*Repo `json:"repos"`
}

type GlobalConfig struct {
	DataDir          string   `json:"data_dir"`
	BackupCount      int      `json:"backup_count"`
	ExcludedPatterns []string `json:"excluded_patterns"`

	DefaultAssetPriority  []string `json:"default_asset_priority,omitempty"`  // e.g. ["x86_64", "amd64", "win64"]
	PreferredArchiveTypes []string `json:"preferred_archive_types,omitempty"` // e.g. [".zip", ".tar.gz"]
	DefaultPrerelease     bool     `json:"default_prerelease,omitempty"`
	DefaultAssetFilter    string   `json:"default_asset_filter,omitempty"`
	DefaultInstallName    string   `json:"default_install_name,omitempty"`
	MatcherMode           string   `json:"matcher_mode,omitempty"` // "strict" or "relaxed"
}

type Repo struct {
	Path              string   `json:"-"` 
	InstallName       string   `json:"install_name,omitempty"`
	AssetFilter       string   `json:"asset_filter,omitempty"`
	AssetExclude      string   `json:"asset_exclude,omitempty"`
	IncludePrerelease bool     `json:"include_prerelease"`
	CurrentVersion    string   `json:"current_version"`
	VersionHistory    []string `json:"version_history"`


	AssetPriority     []string `json:"asset_priority,omitempty"`
	PreferredArchives []string `json:"preferred_archives,omitempty"`
	FallbackArch      []string `json:"fallback_arch,omitempty"`
	FallbackOS        []string `json:"fallback_os,omitempty"`
	MatcherMode       string   `json:"matcher_mode,omitempty"` 
}

func Get() (*Config, error) {
	var err error
	once.Do(func() {
		cfg, err = loadConfig()
	})
	return cfg, err
}

func (c *Config) Save() error {
	mu.Lock()
	defer mu.Unlock()

	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func loadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultCfg := createDefaultConfig()
		if err := defaultCfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return defaultCfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &c, nil
}

func configPath() (string, error) {
	var trackDir string
	if isWindows() {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("LOCALAPPDATA not set")
		}
		trackDir = filepath.Join(localAppData, "track")
	} else {
		
		dataDir, err := dataPath()
		if err != nil {
			return "", err
		}
		trackDir = dataDir
	}
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(trackDir, "config.json"), nil
}

func isWindows() bool {
	return os.PathSeparator == '\\' 
}

func dataPath() (string, error) {
	dataDir, err := os.UserCacheDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".local", "share")
	}
	trackDir := filepath.Join(dataDir, "track")
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		return "", err
	}
	return trackDir, nil
}

func createDefaultConfig() *Config {
	dataPath, _ := dataPath()
	return &Config{
		Global: GlobalConfig{
			DataDir:          dataPath,
			BackupCount:      3,
			ExcludedPatterns: []string{"\\.deb$", "\\.rpm$", "checksums", "\\.sig$", "\\.asc$"},
		},
		Repos: make(map[string]*Repo),
	}
}
