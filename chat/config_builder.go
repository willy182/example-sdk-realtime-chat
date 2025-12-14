package chat

import (
	"errors"
	"time"
)

type ConfigBuilder struct {
	cfg Config
}

func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		cfg: Config{
			DialTimeout: 5 * time.Second,
		},
	}
}

func (b *ConfigBuilder) WithAPIKey(v string) *ConfigBuilder {
	b.cfg.APIKey = v
	return b
}

func (b *ConfigBuilder) WithEndpoint(v string) *ConfigBuilder {
	b.cfg.Endpoint = v
	return b
}

func (b *ConfigBuilder) WithDialTimeout(d time.Duration) *ConfigBuilder {
	b.cfg.DialTimeout = d
	return b
}

func (b *ConfigBuilder) WithRetry(r RetryStrategy) *ConfigBuilder {
	b.cfg.Retry = r
	return b
}

func (b *ConfigBuilder) Build() (Config, error) {
	if b.cfg.APIKey == "" {
		return Config{}, errors.New("api key required")
	}
	if b.cfg.Endpoint == "" {
		return Config{}, errors.New("endpoint required")
	}
	return b.cfg, nil
}
