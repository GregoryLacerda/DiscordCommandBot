package handlers

import (
	"enque-learning/events"
	"enque-learning/integration/discord"
	"fmt"
	"log"
)

type UnknownCommandHandler struct {
	Discord *discord.Discord
}

func NewUnknownCommandHandler(discord *discord.Discord) *UnknownCommandHandler {
	return &UnknownCommandHandler{
		Discord: discord,
	}
}

func (h *UnknownCommandHandler) HandleEvent(event events.EventInterface) error {
	payload, ok := event.GetPayload().(discord.DiscordCommandPayload)
	if !ok {
		return fmt.Errorf("invalid payload for unknown command")
	}

	log.Printf("handling unknown command: %s from user: %s", payload.Command, payload.Username)

	response := fmt.Sprintf("❓ Comando desconhecido: `%s`\n\nUse `!help` para ver os comandos disponíveis.", payload.Command)

	err := h.Discord.ReplyToMessage(payload.ChannelID, payload.MessageID, response)
	if err != nil {
		return fmt.Errorf("failed to send unknown command response: %w", err)
	}

	return nil
}
