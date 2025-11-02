package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type BaseConfig struct {
	BaseURL       string `default:"http://localhost:8080" envconfig:"BASE_URL"`
	Port          string `default:"8080" envconfig:"PORT"`
	RedisAddr     string `default:"localhost:6379" envconfig:"REDIS_ADDR"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`
}

type OAuthGoogleConfig struct {
	GoogleClientID     string `required:"true" envconfig:"OAUTH_GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `required:"true" envconfig:"OAUTH_GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURI  string `required:"true" envconfig:"OAUTH_GOOGLE_REDIRECT_URI"`
}

type ProxyConfig struct {
	Pattern   string
	TargetURL *url.URL
}

type Config struct {
	BaseConfig
	OAuthGoogleConfig
}

func NewConfig() (*Config, []*ProxyConfig, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, nil, err
	}

	if strings.HasSuffix(cfg.BaseURL, "/") {
		return nil, nil, fmt.Errorf("base url must not end with a slash: %s", cfg.BaseURL)
	}

	// Load proxy settings from config.yaml if exists
	type proxyConfig struct {
		Pattern   string `yaml:"pattern"`
		TargetURL string `yaml:"target_url"`
	}

	f, err := os.Open("config.yaml")
	if err != nil {
		return &cfg, nil, nil
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	proxies := struct {
		Proxies []*proxyConfig `yaml:"proxies"`
	}{}
	if err := d.Decode(&proxies); err != nil {
		return &cfg, nil, fmt.Errorf("failed to decode config.yaml: %w", err)
	}

	proxyConfigs := []*ProxyConfig{}
	for _, p := range proxies.Proxies {
		if p.TargetURL == "" || p.Pattern == "" {
			return nil, nil, fmt.Errorf("target url and pattern are required for proxy: %v", p)
		}
		if !strings.HasPrefix(p.TargetURL, "http") {
			return nil, nil, fmt.Errorf("target url must start with http(s): %v", p)
		}
		if strings.HasSuffix(p.TargetURL, "/") {
			return nil, nil, fmt.Errorf("target url must not end with a slash: %v", p)
		}
		if !strings.HasPrefix(p.Pattern, "/") {
			return nil, nil, fmt.Errorf("pattern must start with a slash: %v", p)
		}
		if strings.HasSuffix(p.Pattern, "/") {
			return nil, nil, fmt.Errorf("pattern must not end with a slash: %v", p)
		}
		url, err := url.Parse(p.TargetURL)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse target url: %w", err)
		}
		proxyConfigs = append(proxyConfigs, &ProxyConfig{
			Pattern:   p.Pattern,
			TargetURL: url,
		})
	}

	return &cfg, proxyConfigs, nil
}
