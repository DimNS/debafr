package domain

import "time"

type AppConfig struct {
	Files       FilesConfig
	BinPaths    BinPathsConfig
	Timeouts    TimeoutsConfig
	Healthcheck HealthcheckConfig
}

type FilesConfig struct {
	ComposeBlue  string
	ComposeGreen string
	NginxConf    string
}

type BinPathsConfig struct {
	Docker string
	Curl   string
	Nginx  string
}

type TimeoutsConfig struct {
	Default time.Duration
}

type HealthcheckConfig struct {
	MaxRetries int
	RetryDelay time.Duration
}
