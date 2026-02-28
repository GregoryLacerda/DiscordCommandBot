# Melhoria: Command Handlers usando Sistema de Eventos Existente

## 🎯 Objetivo

Aproveitar o **EventDispatcher** e **EventHandlerInterface** já implementados para criar handlers específicos para cada comando do Discord.

---

## 📊 Arquitetura Atual

Você já tem:
- ✅ `EventInterface` - Interface para eventos
- ✅ `EventHandlerInterface` - Interface para handlers
- ✅ `EventDispatcher` - Registry que mapeia eventos para handlers
- ✅ `Event` - Implementação de eventos

---

## 🔄 Novo Fluxo

### Atual:
```
Discord → Event(discord.command.received) → CommandHandler → RabbitMQ
RabbitMQ → ResponseHandler.processCommand() → switch/case → Discord
```

### Novo:
```
Discord → Event(discord.command.received) → CommandHandler → RabbitMQ
RabbitMQ → Event(discord.command.{nome}) → Handler específico → Discord

Eventos:
- discord.command.ping → PingCommandHandler
- discord.command.help → HelpCommandHandler
- discord.command.calc → CalcCommandHandler
```

---

## 🏗️ Estrutura de Arquivos

```
service/
├── handlers.go                    # Handlers principais (já existe)
└── command_handlers/
    ├── ping_handler.go            # 📝 Novo
    ├── help_handler.go            # 📝 Novo
    ├── hello_handler.go           # 📝 Novo
    ├── info_handler.go            # 📝 Novo
    ├── calc_handler.go            # 📝 Novo
    └── unknown_handler.go         # 📝 Novo
```

---

## Passo 1: Criar Command Handlers

Cada comando é um **EventHandlerInterface** independente.

### 1.1 - Criar `service/command_handlers/ping_handler.go`

```go
package command_handlers

import (
	"enque-learning/events"
	"enque-learning/integration/discord"
	"fmt"
	"log"
)

type PingCommandHandler struct {
	Discord *discord.Discord
}

func NewPingCommandHandler(discord *discord.Discord) *PingCommandHandler {
	return &PingCommandHandler{
		Discord: discord,
	}
}

// HandleEvent implementa EventHandlerInterface
func (h *PingCommandHandler) HandleEvent(event events.EventInterface) error {
	payload, ok := event.GetPayload().(discord.DiscordCommandPayload)
	if !ok {
		return fmt.Errorf("invalid payload for ping command")
	}

	log.Printf("handling ping command from user: %s", payload.Username)

	response := "🏓 Pong!"
	
	err := h.Discord.ReplyToMessage(payload.ChannelID, payload.MessageID, response)
	if err != nil {
		return fmt.Errorf("failed to send ping response: %w", err)
	}

	return nil
}
```

---

### 1.2 - Criar `service/command_handlers/hello_handler.go`

```go
package command_handlers

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
```

---

### 1.3 - Criar `service/command_handlers/help_handler.go`

```go
package command_handlers

import (
	"enque-learning/events"
	"enque-learning/integration/discord"
	"fmt"
	"log"
)

type HelpCommandHandler struct {
	Discord *discord.Discord
}

func NewHelpCommandHandler(discord *discord.Discord) *HelpCommandHandler {
	return &HelpCommandHandler{
		Discord: discord,
	}
}

func (h *HelpCommandHandler) HandleEvent(event events.EventInterface) error {
	payload, ok := event.GetPayload().(discord.DiscordCommandPayload)
	if !ok {
		return fmt.Errorf("invalid payload for help command")
	}

	log.Printf("handling help command from user: %s", payload.Username)

	response := `📚 **Comandos Disponíveis:**

**!ping** - Testa se o bot está respondendo
**!hello** / **!oi** - Recebe uma saudação
**!calc <expressão>** - Calcula uma expressão matemática (ex: !calc 2 + 2)
**!info** - Mostra informações sobre você e o sistema
**!help** - Mostra esta mensagem de ajuda

💡 _Todos os comandos são processados através de uma fila para garantir confiabilidade!_`
	
	err := h.Discord.ReplyToMessage(payload.ChannelID, payload.MessageID, response)
	if err != nil {
		return fmt.Errorf("failed to send help response: %w", err)
	}

	return nil
}
```

---

### 1.4 - Criar `service/command_handlers/info_handler.go`

```go
package command_handlers

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
```

---

### 1.5 - Criar `service/command_handlers/calc_handler.go`

```go
package command_handlers

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
```

---

### 1.6 - Criar `service/command_handlers/unknown_handler.go`

```go
package command_handlers

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
```

---

## Passo 2: Atualizar ResponseHandler

Refatorar para usar EventDispatcher em vez de switch/case.

### Atualizar `service/handlers.go`:

```go
package service

import (
	"encoding/json"
	"enque-learning/events"
	"enque-learning/integration/discord"
	"enque-learning/integration/rabbitmq"
	"enque-learning/service/command_handlers"
	"fmt"
	"log"
	"strings"
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

	h.Discord.SendMessage(payload.ChannelID, fmt.Sprintf("⏳ Comando `%s` recebido e está sendo processado...", payload.Command))

	return nil
}

// ResponseHandler processa respostas da fila usando EventDispatcher
type ResponseHandler struct {
	Discord    *discord.Discord
	Dispatcher *events.EventDispatcher
}

func NewResponseHandler(discord *discord.Discord) *ResponseHandler {
	// Criar dispatcher para comandos
	dispatcher := events.NewEventDispatcher()

	// Registrar handlers específicos para cada comando
	dispatcher.RegisterHandler("discord.command.ping", command_handlers.NewPingCommandHandler(discord))
	dispatcher.RegisterHandler("discord.command.hello", command_handlers.NewHelloCommandHandler(discord))
	dispatcher.RegisterHandler("discord.command.oi", command_handlers.NewHelloCommandHandler(discord))
	dispatcher.RegisterHandler("discord.command.olá", command_handlers.NewHelloCommandHandler(discord))
	dispatcher.RegisterHandler("discord.command.help", command_handlers.NewHelpCommandHandler(discord))
	dispatcher.RegisterHandler("discord.command.ajuda", command_handlers.NewHelpCommandHandler(discord))
	dispatcher.RegisterHandler("discord.command.info", command_handlers.NewInfoCommandHandler(discord))
	dispatcher.RegisterHandler("discord.command.calc", command_handlers.NewCalcCommandHandler(discord))

	return &ResponseHandler{
		Discord:    discord,
		Dispatcher: dispatcher,
	}
}

func (h *ResponseHandler) ProcessMessage(message []byte) error { //// aquiiii
	var payload discord.DiscordCommandPayload
	err := json.Unmarshal(message, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Printf("processing response to Discord channel %s: %s", payload.ChannelID, payload.Command)

	// Criar evento específico para o comando
	eventName := fmt.Sprintf("discord.command.%s", strings.ToLower(payload.Command))
	event := events.NewEvent(eventName)
	event.Payload = payload

	// Despachar evento - se não houver handler registrado, trata como unknown
	err = h.Dispatcher.Dispatch(event)
	if err != nil || !h.hasHandlerForCommand(payload.Command) {
		// Comando desconhecido - usar handler padrão
		unknownHandler := command_handlers.NewUnknownCommandHandler(h.Discord)
		err = unknownHandler.HandleEvent(event)
		if err != nil {
			return fmt.Errorf("failed to handle unknown command: %w", err)
		}
	}

	log.Printf("response sent to Discord with success")
	return nil
}

// hasHandlerForCommand verifica se existe handler para o comando
func (h *ResponseHandler) hasHandlerForCommand(command string) bool {
	eventName := fmt.Sprintf("discord.command.%s", strings.ToLower(command))
	// Criar um handler dummy para testar
	dummyHandler := command_handlers.NewPingCommandHandler(h.Discord)
	return h.Dispatcher.HasHandler(eventName, dummyHandler) || 
	       len(h.Dispatcher.handlers[eventName]) > 0
}
```

**Problema:** O método `HasHandler` do EventDispatcher compara handlers por referência. Vamos criar um método auxiliar melhor:

### Atualizar `service/handlers.go` (versão corrigida):

```go
func (h *ResponseHandler) ProcessMessage(message []byte) error {
	var payload discord.DiscordCommandPayload
	err := json.Unmarshal(message, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Printf("processing response to Discord channel %s: %s", payload.ChannelID, payload.Command)

	// Criar evento específico para o comando
	eventName := fmt.Sprintf("discord.command.%s", strings.ToLower(payload.Command))
	event := events.NewEvent(eventName)
	event.Payload = payload

	// Tentar despachar evento
	err = h.Dispatcher.Dispatch(event)
	
	// Se não há handlers registrados para esse evento, é comando desconhecido
	if err == nil && h.handlerCount(eventName) == 0 {
		unknownHandler := command_handlers.NewUnknownCommandHandler(h.Discord)
		return unknownHandler.HandleEvent(event)
	}

	if err != nil {
		return fmt.Errorf("failed to process command: %w", err)
	}

	log.Printf("response sent to Discord with success")
	return nil
}

// handlerCount retorna o número de handlers registrados para um evento
func (h *ResponseHandler) handlerCount(eventName string) int {
	// Acessa diretamente o map interno (não é ideal, mas funciona)
	// Alternativa: adicionar método Count() no EventDispatcher
	handlers := reflect.ValueOf(h.Dispatcher).Elem().FieldByName("handlers")
	if handlers.IsValid() {
		handlersMap := handlers.Interface().(map[string][]events.EventHandlerInterface)
		return len(handlersMap[eventName])
	}
	return 0
}
```

**Melhor solução:** Adicionar método no EventDispatcher e usar reflection apenas se necessário.

### Versão FINAL simplificada de `service/handlers.go`:

```go
package service

import (
	"encoding/json"
	"enque-learning/events"
	"enque-learning/integration/discord"
	"enque-learning/integration/rabbitmq"
	"enque-learning/service/command_handlers"
	"fmt"
	"log"
	"strings"
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

	h.Discord.SendMessage(payload.ChannelID, fmt.Sprintf("⏳ Comando `%s` recebido e está sendo processado...", payload.Command))

	return nil
}

// ResponseHandler processa respostas da fila usando EventDispatcher
type ResponseHandler struct {
	Discord         *discord.Discord
	Dispatcher      *events.EventDispatcher
	UnknownHandler  *command_handlers.UnknownCommandHandler
	KnownCommands   map[string]bool // para checagem rápida
}

func NewResponseHandler(discord *discord.Discord) *ResponseHandler {
	dispatcher := events.NewEventDispatcher()
	knownCommands := make(map[string]bool)

	// Helper para registrar e marcar como conhecido
	register := func(cmdName string, handler events.EventHandlerInterface) {
		eventName := fmt.Sprintf("discord.command.%s", cmdName)
		dispatcher.RegisterHandler(eventName, handler)
		knownCommands[cmdName] = true
	}

	// Registrar handlers
	register("ping", command_handlers.NewPingCommandHandler(discord))
	
	helloHandler := command_handlers.NewHelloCommandHandler(discord)
	register("hello", helloHandler)
	register("oi", helloHandler)
	register("olá", helloHandler)
	
	helpHandler := command_handlers.NewHelpCommandHandler(discord)
	register("help", helpHandler)
	register("ajuda", helpHandler)
	
	register("info", command_handlers.NewInfoCommandHandler(discord))
	register("calc", command_handlers.NewCalcCommandHandler(discord))

	return &ResponseHandler{
		Discord:        discord,
		Dispatcher:     dispatcher,
		UnknownHandler: command_handlers.NewUnknownCommandHandler(discord),
		KnownCommands:  knownCommands,
	}
}

func (h *ResponseHandler) ProcessMessage(message []byte) error {
	var payload discord.DiscordCommandPayload
	err := json.Unmarshal(message, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Printf("processing response to Discord channel %s: %s", payload.ChannelID, payload.Command)

	// Normalizar comando
	command := strings.ToLower(payload.Command)
	
	// Verificar se é comando conhecido
	if !h.KnownCommands[command] {
		// Comando desconhecido
		event := events.NewEvent("discord.command.unknown")
		event.Payload = payload
		return h.UnknownHandler.HandleEvent(event)
	}

	// Criar e despachar evento para comando conhecido
	eventName := fmt.Sprintf("discord.command.%s", command)
	event := events.NewEvent(eventName)
	event.Payload = payload

	err = h.Dispatcher.Dispatch(event)
	if err != nil {
		return fmt.Errorf("failed to process command %s: %w", command, err)
	}

	log.Printf("response sent to Discord with success")
	return nil
}
```

---

## 📝 Checklist de Implementação

```
[ ] Criar service/command_handlers/ping_handler.go
[ ] Criar service/command_handlers/hello_handler.go
[ ] Criar service/command_handlers/help_handler.go
[ ] Criar service/command_handlers/info_handler.go
[ ] Criar service/command_handlers/calc_handler.go
[ ] Criar service/command_handlers/unknown_handler.go
[ ] Atualizar service/handlers.go
[ ] Testar todos os comandos
```

---

## 🎁 Benefícios

✅ **Usa infraestrutura existente** - Aproveita EventDispatcher  
✅ **Um arquivo por comando** - Fácil de adicionar/modificar  
✅ **Sem switch/case** - Dispatch automático por evento  
✅ **Testável** - Cada handler pode ser testado isoladamente  
✅ **Aliases nativos** - Reutiliza mesma instância do handler  

---

## 🚀 Como Adicionar Novo Comando

```go
// 1. Criar service/command_handlers/weather_handler.go
type WeatherCommandHandler struct {
    Discord *discord.Discord
}

func (h *WeatherCommandHandler) HandleEvent(event events.EventInterface) error {
    payload := event.GetPayload().(discord.DiscordCommandPayload)
    // ... lógica do comando
    return nil
}

// 2. Registrar em service/handlers.go (no NewResponseHandler)
register("weather", command_handlers.NewWeatherCommandHandler(discord))
register("tempo", command_handlers.NewWeatherCommandHandler(discord)) // alias
```

Pronto! Sistema estendido sem modificar código existente. 🎉
