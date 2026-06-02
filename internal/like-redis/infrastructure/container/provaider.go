package container

import (
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/adapter/handler"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/adapter/protocol"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/domain/repository"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/usercase"
)

type Container struct {
	Store          repository.KeyValueRepository
	Persistence    repository.PersistenceRepository
	CommandHandler *usercase.CommandHandler
	TCPHandler     *handler.TCPHandler
	Parser         *protocol.Parser
}

// NewContainer creates a new dependency injection container
func NewContainer(
	store repository.KeyValueRepository,
	persist repository.PersistenceRepository,
	parser *protocol.Parser,
	commandHandler *usercase.CommandHandler,
	tcpHandler *handler.TCPHandler,
) *Container {
	return &Container{
		Store:          store,
		Persistence:    persist,
		CommandHandler: commandHandler,
		TCPHandler:     tcpHandler,
		Parser:         parser,
	}
}

// Close closes all resources that need cleanup
func (c *Container) Close() error {
	if c.Persistence != nil {
		return c.Persistence.Close()
	}
	return nil
}
