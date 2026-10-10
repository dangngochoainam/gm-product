package configs

import (
	"fmt"
	"time"

	"github.com/dangngochoainam/gopkg/confighelper"
)

// yaml format
var defaultConfig = []byte(`
env: dev
grpc_port: 50060
admin_port: 50061
shutdown_drain_delay: 5s
shutdown_timeout: 25s
database_postgres:
  host: localhost
  port: 5434
  database: ecommerce
  schema: user
  username: root
  password: Abc12345
  logging_enabled: true
http_client:
  timeout: 30s
  max_idle_conns: 100
  max_idle_conns_per_host: 10
  max_conns_per_host: 100
  max_retries: 3
  retry_initial_delay: 100ms
  retry_max_delay: 10s
  retry_multiplier: 2.0
  max_response_body_size: 10485760

`)

type (
	Config struct {
		Env      string `mapstructure:"env"`
		GrpcPort string `mapstructure:"grpc_port"`
		// AdminPort carries probes and metrics only. It is never an ingress
		// target: business traffic reaches this service over gRPC, through
		// gm-gateway. See docs/service-ports.md.
		AdminPort string `mapstructure:"admin_port"`
		// ShutdownDrainDelay is how long the process keeps serving after
		// readiness fails, so Kubernetes can drop it from the endpoint list
		// before the listener closes.
		ShutdownDrainDelay time.Duration     `mapstructure:"shutdown_drain_delay"`
		ShutdownTimeout    time.Duration     `mapstructure:"shutdown_timeout"`
		DatabasePostgres   SqlDatabaseConfig `mapstructure:"database_postgres"`
		HttpClientConfig   HttpClientConfig  `mapstructure:"http_client"`
	}

	SqlDatabaseConfig struct {
		Host           string `mapstructure:"host"`
		Port           uint32 `mapstructure:"port"`
		Database       string `mapstructure:"database"`
		Schema         string `mapstructure:"schema"`
		Username       string `mapstructure:"username"`
		Password       string `mapstructure:"password"`
		LoggingEnabled bool   `mapstructure:"logging_enabled"`
		AutoMigration  bool   `mapstructure:"auto_migration"`
	}

	HttpClientConfig struct {
		Timeout             time.Duration `mapstructure:"timeout"`
		MaxIdleConns        int           `mapstructure:"max_idle_conns"`
		MaxIdleConnsPerHost int           `mapstructure:"max_idle_conns_per_host"`
		MaxConnsPerHost     int           `mapstructure:"max_conns_per_host"`
		MaxRetries          int           `mapstructure:"max_retries"`
		RetryInitialDelay   time.Duration `mapstructure:"retry_initial_delay"`
		RetryMaxDelay       time.Duration `mapstructure:"retry_max_delay"`
		RetryMultiplier     float64       `mapstructure:"retry_multiplier"`
		MaxResponseBodySize int64         `mapstructure:"max_response_body_size"`
	}
)

func LoadConfig() (*Config, error) {
	configMap := &Config{}

	err := confighelper.Load(configMap, defaultConfig)
	if err != nil {
		fmt.Printf("Error while load config, err: %v", err)
		return nil, err
	}

	return configMap, nil
}
