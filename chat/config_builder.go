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

func (b *ConfigBuilder) WithAPIKey(key string) *ConfigBuilder {
    b.cfg.APIKey = key
    return b
}

func (b *ConfigBuilder) WithEndpoint(ep string) *ConfigBuilder {
    b.cfg.Endpoint = ep
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
