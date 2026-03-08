# enque-learning - Development Guidelines

## Project Overview
Event-driven Discord bot in Go using RabbitMQ for async command processing. Commands flow through EventDispatcher → Handlers → Services → Discord responses.

**Full architecture**: See AGENTS.md for complete system documentation.

## Critical Patterns

### File Organization (MANDATORY)
**ONE FILE PER COMMAND** - Non-negotiable for handlers and services.

```
handlers/<command>.go     # One handler per command
service/<command>.go      # One service method file per command
service/<feature>_utils.go # Shared utilities only
```

**DO NOT** create monolithic files. Each command gets its own file.

### Code Language
- ALL logs, errors, comments: **English only**
- Emoji prefixes: ✅ (success), ❌ (error), 🚀 (start), 🔍 (check), 🛑 (stop)
- **NO log.Printf()**: Use `logger.Debug()`, `logger.Info()`, `logger.Warn()`, `logger.Critical()`
- **NO fmt.Errorf()**: Use `errors.New*()` constructors

### Error & Logging System
**Use custom error system in pkg/**

```go
// Create errors with specific constructors
err := errors.NewValidation("missing argument", nil)
err := errors.NewService("monitoring in progress", nil)
err := errors.NewAPI("Twitch API failed", originalErr)

// Log with appropriate level
logger.Debug("Cache check: %s", key)          // Only shown if DEBUG_MODE=true
logger.Info("✅ Command executed: %s", cmd)   // General information
logger.Warn("⚠️ API rate limit: %d", limit)  // Warnings
logger.Critical("❌ Database error: %v", err) // Critical issues

// Auto-detect error level
appErr := errors.NewAPI("failed", err)
logger.LogError(appErr)  // Logs at appropriate level
```

**Error types**: ValidationError, ServiceError, IntegrationError, APIError, ConfigError, NetworkError, AuthError, DatabaseError

**Debug mode**: Set `DEBUG_MODE=true` in `.env` to see debug logs

### Adding New Commands
1. **constants/commands.go**: Add usage/error constants
2. **service/<command>.go**: Create business logic methods
3. **handlers/<command>.go**: Create event handler
4. **handlers/handlers.go**: Register with `dispatcher.RegisterHandler()`

### Handler Pattern
```go
type CommandHandler struct {
    Discord *discord.Discord
    Service *service.Service
}

func (h *CommandHandler) HandleEvent(event EventInterface) error {
    // 1. Cast payload
    // 2. Log action
    // 3. Validate arguments → reply with constants.UsageMessage
    // 4. Execute via h.Service.Method()
    // 5. Reply to Discord
    return nil
}
```

### Service Layer Rules
- **service.go**: ONLY struct definition + constructor
- **<command>.go**: Business logic methods for that command
- No business logic in handlers
- Use constants for all messages
- Return custom errors: `errors.NewService()`, `errors.NewValidation()`

### Architecture Constraints
- Handlers implement `EventHandlerInterface`
- Services depend on `*integration.Integrations`
- Event names: `discord.command.<lowercase_command>`
- Inject dependencies via constructors (no globals)

### Concurrency
- Shared state: Always use mutex (`s.twitchMu.Lock()`)
- Long-running tasks: Use `context.WithCancel()` or `context.WithTimeout()`
- Goroutines: Ensure proper cleanup and error handling

## Common Mistakes
❌ Multiple commands in one file
❌ Business logic in handlers
❌ Hardcoded strings (use constants)
❌ Mixed language (Portuguese + English)
❌ Direct integration access in handlers
❌ Using `log.Printf()` instead of `logger.Info()`, `logger.Debug()`, etc.
❌ Using `fmt.Errorf()` instead of `errors.New*()` constructors

## Build & Test
```bash
go mod tidy                      # Update dependencies
go build -o bot.exe cmd/main.go  # Compile
docker-compose up -d             # Start RabbitMQ
go run cmd/main.go               # Run bot
```

## References
- **AGENTS.md**: Complete architecture and data flow
- **utils/USAGE_EXAMPLES.go**: Error and logging system examples
- **service/twitch_*.go**: Reference implementation following all patterns
- **handlers/twitch_*.go**: Handler examples with proper structure

---
*Keep this file minimal. For detailed architecture, see AGENTS.md.*

