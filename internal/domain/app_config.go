package domain

import "time"

type AppConfig struct {
	ProjectName     string
	ProxyPassPrefix string
	LocationPorts   []AppConfigLocationPort
	Files           AppConfigFiles
	BinPaths        AppConfigBinPaths
	Timeouts        AppConfigTimeouts
	Healthcheck     AppConfigHealthcheck
}

type AppConfigLocationPort struct {
	Location  string
	BluePort  string
	GreenPort string
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
