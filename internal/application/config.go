package application

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Configuration представляет конфигурацию приложения.
type Configuration struct {
	DevMode     bool   `env:"DEBAFR_DEV_MODE" envDefault:"false"`
	ProjectName string `env:"DEBAFR_PROJECT_NAME,required"`
	Filename    Filename
}

type Filename struct {
	// ComposeBlue содержит имя compose-файла для Blue.
	ComposeBlue string `env:"DEBAFR_FILENAME_COMPOSE_BLUE" envDefault:"compose.blue.yaml"`

	// ComposeGreen содержит имя compose-файла для Green.
	ComposeGreen string `env:"DEBAFR_FILENAME_COMPOSE_GREEN" envDefault:"compose.green.yaml"`

	// NginxConf содержит имя СИМВОЛИЧЕСКОЙ ССЫЛКИ на конфигурационный файл Nginx.
	NginxConf string `env:"DEBAFR_FILENAME_NGINX_CONF" envDefault:"nginx.conf"`
}

// LoadConfiguration возвращает новую конфигурацию приложения на основе переменных среды.
func LoadConfiguration() (*Configuration, error) {
	var config Configuration
	if err := env.Parse(&config); err != nil {
		return nil, fmt.Errorf("parse configuration: %v", err)
	}

	return &config, nil
}
