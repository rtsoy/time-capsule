package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	MongoHost     string `env:"MONGO_HOST"`
	MongoPort     string `env:"MONGO_PORT"`
	MongoUsername string `env:"MONGO_USERNAME"`
	MongoPassword string `env:"MONGO_PASSWORD"`
	MongoDBName   string `env:"MONGO_DBNAME"`
}

func New() (*Config, error) {
	cfg := &Config{}

	return cfg, cleanenv.ReadEnv(cfg)
}
