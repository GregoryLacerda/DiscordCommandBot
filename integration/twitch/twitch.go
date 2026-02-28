package twitch

import "enque-learning/internal/config"

type Twitch struct {
	Config *config.Config
}

func NewTwitchIntegration(config *config.Config) (*Twitch, error) {
	return &Twitch{
		Config: config,
	}, nil
}
