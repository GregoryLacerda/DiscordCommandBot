package server

import (
	"enque-learning/events"
	"enque-learning/handlers"
	"enque-learning/integration"
	"enque-learning/internal/config"
	"enque-learning/service"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Server struct {
	config          *config.Config
	EventDispatcher *events.EventDispatcher
	integrations    *integration.Integrations
	service         *service.Service
	CommandHandler  *handlers.CommandHandler
	ResponseHandler *handlers.ResponseHandler
}

func NewServer(cfg *config.Config, eventDispatcher *events.EventDispatcher, integrations *integration.Integrations, service *service.Service) *Server {
	return &Server{
		config:          cfg,
		EventDispatcher: eventDispatcher,
		integrations:    integrations,
		service:         service,
		CommandHandler:  handlers.NewCommandHandler(integrations.RabbitMQ, integrations.Discord),
		ResponseHandler: handlers.NewResponseHandler(integrations.Discord, eventDispatcher, service),
	}
}

func (s *Server) StartAll() error {
	log.Println("🚀 Iniciando sistema completo (Producer + Consumer)...")

	s.EventDispatcher.RegisterHandler("discord.command.received", s.CommandHandler)

	// Iniciar Discord
	err := s.integrations.Discord.Start()
	if err != nil {
		return fmt.Errorf("erro ao iniciar Discord: %w", err)
	}

	// Iniciar Consumer
	msgs, err := s.integrations.RabbitMQ.Consumer()
	if err != nil {
		return fmt.Errorf("erro ao iniciar consumer: %w", err)
	}

	log.Println("✅ Sistema completo iniciado com sucesso!")

	// Processar mensagens em background

	for msg := range msgs {
		log.Printf("📨 Mensagem recebida da fila")

		go func(msg amqp.Delivery) {
			err := s.ResponseHandler.ProcessMessage(msg.Body)
			if err != nil {
				log.Printf("❌ Erro ao processar mensagem: %v", err)
				msg.Nack(false, true)
			} else {
				log.Printf("✅ Mensagem processada com sucesso")
				msg.Ack(false)
			}
		}(msg)
	}

	// Aguardar shutdown
	s.waitForShutdown()

	return s.Shutdown()
}

func (s *Server) waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("⚠️ Sinal de interrupção recebido...")
}

func (s *Server) Shutdown() error {
	log.Println("🛑 Encerrando sistema...")

	if s.integrations.Discord != nil {
		s.integrations.Discord.Stop()
	}

	if s.integrations.RabbitMQ != nil {
		s.integrations.RabbitMQ.Close()
	}

	log.Println("✅ Sistema encerrado com sucesso!")
	return nil
}
