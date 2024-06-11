package configs

import (
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App           App
	Deployment    Deployment
	MonitorServer MonitorServer
}

type App struct {
	Version  string `envconfig:"APP_VERSION" default:"unknown"`
	GRPCPort string `envconfig:"APP_GRPC_PORT" default:"5000"`
	LogLevel string `envconfig:"APP_LOG_LEVEL" default:"INFO"`
}

type Deployment struct {
	Name string `envconfig:"DEPLOYMENT_NAME" default:"unknown"`
}

type MonitorServer struct {
	Port              string        `envconfig:"MONITOR_SERVER_PORT" default:"9090"`
	ReadHeaderTimeout time.Duration `envconfig:"MONITOR_SERVER_READ_HEADER_TIMEOUT" default:"15s"`
}

func ReadENV(cfg *Config) {
	err := godotenv.Load()
	if err != nil {
		envFileNotFound := strings.Contains(err.Error(), "no such file or directory")
		if !envFileNotFound {
			panic(err)
		}
	} else {
		log.Println("found env file")
	}
	err = envconfig.Process("", cfg)
	if err != nil {
		panic(err)
	}
}
