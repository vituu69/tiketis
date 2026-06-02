package handler

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"time"

	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/adapter/protocol"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/domain/command"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/usercase"
)

// TCPHandler handles TCP connections
type TCPHandler struct {
	commandHandler *usercase.CommandHandler
	parser         *protocol.Parser
}

// NewTCPHandler creates a new TCP handler
func NewTCPHandler(commandHandler *usercase.CommandHandler, parser *protocol.Parser) *TCPHandler {
	return &TCPHandler{
		commandHandler: commandHandler,
		parser:         parser,
	}
}

// HandleConnection handles a single TCP connection
func (h *TCPHandler) HandleConnection(conn net.Conn) {
	defer conn.Close()

	// Create context for this connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Increment connection counter (if we had access to stats, we'd do it here)
	// For now, this is handled at a higher level if needed

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		// Check context cancellation
		if ctx.Err() != nil {
			return
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse command
		cmd, err := h.parser.ParseCommand(line)
		if err != nil {
			response := h.parser.FormatError(err.Error())
			h.writeResponse(conn, response)
			continue
		}

		// Handle QUIT command
		if cmd.Type == command.QUIT {
			h.writeResponse(conn, h.parser.FormatOK())
			return
		}

		// Execute command with context
		response := h.commandHandler.ExecuteCommand(ctx, cmd)
		h.writeResponse(conn, response)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Error reading from connection: %v", err)
	}
}

func (h *TCPHandler) writeResponse(conn net.Conn, response string) {
	conn.Write([]byte(response + "\n"))
}
