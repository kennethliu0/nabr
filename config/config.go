package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Commands []Command `mapstructure:"commands"`
}

type Command struct {
	Name        string            `mapstructure:"name"`
	Description string            `mapstructure:"description"`
	Method      string            `mapstructure:"method"`
	URL         string            `mapstructure:"url"`
	Headers     map[string]string `mapstructure:"headers"`
	Body        string            `mapstructure:"body"`
	QueryParams map[string]string `mapstructure:"query_params"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
