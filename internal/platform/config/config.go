package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	AppEnv   string
	HttpHost string
	HttpPort string
	LogLevel string

	PostgresUser     string
	PostgresPassword string
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	PostgresMaxConn  int
}

func EnvLoad() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env\n", err)
	}

	cfg := Config{}

	cfg.AppEnv = cast.ToString(getOrReturnDefault("APP_ENV", ""))
	cfg.HttpHost = cast.ToString(getOrReturnDefault("HTTP_HOST", ""))
	cfg.HttpPort = cast.ToString(getOrReturnDefault("HTTP_PORT", ""))
	cfg.LogLevel = cast.ToString(getOrReturnDefault("LOG_LEVEL", ""))

	cfg.PostgresUser = cast.ToString(getOrReturnDefault("POSTGRES_USER", ""))
	cfg.PostgresPassword = cast.ToString(getOrReturnDefault("POSTGRES_PASSWORD", ""))
	cfg.PostgresHost = cast.ToString(getOrReturnDefault("POSTGRES_HOST", "0.0.0.0"))
	cfg.PostgresPort = cast.ToString(getOrReturnDefault("POSTGRES_PORT", ""))
	cfg.PostgresDB = cast.ToString(getOrReturnDefault("POSTGRES_DB", ""))
	cfg.PostgresMaxConn = cast.ToInt(getOrReturnDefault("POSTGRES_MAX_CONNS", ""))

	return cfg
}

func getOrReturnDefault(key string, defaultValue interface{}) interface{} {
	_, exists := os.LookupEnv(key)
	if exists {
		return os.Getenv(key)
	}

	return defaultValue
}
