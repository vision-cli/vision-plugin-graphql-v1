package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"log"
)

type Config struct {
	Host               string `envconfig:"HOST" default:"0.0.0.0"`
	Port               int    `envconfig:"PORT" default:"8080"`
	CORSAllowedOrigins string `envconfig:"CORS_ALLOWED_ORIGINS"`
	GraphiQLEnabled    bool   `envconfig:"GRAPHIQL_ENABLED" default:"false"`
	SchemaRoot         string `envconfig:"SCHEMA_ROOT" default:"../../"`
}

// Load attempts to read all config vars and return the struct or an error.
func Load() (Config, error) {
	var c Config
	if err := envconfig.Process("cmt", &c); err != nil {
		return Config{}, fmt.Errorf("failed to load config: %v", err)
	}
	return c, nil
}

// MustLoad will Load all config vars or cause a fatal exit.
func MustLoad() Config {
	config, err := Load()
	if err != nil {
		log.Fatalf("failed to load env config: %v", err)
	}
	return config
}
