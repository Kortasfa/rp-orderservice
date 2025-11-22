package main

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func parseEnvs[T any]() (T, error) {
	var c T
	err := envconfig.Process(appID, &c)
	return c, errors.WithStack(err)
}

type Service struct {
	GracePeriod time.Duration `envconfig:"grace_period" default:"15s"`

	GRPCAddress string `envconfig:"grpc_address" default:":8081"`
	HTTPAddress string `envconfig:"http_address" default:":8082"`

	TemporalAddress string `envconfig:"temporal_address" default:"temporal:7233"`

	ProductServiceAddress      string `envconfig:"product_service_address" default:"product-service:8081"`
	PaymentServiceAddress      string `envconfig:"payment_service_address" default:"payment-service:8081"`
	NotificationServiceAddress string `envconfig:"notification_service_address" default:"notification-service:8081"`
}

type AMQP struct {
	User           string        `envconfig:"user" required:"true"`
	Password       string        `envconfig:"password" required:"true"`
	Host           string        `envconfig:"host" required:"true"`
	ConnectTimeout time.Duration `envconfig:"connect_timeout" default:"30s"`
}

type Database struct {
	User                  string        `envconfig:"user" required:"true"`
	Password              string        `envconfig:"password" required:"true"`
	Host                  string        `envconfig:"host" required:"true"`
	Name                  string        `envconfig:"name" required:"true"`
	MaxConnections        int           `envconfig:"max_connections" default:"20"`
	ConnectionMaxLifeTime time.Duration `envconfig:"connection_max_life_time" default:"10m"`
	ConnectionMaxIdleTime time.Duration `envconfig:"connection_max_idle_time" default:"1m"`
}
