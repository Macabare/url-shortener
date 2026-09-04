package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-required:"true"`
	StoragePath string `yaml:"storage_path" env-required:"true"`

	DBHost     string `env:"DB_HOST"`
	DBPort     string `env:"DB_PORT"`
	DBName     string `env:"DB_NAME"`
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASSWORD"`
	DBSSLMode  string `env:"DB_SSLMODE"`

	HttpServer `yaml:"http_server"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set.")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s.", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config file: %s", err)
	}

	return &cfg
}
