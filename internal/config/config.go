package config

import "github.com/caarlos0/env/v11"

type Config struct {
	HTTPAddr    string `env:"PILLAR_HTTP_ADDR" envDefault:":8080"`
	GRPCAddr    string `env:"PILLAR_GRPC_ADDR" envDefault:":9090"`
	PostgresURL string `env:"PILLAR_POSTGRES_URL" envDefault:"postgres://pillar:pillar@localhost:5432/pillar?sslmode=disable"`
	RedisAddr   string `env:"PILLAR_REDIS_ADDR" envDefault:"localhost:6379"`
	LogLevel    string `env:"PILLAR_LOG_LEVEL" envDefault:"info"`

	KubeEnabled      bool   `env:"PILLAR_KUBE_ENABLED" envDefault:"false"`
	KubeNamespace    string `env:"PILLAR_KUBE_NAMESPACE" envDefault:"default"`
	AgentImage       string `env:"PILLAR_AGENT_IMAGE" envDefault:"pillar-agent:latest"`
	GRPCExternalAddr string `env:"PILLAR_GRPC_EXTERNAL_ADDR" envDefault:"host.docker.internal:9090"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
