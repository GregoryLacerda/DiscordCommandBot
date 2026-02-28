package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Discord
	DiscordToken         string
	DiscordCommandPrefix string

	// RabbitMQ
	RabbitMQURL  string
	QueueName    string
	ExchangeName string
	RoutingKey   string

	// Server
	WebServerPort string
}

func LoadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Fatal Error loading .env")
	}

	return &Config{
		DiscordToken:         os.Getenv("DISCORD_TOKEN"),
		DiscordCommandPrefix: os.Getenv("DISCORD_COMMAND_PREFIX"),
		RabbitMQURL:          os.Getenv("RABBITMQ_URL"),
		QueueName:            os.Getenv("QUEUE_NAME"),
		ExchangeName:         os.Getenv("EXCHANGE_NAME"),
		RoutingKey:           os.Getenv("ROUTING_KEY"),
		WebServerPort:        os.Getenv("WEB_SERVER_PORT"),
	}
}
