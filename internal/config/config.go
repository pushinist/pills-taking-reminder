package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env string `yaml:"env" env-required:"true"`
	HTTPServer
	GRPCServer
	DB
	NearTakingInterval time.Duration `yaml:"near_taking_interval" env-default:"60m"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println(".env file loaded")
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}
	log.Println("CONFIG_PATH environment is set")

	if _, err = os.Stat(configPath); err != nil {
		log.Fatalf("Config file %s does not exist", configPath)
	}
	log.Printf("Config file found: %s \n", configPath)

	var cfg Config

	if err = cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	log.Println("config has been read successfully!")

	return &cfg

}
