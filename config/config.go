package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	HttpAddr string `env:"HTTP_ADDR"`

	MongoHost     string `env:"MONGO_HOST"`
	MongoPort     string `env:"MONGO_PORT"`
	MongoUsername string `env:"MONGO_USERNAME"`
	MongoPassword string `env:"MONGO_PASSWORD"`
	MongoDBName   string `env:"MONGO_DBNAME"`

	SMTPHost     string `env:"SMTP_HOST"`
	SMTPPort     string `env:"SMTP_PORT"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`

	MinioHost       string `env:"MINIO_HOST"`
	MinioPort       string `env:"MINIO_PORT"`
	MinioUsername   string `env:"MINIO_USERNAME"`
	MinioPassword   string `env:"MINIO_PASSWORD"`
	MinioBucketName string `env:"BUCKET_NAME"`
}

func New() (*Config, error) {
	cfg := &Config{}

	return cfg, cleanenv.ReadEnv(cfg)
}
