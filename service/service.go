package service

import (
	"enque-learning/integration"
	"enque-learning/internal/config"
)

type Service struct {
	config       *config.Config
	integrations *integration.Integrations
}

func NewService(cfg *config.Config, integrations *integration.Integrations) *Service {
	return &Service{
		config:       cfg,
		integrations: integrations,
	}
}
