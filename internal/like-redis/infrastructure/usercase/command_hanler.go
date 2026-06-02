package usercase

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/adapter/protocol"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/domain/command"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/domain/repository"
)

type CommandHandler struct {
	store   repository.KeyValueRepository
	persist repository.PersistenceRepository
	parser  *protocol.Parser
	stats   *Stats
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(
	store repository.KeyValueRepository,
	persist repository.PersistenceRepository,
	stats *Stats,
	parser *protocol.Parser,
) *CommandHandler {
	return &CommandHandler{
		store:   store,
		persist: persist,
		parser:  parser,
		stats:   stats,
	}
}

// ExecuteCommand executes a command and returns the response
func (h *CommandHandler) ExecuteCommand(ctx context.Context, cmd *protocol.Command) string {
	// Increment command counter
	h.stats.IncrementCommands()

	switch cmd.Type {
	case command.SET:
		return h.handleSet(ctx, cmd.Args)
	case command.GET:
		return h.handleGet(ctx, cmd.Args)
	case command.DEL:
		return h.handleDel(ctx, cmd.Args)
	case command.EXPIRE:
		return h.handleExpire(ctx, cmd.Args)
	case command.TTL:
		return h.handleTTL(ctx, cmd.Args)
	case command.PERSIST:
		return h.handlePersist(ctx, cmd.Args)
	case command.KEYS:
		return h.handleKeys(ctx, cmd.Args)
	case command.EXISTS:
		return h.handleExists(ctx, cmd.Args)
	case command.PING:
		return h.handlePing(ctx, cmd.Args)
	case command.INFO:
		return h.handleInfo(ctx, cmd.Args)
	default:
		return h.parser.FormatError(fmt.Sprintf("unknown command: %s", cmd.Type))
	}
}

func (h *CommandHandler) handleSet(ctx context.Context, args []string) string {
	if len(args) < 2 {
		return h.parser.FormatError("SET requires at least 2 arguments")
	}

	key := args[0]
	value := strings.Join(args[1:], " ")
	h.store.Set(ctx, key, value)

	if h.persist != nil {
		h.persist.Append(ctx, command.SET.String(), args)
	}

	return h.parser.FormatOK()
}

func (h *CommandHandler) handleGet(ctx context.Context, args []string) string {
	if len(args) < 1 {
		return h.parser.FormatError("GET requires 1 argument")
	}

	value, found := h.store.Get(ctx, args[0])
	if found {
		return value
	}
	return h.parser.FormatNil()
}

func (h *CommandHandler) handleDel(ctx context.Context, args []string) string {
	if len(args) < 1 {
		return h.parser.FormatError("DEL requires at least 1 argument")
	}

	count := 0
	for _, key := range args {
		count += h.store.Del(ctx, key)
	}

	if h.persist != nil && count > 0 {
		for _, key := range args {
			h.persist.Append(ctx, command.DEL.String(), []string{key})
		}
	}

	return h.parser.FormatResponse(count)
}

func (h *CommandHandler) handleExpire(ctx context.Context, args []string) string {
	if len(args) < 2 {
		return h.parser.FormatError("EXPIRE requires 2 arguments")
	}

	key := args[0]
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		return h.parser.FormatError("invalid seconds value")
	}

	success := h.store.Expire(ctx, key, seconds)
	if success {
		if h.persist != nil {
			h.persist.Append(ctx, command.EXPIRE.String(), args)
		}
		return h.parser.FormatOK()
	}

	return h.parser.FormatResponse(0)
}

func (h *CommandHandler) handleTTL(ctx context.Context, args []string) string {
	if len(args) < 1 {
		return h.parser.FormatError("TTL requires 1 argument")
	}

	ttl := h.store.TTL(ctx, args[0])
	return h.parser.FormatResponse(ttl)
}

func (h *CommandHandler) handlePersist(ctx context.Context, args []string) string {
	if len(args) < 1 {
		return h.parser.FormatError("PERSIST requires 1 argument")
	}

	success := h.store.Persist(ctx, args[0])
	if success {
		if h.persist != nil {
			h.persist.Append(ctx, command.PERSIST.String(), args)
		}
		return h.parser.FormatOK()
	}

	return h.parser.FormatResponse(0)
}

func (h *CommandHandler) handleKeys(ctx context.Context, args []string) string {
	pattern := "*"
	if len(args) > 0 {
		pattern = args[0]
	}

	keys := h.store.Keys(ctx, pattern)

	// Format as space-separated list
	if len(keys) == 0 {
		return ""
	}
	return strings.Join(keys, " ")
}

func (h *CommandHandler) handleExists(ctx context.Context, args []string) string {
	if len(args) < 1 {
		return h.parser.FormatError("EXISTS requires at least 1 argument")
	}

	count := 0
	for _, key := range args {
		if h.store.Exists(ctx, key) {
			count++
		}
	}

	return h.parser.FormatResponse(count)
}

func (h *CommandHandler) handlePing(ctx context.Context, args []string) string {
	message := "PONG"
	if len(args) > 0 {
		message = strings.Join(args, " ")
	}
	return message
}

func (h *CommandHandler) handleInfo(ctx context.Context, args []string) string {
	section := ""
	if len(args) > 0 {
		section = strings.ToUpper(args[0])
	}

	info := h.stats.GetInfo(ctx)

	// If section is specified, filter output (for simplicity, return all for now)
	// In a real implementation, you'd parse and filter by section
	if section != "" && section != "ALL" && section != "DEFAULT" {
		// For now, return all info regardless of section
		// This could be enhanced to filter by section (server, clients, stats, keyspace)
	}

	return info
}
