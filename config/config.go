package config

import "github.com/spf13/viper"

type AppConfig struct {
	ProjectID      string
	Credentials    string
	DefaultRegion  string
}

func LoadAppConfig() AppConfig {
	return AppConfig{
		ProjectID:     viper.GetString("project_id"),
		Credentials:   viper.GetString("credentials"),
		DefaultRegion: viper.GetString("default_region"),
	}
}
