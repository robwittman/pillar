package config

import (
	"os"
	"strconv"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

type PluginConfig struct {
	Name    string            `yaml:"name"`
	Source  string            `yaml:"source"`  // e.g. "github.com/robwittman/pillar-plugin-keycloak"
	Version string            `yaml:"version"` // e.g. "1.0.0" or "latest"
	Path    string            `yaml:"path"`    // local binary path (overrides source)
	Config  map[string]string `yaml:"config"`
}

type PluginSettings struct {
	CacheDir string `yaml:"cache_dir" env:"PILLAR_PLUGIN_CACHE_DIR"`
}

type Config struct {
	HTTPAddr    string `env:"PILLAR_HTTP_ADDR" envDefault:":8080" yaml:"http_addr"`
	GRPCAddr    string `env:"PILLAR_GRPC_ADDR" envDefault:":9090" yaml:"grpc_addr"`
	PostgresURL string `env:"PILLAR_POSTGRES_URL" envDefault:"postgres://pillar:pillar@localhost:5432/pillar?sslmode=disable" yaml:"postgres_url"`
	RedisAddr   string `env:"PILLAR_REDIS_ADDR" envDefault:"localhost:6379" yaml:"redis_addr"`
	LogLevel    string `env:"PILLAR_LOG_LEVEL" envDefault:"info" yaml:"log_level"`

	KubeEnabled      bool   `env:"PILLAR_KUBE_ENABLED" envDefault:"false" yaml:"kube_enabled"`
	KubeContext      string `env:"PILLAR_KUBE_CONTEXT" yaml:"kube_context"`
	KubeNamespace    string `env:"PILLAR_KUBE_NAMESPACE" envDefault:"default" yaml:"kube_namespace"`
	AgentImage       string `env:"PILLAR_AGENT_IMAGE" envDefault:"pillar-agent:latest" yaml:"agent_image"`
	GRPCExternalAddr string `env:"PILLAR_GRPC_EXTERNAL_ADDR" envDefault:"host.docker.internal:9090" yaml:"grpc_external_addr"`

	PluginSettings PluginSettings `yaml:"plugin_settings"`
	Plugins        []PluginConfig `yaml:"plugins"`
}

func defaultConfig() *Config {
	return &Config{
		HTTPAddr:         ":8080",
		GRPCAddr:         ":9090",
		PostgresURL:      "postgres://pillar:pillar@localhost:5432/pillar?sslmode=disable",
		RedisAddr:        "localhost:6379",
		LogLevel:         "info",
		KubeNamespace:    "default",
		AgentImage:       "pillar-agent:latest",
		GRPCExternalAddr: "host.docker.internal:9090",
	}
}

func Load() (*Config, error) {
	configFile := os.Getenv("PILLAR_CONFIG_FILE")

	if configFile == "" {
		// No config file — standard env-only loading with defaults
		cfg := &Config{}
		if err := env.Parse(cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	// Load from YAML file — start with defaults since envDefault tags
	// are not applied when loading from YAML.
	cfg := defaultConfig()
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply explicitly set env vars (override YAML values)
	overrideFromEnv(cfg)

	return cfg, nil
}

// overrideFromEnv applies only env vars that are explicitly set.
func overrideFromEnv(cfg *Config) {
	if v, ok := os.LookupEnv("PILLAR_HTTP_ADDR"); ok {
		cfg.HTTPAddr = v
	}
	if v, ok := os.LookupEnv("PILLAR_GRPC_ADDR"); ok {
		cfg.GRPCAddr = v
	}
	if v, ok := os.LookupEnv("PILLAR_POSTGRES_URL"); ok {
		cfg.PostgresURL = v
	}
	if v, ok := os.LookupEnv("PILLAR_REDIS_ADDR"); ok {
		cfg.RedisAddr = v
	}
	if v, ok := os.LookupEnv("PILLAR_LOG_LEVEL"); ok {
		cfg.LogLevel = v
	}
	if v, ok := os.LookupEnv("PILLAR_KUBE_ENABLED"); ok {
		cfg.KubeEnabled, _ = strconv.ParseBool(v)
	}
	if v, ok := os.LookupEnv("PILLAR_KUBE_CONTEXT"); ok {
		cfg.KubeContext = v
	}
	if v, ok := os.LookupEnv("PILLAR_KUBE_NAMESPACE"); ok {
		cfg.KubeNamespace = v
	}
	if v, ok := os.LookupEnv("PILLAR_AGENT_IMAGE"); ok {
		cfg.AgentImage = v
	}
	if v, ok := os.LookupEnv("PILLAR_GRPC_EXTERNAL_ADDR"); ok {
		cfg.GRPCExternalAddr = v
	}
	if v, ok := os.LookupEnv("PILLAR_PLUGIN_CACHE_DIR"); ok {
		cfg.PluginSettings.CacheDir = v
	}
}
