# Plano de Implementação - Sistema Discord com RabbitMQ

## Visão Geral do Fluxo

```
Discord (Comando) → Event Dispatcher → Handler → RabbitMQ (Publisher) → Fila
                                                                           ↓
Discord (Resposta) ← Handler ← Consumer ← RabbitMQ (Consumer) ← Fila
```

## Estado Atual do Código

### ✅ Já Implementado:
- Sistema de eventos (Event, EventDispatcher, Interfaces)
- Estrutura básica do RabbitMQ (Publisher e Consumer)
- Estrutura de configuração
- Estrutura básica do Discord e Service

### ❌ Falta Implementar:
1. Integração completa com Discord
2. Handlers de eventos
3. Service orquestrador
4. Main.go (producer e consumer)
5. Configurações de ambiente

---

## Passo 1: Adicionar Dependências

### Adicionar no go.mod:
```bash
go get github.com/bwmarrin/discordgo
```

### go.mod atualizado:
```go
module enque-learning

go 1.26

require (
	github.com/bwmarrin/discordgo v0.28.1
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/rabbitmq/amqp091-go v1.10.0
)
```

---

## Passo 2: Configurar Variáveis de Ambiente

### Criar arquivo `.env`:
```env
# Discord
DISCORD_TOKEN=seu_token_aqui
DISCORD_COMMAND_PREFIX=!

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
QUEUE_NAME=discord-commands
EXCHANGE_NAME=discord-exchange
ROUTING_KEY=discord.command

# Server
WEB_SERVER_PORT=8080
```

### Atualizar `internal/config/config.go`:
```go
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
	RabbitMQURL   string
	QueueName     string
	ExchangeName  string
	RoutingKey    string
	
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
```

---

## Passo 3: Criar Estrutura de Payload

### Criar `events/payload.go`:
```go
package events

// DiscordCommandPayload representa o payload de um comando do Discord
type DiscordCommandPayload struct {
	UserID      string            `json:"user_id"`
	Username    string            `json:"username"`
	ChannelID   string            `json:"channel_id"`
	GuildID     string            `json:"guild_id"`
	Command     string            `json:"command"`
	Arguments   []string          `json:"arguments"`
	MessageID   string            `json:"message_id"`
	Timestamp   string            `json:"timestamp"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// DiscordResponsePayload representa a resposta a ser enviada ao Discord
type DiscordResponsePayload struct {
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
	MessageID string `json:"message_id,omitempty"` // Para responder a uma mensagem específica
}
```

---

## Passo 4: Implementar Integração com Discord

### Atualizar `integration/discord/discord.go`:
```go
package discord

import (
	"enque-learning/events"
	"enque-learning/internal/config"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Config     config.Config
	Session    *discordgo.Session
	Dispatcher events.EventDispatcherInterface
}

func NewDiscordIntegration(config config.Config, dispatcher events.EventDispatcherInterface) (*Discord, error) {
	session, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar sessão do Discord: %w", err)
	}

	discord := &Discord{
		Config:     config,
		Session:    session,
		Dispatcher: dispatcher,
	}

	// Registrar handler de mensagens
	session.AddHandler(discord.messageHandler)

	return discord, nil
}

// Start inicia a conexão com o Discord
func (d *Discord) Start() error {
	err := d.Session.Open()
	if err != nil {
		return fmt.Errorf("erro ao abrir conexão com Discord: %w", err)
	}

	log.Println("Bot Discord conectado e online!")
	return nil
}

// Stop encerra a conexão com o Discord
func (d *Discord) Stop() error {
	return d.Session.Close()
}

// messageHandler processa mensagens recebidas do Discord
func (d *Discord) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignorar mensagens do próprio bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Verificar se é um comando (começa com o prefixo)
	if !strings.HasPrefix(m.Content, d.Config.DiscordCommandPrefix) {
		return
	}

	// Parsear comando e argumentos
	content := strings.TrimPrefix(m.Content, d.Config.DiscordCommandPrefix)
	parts := strings.Fields(content)
	
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	arguments := []string{}
	if len(parts) > 1 {
		arguments = parts[1:]
	}

	// Criar payload do comando
	payload := events.DiscordCommandPayload{
		UserID:    m.Author.ID,
		Username:  m.Author.Username,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
		Command:   command,
		Arguments: arguments,
		MessageID: m.ID,
		Timestamp: m.Timestamp.String(),
	}

	// Criar e despachar evento
	event := events.NewEvent("discord.command.received")
	event.Payload = payload

	log.Printf("Comando recebido: %s de %s", command, m.Author.Username)

	// Despachar evento (será processado pelos handlers registrados)
	err := d.Dispatcher.Dispatch(event)
	if err != nil {
		log.Printf("Erro ao processar comando: %v", err)
		d.SendMessage(m.ChannelID, "❌ Erro ao processar comando!")
	}
}

// SendMessage envia uma mensagem para um canal do Discord
func (d *Discord) SendMessage(channelID, message string) error {
	_, err := d.Session.ChannelMessageSend(channelID, message)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}
	return nil
}

// ReplyToMessage responde a uma mensagem específica
func (d *Discord) ReplyToMessage(channelID, messageID, message string) error {
	_, err := d.Session.ChannelMessageSendReply(channelID, message, &discordgo.MessageReference{
		MessageID: messageID,
		ChannelID: channelID,
	})
	if err != nil {
		return fmt.Errorf("erro ao responder mensagem: %w", err)
	}
	return nil
}
```

---

## Passo 5: Atualizar RabbitMQ

### Atualizar `integration/rabbitmq/rabbitmq.go`:
```go
package rabbitmq

import (
	"context"
	"enque-learning/internal/config"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Config config.Config
	Conn   *amqp.Connection
	Ch     *amqp.Channel
}

func NewRabbitMQIntegration(config config.Config) (*RabbitMQ, error) {
	conn, err := amqp.Dial(config.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao abrir canal: %w", err)
	}

	rmq := &RabbitMQ{
		Config: config,
		Conn:   conn,
		Ch:     ch,
	}

	// Configurar exchange e fila
	if err := rmq.setup(); err != nil {
		rmq.Close()
		return nil, err
	}

	return rmq, nil
}

// setup configura exchange, fila e binding
func (r *RabbitMQ) setup() error {
	// Declarar exchange
	err := r.Ch.ExchangeDeclare(
		r.Config.ExchangeName, // name
		"topic",               // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return fmt.Errorf("erro ao declarar exchange: %w", err)
	}

	// Declarar fila
	_, err = r.Ch.QueueDeclare(
		r.Config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("erro ao declarar fila: %w", err)
	}

	// Bind da fila ao exchange
	err = r.Ch.QueueBind(
		r.Config.QueueName,    // queue name
		r.Config.RoutingKey,   // routing key
		r.Config.ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("erro ao fazer bind da fila: %w", err)
	}

	log.Println("RabbitMQ configurado com sucesso!")
	return nil
}

// Publisher publica uma mensagem na fila
func (r *RabbitMQ) Publisher(body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.Ch.PublishWithContext(
		ctx,
		r.Config.ExchangeName, // exchange
		r.Config.RoutingKey,   // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Mensagem persistente
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("erro ao publicar mensagem: %w", err)
	}

	log.Printf("Mensagem publicada: %s", string(body))
	return nil
}

// Consumer retorna um canal para consumir mensagens
func (r *RabbitMQ) Consumer() (<-chan amqp.Delivery, error) {
	msgs, err := r.Ch.Consume(
		r.Config.QueueName, // queue
		"go-consumer",      // consumer
		false,              // auto-ack (false para controle manual)
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar consumer: %w", err)
	}

	return msgs, nil
}

// Close fecha a conexão com o RabbitMQ
func (r *RabbitMQ) Close() error {
	if r.Ch != nil {
		if err := r.Ch.Close(); err != nil {
			return err
		}
	}
	if r.Conn != nil {
		if err := r.Conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
```

---

## Passo 6: Criar Handlers de Eventos

### Criar `service/handlers.go`:
```go
package service

import (
	"encoding/json"
	"enque-learning/events"
	"enque-learning/integration/discord"
	"enque-learning/integration/rabbitmq"
	"fmt"
	"log"
	"strings"
)

// CommandHandler processa comandos do Discord e envia para a fila
type CommandHandler struct {
	RabbitMQ *rabbitmq.RabbitMQ
	Discord  *discord.Discord
}

func NewCommandHandler(rmq *rabbitmq.RabbitMQ, discord *discord.Discord) *CommandHandler {
	return &CommandHandler{
		RabbitMQ: rmq,
		Discord:  discord,
	}
}

// HandleEvent implementa EventHandlerInterface
func (h *CommandHandler) HandleEvent(event events.EventInterface) error {
	payload, ok := event.GetPayload().(events.DiscordCommandPayload)
	if !ok {
		return fmt.Errorf("payload inválido")
	}

	log.Printf("Processando comando: %s", payload.Command)

	// Serializar payload para JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	// Publicar na fila RabbitMQ
	err = h.RabbitMQ.Publisher(jsonData)
	if err != nil {
		return fmt.Errorf("erro ao publicar na fila: %w", err)
	}

	// Enviar confirmação no Discord
	h.Discord.SendMessage(payload.ChannelID, 
		fmt.Sprintf("⏳ Comando `%s` recebido e está sendo processado...", payload.Command))

	return nil
}

// ResponseHandler processa respostas da fila e envia para o Discord
type ResponseHandler struct {
	Discord *discord.Discord
}

func NewResponseHandler(discord *discord.Discord) *ResponseHandler {
	return &ResponseHandler{
		Discord: discord,
	}
}

// ProcessMessage processa uma mensagem da fila e envia resposta ao Discord
func (h *ResponseHandler) ProcessMessage(data []byte) error {
	var payload events.DiscordCommandPayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return fmt.Errorf("erro ao deserializar payload: %w", err)
	}

	log.Printf("Processando resposta para comando: %s", payload.Command)

	// Processar o comando e gerar resposta
	response := h.processCommand(payload)

	// Enviar resposta ao Discord
	err = h.Discord.ReplyToMessage(payload.ChannelID, payload.MessageID, response)
	if err != nil {
		return fmt.Errorf("erro ao enviar resposta: %w", err)
	}

	log.Printf("Resposta enviada ao Discord com sucesso")
	return nil
}

// processCommand processa o comando e retorna uma resposta
func (h *ResponseHandler) processCommand(payload events.DiscordCommandPayload) string {
	command := strings.ToLower(payload.Command)
	
	switch command {
	case "ping":
		return "🏓 Pong!"
	
	case "hello", "oi", "olá":
		return fmt.Sprintf("👋 Olá, %s! Como posso ajudar?", payload.Username)
	
	case "help", "ajuda":
		return `📚 **Comandos Disponíveis:**
		
!ping - Testa se o bot está respondendo
!hello - Recebe uma saudação
!calc <expressão> - Calcula uma expressão matemática
!info - Mostra informações sobre o sistema
!help - Mostra esta mensagem de ajuda`
	
	case "calc":
		if len(payload.Arguments) == 0 {
			return "❌ Uso: !calc <expressão>\nExemplo: !calc 2 + 2"
		}
		// Aqui você implementaria a lógica de cálculo
		expression := strings.Join(payload.Arguments, " ")
		return fmt.Sprintf("🧮 Calculando: %s\n(Implementar lógica de cálculo)", expression)
	
	case "info":
		return fmt.Sprintf(`ℹ️ **Informações do Sistema:**
		
👤 Usuário: %s
🆔 User ID: %s
📝 Comando: %s
⏰ Timestamp: %s`, 
			payload.Username, 
			payload.UserID, 
			payload.Command,
			payload.Timestamp)
	
	default:
		return fmt.Sprintf("❓ Comando desconhecido: `%s`\nUse !help para ver os comandos disponíveis.", command)
	}
}
```

---

## Passo 7: Implementar Service

### Atualizar `service/service.go`:
```go
package service

import (
	"enque-learning/events"
	"enque-learning/integration/discord"
	"enque-learning/integration/rabbitmq"
	"enque-learning/internal/config"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Service struct {
	Config           *config.Config
	EventDispatcher  *events.EventDispatcher
	Discord          *discord.Discord
	RabbitMQ         *rabbitmq.RabbitMQ
	CommandHandler   *CommandHandler
	ResponseHandler  *ResponseHandler
}

func NewService() (*Service, error) {
	// Carregar configurações
	cfg := config.LoadConfig()

	// Criar Event Dispatcher
	dispatcher := events.NewEventDispatcher()

	// Criar integração RabbitMQ
	rmq, err := rabbitmq.NewRabbitMQIntegration(*cfg)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar RabbitMQ: %w", err)
	}

	// Criar integração Discord
	discordBot, err := discord.NewDiscordIntegration(*cfg, dispatcher)
	if err != nil {
		rmq.Close()
		return nil, fmt.Errorf("erro ao inicializar Discord: %w", err)
	}

	// Criar handlers
	commandHandler := NewCommandHandler(rmq, discordBot)
	responseHandler := NewResponseHandler(discordBot)

	// Registrar handler para comandos do Discord
	err = dispatcher.RegisterHandler("discord.command.received", commandHandler)
	if err != nil {
		rmq.Close()
		return nil, fmt.Errorf("erro ao registrar handler: %w", err)
	}

	return &Service{
		Config:          cfg,
		EventDispatcher: dispatcher,
		Discord:         discordBot,
		RabbitMQ:        rmq,
		CommandHandler:  commandHandler,
		ResponseHandler: responseHandler,
	}, nil
}

// StartProducer inicia apenas o producer (Discord -> RabbitMQ)
func (s *Service) StartProducer() error {
	log.Println("🚀 Iniciando Producer (Discord Bot)...")

	// Iniciar Discord
	err := s.Discord.Start()
	if err != nil {
		return fmt.Errorf("erro ao iniciar Discord: %w", err)
	}

	log.Println("✅ Producer iniciado com sucesso!")
	
	// Aguardar sinal de interrupção
	s.waitForShutdown()
	
	return s.Shutdown()
}

// StartConsumer inicia apenas o consumer (RabbitMQ -> Discord)
func (s *Service) StartConsumer() error {
	log.Println("🚀 Iniciando Consumer (Processador de Fila)...")

	// Obter canal de mensagens do RabbitMQ
	msgs, err := s.RabbitMQ.Consumer()
	if err != nil {
		return fmt.Errorf("erro ao iniciar consumer: %w", err)
	}

	log.Println("✅ Consumer iniciado com sucesso! Aguardando mensagens...")

	// Canal para shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Processar mensagens
	go func() {
		for msg := range msgs {
			log.Printf("📨 Mensagem recebida da fila")

			// Processar mensagem
			err := s.ResponseHandler.ProcessMessage(msg.Body)
			if err != nil {
				log.Printf("❌ Erro ao processar mensagem: %v", err)
				msg.Nack(false, true) // Rejeitar e recolocar na fila
			} else {
				log.Printf("✅ Mensagem processada com sucesso")
				msg.Ack(false) // Confirmar processamento
			}
		}
	}()

	// Aguardar sinal de shutdown
	<-stop
	log.Println("⚠️ Sinal de interrupção recebido, encerrando consumer...")

	return s.Shutdown()
}

// StartAll inicia producer e consumer juntos
func (s *Service) StartAll() error {
	log.Println("🚀 Iniciando sistema completo (Producer + Consumer)...")

	// Iniciar Discord
	err := s.Discord.Start()
	if err != nil {
		return fmt.Errorf("erro ao iniciar Discord: %w", err)
	}

	// Iniciar Consumer
	msgs, err := s.RabbitMQ.Consumer()
	if err != nil {
		return fmt.Errorf("erro ao iniciar consumer: %w", err)
	}

	log.Println("✅ Sistema completo iniciado com sucesso!")

	// Processar mensagens em background
	go func() {
		for msg := range msgs {
			log.Printf("📨 Mensagem recebida da fila")

			err := s.ResponseHandler.ProcessMessage(msg.Body)
			if err != nil {
				log.Printf("❌ Erro ao processar mensagem: %v", err)
				msg.Nack(false, true)
			} else {
				log.Printf("✅ Mensagem processada com sucesso")
				msg.Ack(false)
			}
		}
	}()

	// Aguardar shutdown
	s.waitForShutdown()
	
	return s.Shutdown()
}

// waitForShutdown aguarda sinal de interrupção
func (s *Service) waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("⚠️ Sinal de interrupção recebido...")
}

// Shutdown encerra todas as conexões
func (s *Service) Shutdown() error {
	log.Println("🛑 Encerrando sistema...")

	if s.Discord != nil {
		s.Discord.Stop()
	}

	if s.RabbitMQ != nil {
		s.RabbitMQ.Close()
	}

	log.Println("✅ Sistema encerrado com sucesso!")
	return nil
}
```

---

## Passo 8: Implementar Main

### Atualizar `cmd/main.go`:
```go
package main

import (
	"enque-learning/service"
	"flag"
	"log"
)

func main() {
	// Flags para escolher o modo de execução
	mode := flag.String("mode", "all", "Modo de execução: producer, consumer ou all")
	flag.Parse()

	// Criar service
	svc, err := service.NewService()
	if err != nil {
		log.Fatalf("❌ Erro ao criar service: %v", err)
	}

	// Executar baseado no modo
	switch *mode {
	case "producer":
		log.Println("Modo: PRODUCER (Discord -> RabbitMQ)")
		err = svc.StartProducer()
	case "consumer":
		log.Println("Modo: CONSUMER (RabbitMQ -> Discord)")
		err = svc.StartConsumer()
	case "all":
		log.Println("Modo: ALL (Producer + Consumer)")
		err = svc.StartAll()
	default:
		log.Fatalf("❌ Modo inválido: %s. Use: producer, consumer ou all", *mode)
	}

	if err != nil {
		log.Fatalf("❌ Erro ao executar service: %v", err)
	}
}
```

---

## Passo 9: Configurar RabbitMQ

### Docker Compose (Opcional - `docker-compose.yml`):
```yaml
version: '3.8'

services:
  rabbitmq:
    image: rabbitmq:3.12-management
    container_name: rabbitmq
    ports:
      - "5672:5672"   # AMQP
      - "15672:15672" # Management UI
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

volumes:
  rabbitmq_data:
```

### Comandos do RabbitMQ:
```bash
# Iniciar RabbitMQ com Docker
docker-compose up -d

# Acessar Management UI
http://localhost:15672
# Login: guest / guest
```

---

## Passo 10: Criar Bot no Discord

### Passos:
1. Acesse: https://discord.com/developers/applications
2. Clique em "New Application"
3. Dê um nome ao bot
4. Vá em "Bot" no menu lateral
5. Clique em "Add Bot"
6. Copie o Token e adicione no `.env`
7. Em "Privileged Gateway Intents", habilite:
   - Message Content Intent
8. Vá em "OAuth2" > "URL Generator"
9. Selecione scopes: `bot`
10. Selecione permissões: `Send Messages`, `Read Messages/View Channels`
11. Copie a URL gerada e adicione o bot ao seu servidor

---

## Passo 11: Executar o Sistema

### Opção 1: Executar Tudo Junto
```bash
go run cmd/main.go -mode=all
```

### Opção 2: Executar Producer e Consumer Separados

**Terminal 1 - Producer:**
```bash
go run cmd/main.go -mode=producer
```

**Terminal 2 - Consumer:**
```bash
go run cmd/main.go -mode=consumer
```

---

## Passo 12: Testar

### No Discord:
```
!ping
!hello
!help
!info
!calc 2 + 2
```

### Fluxo Esperado:
1. Você envia `!ping` no Discord
2. Bot responde: "⏳ Comando `ping` recebido e está sendo processado..."
3. Mensagem é enviada para o RabbitMQ
4. Consumer processa a mensagem
5. Bot responde: "🏓 Pong!"

---

## Estrutura Final de Arquivos

```
enque-learning/
├── cmd/
│   └── main.go                    # ✏️ Atualizar
├── events/
│   ├── dispatcher.go              # ✅ OK
│   ├── event.go                   # ✅ OK
│   ├── interfaces.go              # ✅ OK
│   └── payload.go                 # 📝 Criar
├── integration/
│   ├── discord/
│   │   └── discord.go             # ✏️ Atualizar
│   └── rabbitmq/
│       └── rabbitmq.go            # ✏️ Atualizar
├── internal/
│   ├── config/
│   │   └── config.go              # ✏️ Atualizar
│   └── errors/
│       └── errors.go              # ✅ OK
├── service/
│   ├── service.go                 # ✏️ Atualizar
│   └── handlers.go                # 📝 Criar
├── .env                           # 📝 Criar
├── docker-compose.yml             # 📝 Criar (opcional)
└── go.mod                         # ✏️ Atualizar
```

---

## Melhorias Futuras

### 1. **Adicionar Mais Comandos**
- Implementar calculadora funcional
- Integração com APIs externas (clima, cotações, etc.)
- Comandos administrativos

### 2. **Persistência**
- Adicionar banco de dados (PostgreSQL, MongoDB)
- Armazenar histórico de comandos
- Sistema de usuários

### 3. **Observabilidade**
- Adicionar logs estruturados (zerolog, zap)
- Métricas com Prometheus
- Tracing com OpenTelemetry

### 4. **Resiliência**
- Retry com backoff exponencial
- Circuit breaker
- Dead letter queue

### 5. **Testes**
- Testes unitários
- Testes de integração
- Mocks para Discord e RabbitMQ

---

## Troubleshooting

### Erro: "Invalid Discord Token"
- Verifique se o token está correto no `.env`
- Certifique-se de incluir "Bot " antes do token no código

### Erro: "Connection refused" (RabbitMQ)
- Verifique se o RabbitMQ está rodando: `docker ps`
- Verifique a URL no `.env`

### Bot não responde a comandos
- Verifique se "Message Content Intent" está habilitado
- Verifique os logs para ver se o bot está recebendo mensagens

### Mensagens ficam presas na fila
- Verifique se o consumer está rodando
- Verifique os logs para erros de processamento

---

## Comandos Úteis

```bash
# Instalar dependências
go mod tidy

# Executar
go run cmd/main.go -mode=all

# Build
go build -o bin/app cmd/main.go

# Executar binário
./bin/app -mode=all

# Ver logs do RabbitMQ
docker logs rabbitmq

# Limpar fila do RabbitMQ
docker exec rabbitmq rabbitmqctl purge_queue discord-commands
```

---

## Conclusão

Este plano fornece uma implementação completa do sistema Discord + RabbitMQ.

**Ordem de Implementação Recomendada:**
1. ✅ Configurar ambiente (.env, dependências)
2. ✅ Atualizar config.go
3. ✅ Criar payload.go
4. ✅ Atualizar rabbitmq.go
5. ✅ Atualizar discord.go
6. ✅ Criar handlers.go
7. ✅ Atualizar service.go
8. ✅ Atualizar main.go
9. ✅ Testar!

Boa sorte com a implementação! 🚀
