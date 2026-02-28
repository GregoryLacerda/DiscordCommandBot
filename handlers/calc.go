package handlers

import (
	"enque-learning/events"
	"enque-learning/integration/discord"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type CalcCommandHandler struct {
	Discord *discord.Discord
}

func NewCalcCommandHandler(discord *discord.Discord) *CalcCommandHandler {
	return &CalcCommandHandler{
		Discord: discord,
	}
}

func (h *CalcCommandHandler) HandleEvent(event events.EventInterface) error {
	payload, ok := event.GetPayload().(discord.DiscordCommandPayload)
	if !ok {
		return fmt.Errorf("invalid payload for calc command")
	}

	log.Printf("handling calc command from user: %s", payload.Username)

	// Validar argumentos
	if len(payload.Arguments) == 0 {
		response := "❌ Uso: `!calc <expressão>`\n**Exemplo:** !calc 2 + 2"
		return h.sendResponse(payload, response)
	}

	expression := strings.Join(payload.Arguments, " ")

	// Calcular
	result, err := h.evaluate(expression)
	if err != nil {
		response := fmt.Sprintf("❌ Erro ao calcular: %s", err.Error())
		return h.sendResponse(payload, response)
	}

	response := fmt.Sprintf("🧮 **Resultado:**\n`%s = %.2f`", expression, result)
	return h.sendResponse(payload, response)
}

func (h *CalcCommandHandler) evaluate(expr string) (float64, error) {
	parts := strings.Fields(expr)

	if len(parts) != 3 {
		return 0, fmt.Errorf("formato esperado: número operador número (ex: 2 + 2)")
	}

	num1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("primeiro número inválido: %s", parts[0])
	}

	num2, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, fmt.Errorf("segundo número inválido: %s", parts[2])
	}

	operator := parts[1]

	switch operator {
	case "+":
		return num1 + num2, nil
	case "-":
		return num1 - num2, nil
	case "*", "x", "×":
		return num1 * num2, nil
	case "/", "÷":
		if num2 == 0 {
			return 0, fmt.Errorf("divisão por zero")
		}
		return num1 / num2, nil
	case "^", "**":
		result := 1.0
		for i := 0; i < int(num2); i++ {
			result *= num1
		}
		return result, nil
	default:
		return 0, fmt.Errorf("operador inválido: %s (use: +, -, *, /, ^)", operator)
	}
}

func (h *CalcCommandHandler) sendResponse(payload discord.DiscordCommandPayload, response string) error {
	err := h.Discord.ReplyToMessage(payload.ChannelID, payload.MessageID, response)
	if err != nil {
		return fmt.Errorf("failed to send calc response: %w", err)
	}
	return nil
}
