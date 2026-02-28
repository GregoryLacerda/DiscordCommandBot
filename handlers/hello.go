package handlers

import (
	"enque-learning/events"
	"enque-learning/integration/discord"
	"fmt"
	"log"
)

type HelloCommandHandler struct {
	Discord *discord.Discord
}

func NewHelloCommandHandler(discord *discord.Discord) *HelloCommandHandler {
	return &HelloCommandHandler{
		Discord: discord,
	}
}

func (h *HelloCommandHandler) HandleEvent(event events.EventInterface) error {
	payload, ok := event.GetPayload().(discord.DiscordCommandPayload)
	if !ok {
		return fmt.Errorf("invalid payload for hello command")
	}

	log.Printf("handling hello command from user: %s", payload.Username)

	response := fmt.Sprintf("👋 Olá, %s! Como posso ajudar?", payload.Username)

	err := h.Discord.ReplyToMessage(payload.ChannelID, payload.MessageID, response)
	if err != nil {
		return fmt.Errorf("failed to send hello response: %w", err)
	}

	return nil
}
