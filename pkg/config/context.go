package config

import (
	"context"
)

type contextKey uint

const (
	configKey contextKey = iota
	globalConfigKey
)

func WithConfig(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, configKey, config)
}

func FromContext(ctx context.Context) *Config {
	return ctx.Value(configKey).(*Config) //nolint:forcetypeassert
}

func WithGlobalConfig(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, globalConfigKey, config)
}

func GlobalConfigFromContext(ctx context.Context) *Config {
	return ctx.Value(globalConfigKey).(*Config) //nolint:forcetypeassert
}
