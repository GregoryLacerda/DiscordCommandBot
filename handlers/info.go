package handlers

import (
	"enque-learning/events"
	"enque-learning/integration/discord"
	"fmt"
	"log"
)

type InfoCommandHandler struct {
	Discord *discord.Discord
}

func NewInfoCommandHandler(discord *discord.Discord) *InfoCommandHandler {
	return &InfoCommandHandler{
		Discord: discord,
	}
}

func (h *InfoCommandHandler) HandleEvent(event events.EventInterface) error {
	payload, ok := event.GetPayload().(discord.DiscordCommandPayload)
	if !ok {
		return fmt.Errorf("invalid payload for info command")
	}

	log.Printf("handling info command from user: %s", payload.Username)

	response := fmt.Sprintf(`ℹ️ **Informações do Sistema:**

👤 **Usuário:** %s
🆔 **User ID:** %s
📝 **Comando:** %s
📅 **Canal ID:** %s
🏢 **Guild ID:** %s
⏰ **Timestamp:** %s`,
		payload.Username,
		payload.UserID,
		payload.Command,
		payload.ChannelID,
		payload.GuildID,
		payload.Timestamp,
	)

	err := h.Discord.ReplyToMessage(payload.ChannelID, payload.MessageID, response)
	if err != nil {
		return fmt.Errorf("failed to send info response: %w", err)
	}

	return nil
}
