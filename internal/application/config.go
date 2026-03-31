package application

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pelletier/go-toml/v2"

	"debafr/internal/domain"
)

const (
	defaultCmdTimeout = 30 * time.Second
	defaultRetryDelay = 3 * time.Second
	defaultMaxRetries = 10
)

type Configuration struct {
	DevMode bool `env:"DEBAFR_DEV_MODE" envDefault:"false"`

	Toml TomlConfig
}

type TomlConfig struct {
	App         AppConfig         `toml:"app"`
	Files       FilesConfig       `toml:"files"`
	BinPaths    BinPathsConfig    `toml:"binpaths"`
	Timeouts    TimeoutsConfig    `toml:"timeouts"`
	Healthcheck HealthcheckConfig `toml:"healthcheck"`
}

type AppConfig struct {
	ProjectName string `toml:"project_name"`
}

type FilesConfig struct {
	ComposeBlue  string `toml:"compose_blue"`
	ComposeGreen string `toml:"compose_green"`
	NginxConf    string `toml:"nginx_conf"`
}

type BinPathsConfig struct {
	Docker string `toml:"docker"`
	Curl   string `toml:"curl"`
	Nginx  string `toml:"nginx"`
}

type TimeoutsConfig struct {
	Default Duration `toml:"default"`
}

type HealthcheckConfig struct {
	MaxRetries int      `toml:"max_retries"`
	RetryDelay Duration `toml:"retry_delay"`
}

type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	v, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}
	*d = Duration(v)
	return nil
}

func (tc *TomlConfig) Validate() error {
	if tc.App.ProjectName == "" {
		return errors.New("app name is empty")
	}

	return nil
}

func (tc *TomlConfig) GetDomainConfig() domain.AppConfig {
	return domain.AppConfig{
		Files: domain.FilesConfig{
			ComposeBlue:  tc.Files.ComposeBlue,
			ComposeGreen: tc.Files.ComposeGreen,
			NginxConf:    tc.Files.NginxConf,
		},
		BinPaths: domain.BinPathsConfig{
			Docker: tc.BinPaths.Docker,
			Curl:   tc.BinPaths.Curl,
			Nginx:  tc.BinPaths.Nginx,
		},
		Timeouts: domain.TimeoutsConfig{
			Default: time.Duration(tc.Timeouts.Default),
		},
		Healthcheck: domain.HealthcheckConfig{
			MaxRetries: tc.Healthcheck.MaxRetries,
			RetryDelay: time.Duration(tc.Healthcheck.RetryDelay),
		},
	}
}

func LoadConfiguration(tomlName string) (*Configuration, error) {
	var config Configuration
	if err := env.Parse(&config); err != nil {
		return nil, fmt.Errorf("parse configuration: %v", err)
	}

	tomlConfig, err := loadTomlConfig(tomlName)
	if err != nil {
		return nil, fmt.Errorf("load toml config: %v", err)
	}
	config.Toml = tomlConfig

	return &config, nil
}

func loadTomlConfig(name string) (TomlConfig, error) {
	var tomlConfig TomlConfig

	wdPath, err := os.Getwd()
	if err != nil {
		return tomlConfig, fmt.Errorf("get current directory: %w", err)
	}

	root, err := os.OpenRoot(wdPath)
	if err != nil {
		return tomlConfig, fmt.Errorf("open root: %w", err)
	}
	defer root.Close()

	data, err := root.ReadFile(name)
	if err != nil {
		if os.IsNotExist(err) {
			SetDefaults(&tomlConfig)
			return tomlConfig, nil
		}
		return tomlConfig, fmt.Errorf("read toml file: %w", err)
	}

	err = toml.Unmarshal(data, &tomlConfig)
	if err != nil {
		return tomlConfig, fmt.Errorf("unmarshal toml: %w", err)
	}

	SetDefaults(&tomlConfig)

	if err := tomlConfig.Validate(); err != nil {
		return tomlConfig, fmt.Errorf("validate toml: %w", err)
	}

	return tomlConfig, nil
}

func SetDefaults(cfg *TomlConfig) {
	if cfg.BinPaths.Docker == "" {
		cfg.BinPaths.Docker = "/usr/bin/docker"
	}
	if cfg.BinPaths.Curl == "" {
		cfg.BinPaths.Curl = "/usr/bin/curl"
	}
	if cfg.BinPaths.Nginx == "" {
		cfg.BinPaths.Nginx = "/usr/sbin/nginx"
	}
	if cfg.Timeouts.Default == 0 {
		cfg.Timeouts.Default = Duration(defaultCmdTimeout)
	}
	if cfg.Healthcheck.MaxRetries == 0 {
		cfg.Healthcheck.MaxRetries = defaultMaxRetries
	}
	if cfg.Healthcheck.RetryDelay == 0 {
		cfg.Healthcheck.RetryDelay = Duration(defaultRetryDelay)
	}
	if cfg.Files.ComposeBlue == "" {
		cfg.Files.ComposeBlue = "compose.blue.yaml"
	}
	if cfg.Files.ComposeGreen == "" {
		cfg.Files.ComposeGreen = "compose.green.yaml"
	}
	if cfg.Files.NginxConf == "" {
		cfg.Files.NginxConf = "nginx.conf"
	}
}
