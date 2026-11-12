package domain

import "time"

type AppConfig struct {
	ProjectName string
	Images      []string
	Replacer    []AppConfigReplacer
	DockerLogin AppConfigDockerLogin
	Files       AppConfigFiles
	BinPaths    AppConfigBinPaths
	Timeouts    AppConfigTimeouts
	Healthcheck AppConfigHealthcheck
	Nginx       AppConfigNginx
}

type AppConfigReplacer struct {
	Path  string
	Blue  string
	Green string
}

type AppConfigNginx struct {
	Enabled bool
	Test    []string
	Reload  []string
}

type AppConfigDockerLogin struct {
	Enabled  bool
	Registry string
	Username string
	Password string
}

type AppConfigFiles struct {
	ComposeBlue  string
	ComposeGreen string
	NginxConf    string
}

type AppConfigBinPaths struct {
	Docker string
	Curl   string
	Nginx  string
}

type AppConfigTimeouts struct {
	Default time.Duration
}

type AppConfigHealthcheck struct {
	MaxRetries int
	RetryDelay time.Duration
}
