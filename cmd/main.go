package main

import (
	"context"
	"enque-learning/events"
	"enque-learning/integration"
	"enque-learning/internal/config"
	"enque-learning/server"
	"enque-learning/service"
)

func main() {

	ctx := context.Background()

	cfg := config.LoadConfig()

	dispatcher := events.NewEventDispatcher()

	integrations, err := integration.NewIntegrations(ctx, cfg, dispatcher)
	if err != nil {
		endAsError(err)
	}

	srv := service.NewService(cfg, integrations)

	server := server.NewServer(cfg, dispatcher, integrations, srv)

	if err := server.StartAll(); err != nil {
		endAsError(err)
	}
}

func endAsError(err error) {
	panic(err)
}
