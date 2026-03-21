package config

import (
	"os"
	"regexp"

	"github.com/spf13/viper"
)

var envRe = regexp.MustCompile(`\$\{(\w+)\}`)

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
	Output      string            `mapstructure:"output"`
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

	cfg.expandEnv()
	return &cfg, nil
}

func expandString(s string) string {
	return envRe.ReplaceAllStringFunc(s, func(match string) string {
		name := envRe.FindStringSubmatch(match)[1]
		if val, ok := os.LookupEnv(name); ok {
			return val
		}
		return match
	})
}

func expandMap(m map[string]string) {
	for k, v := range m {
		m[k] = expandString(v)
	}
}

func (cfg *Config) expandEnv() {
	for i := range cfg.Commands {
		c := &cfg.Commands[i]
		c.URL = expandString(c.URL)
		c.Body = expandString(c.Body)
		c.Output = expandString(c.Output)
		expandMap(c.Headers)
		expandMap(c.QueryParams)
	}
}
