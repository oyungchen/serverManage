package config

import (
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Web      WebConfig      `yaml:"web"`
	Storage  StorageConfig  `yaml:"storage"`
	Discover DiscoverConfig `yaml:"discover"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type WebConfig struct {
	Static string `yaml:"static"`
}

type StorageConfig struct {
	DataFile string `yaml:"dataFile"`
	LogDir   string `yaml:"logDir"`
}

type DiscoverConfig struct {
	ScanDirs   []string `yaml:"scanDirs"`
	ExcludeDirs []string `yaml:"excludeDirs"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 展开 ~ 为用户目录
	homeDir := getHomeDir()
	cfg.Storage.DataFile = expandHome(cfg.Storage.DataFile, homeDir)
	cfg.Storage.LogDir = expandHome(cfg.Storage.LogDir, homeDir)

	return &cfg, nil
}

func getHomeDir() string {
	user, _ := user.Current()
	if user.HomeDir != "" {
		return user.HomeDir
	}
	return os.Getenv("HOME")
}

func expandHome(path, home string) string {
	if len(path) > 0 && path[0] == '~' {
		return filepath.Join(home, path[1:])
	}
	return path
}