package handlers

import (
	"encoding/json"
	"enque-learning/events"
	"enque-learning/integration/discord"
	"enque-learning/integration/rabbitmq"
	"fmt"
	"log"
)

type CommandHandler struct {
	RabbitMQ *rabbitmq.RabbitMQ
	Discord  *discord.Discord
}

func NewCommandHandler(rabbitMQ *rabbitmq.RabbitMQ, discord *discord.Discord) *CommandHandler {
	return &CommandHandler{
		RabbitMQ: rabbitMQ,
		Discord:  discord,
	}
}

func (h *CommandHandler) HandleEvent(event events.EventInterface) error {
	payload, ok := event.GetPayload().(discord.DiscordCommandPayload)
	if !ok {
		return fmt.Errorf("invalid payload")
	}

	log.Printf("processing command: %s", payload.Command)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = h.RabbitMQ.Publisher(jsonData)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	//h.Discord.SendMessage(payload.ChannelID, fmt.Sprintf("⏳ Comando `%s` recebido e está sendo processado...", payload.Command))

	return nil
}
