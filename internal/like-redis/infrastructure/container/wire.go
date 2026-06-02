//go:build wireinject
// +build wireinject

package container

import (
	"github.com/google/wire"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/adapter/handler"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/adapter/protocol"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/persistence"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/storage"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/usercase"
)

func InitializeContainer(opt persistence.AOFProviderOption) (*Container, func(), error) {
	wire.Build(
		// Infrastructure providers
		storage.NewStore,
		persistence.NewAOFProvider,

		// Adapter providers
		protocol.NewParser,

		// Use case providers
		usercase.NewStats,
		usercase.NewCommandHandler,

		// Handler providers
		handler.NewTCPHandler,

		// Container provider
		NewContainer,
	)
	return nil, nil, nil
}
