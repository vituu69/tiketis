package container

import (
	"context"

	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/adapter/handler"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/adapter/protocol"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/persistence"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/storage"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/usercase"
)

func InitializeContainer(opt persistence.AOFProviderOption) (*Container, func(), error) {
	keyValueRepository := storage.NewStore()
	persistenceRepository, err := persistence.NewAOFProvider(opt)
	if err != nil {
		return nil, nil, err
	}
	if persistenceRepository != nil {
		if err := persistenceRepository.Replay(context.Background(), keyValueRepository); err != nil {
			persistenceRepository.Close()
			return nil, nil, err
		}
	}
	keyValueRepository.StartCleanup(1000)
	parser := protocol.NewParser()
	stats := usercase.NewStats(keyValueRepository)
	commandHandler := usercase.NewCommandHandler(keyValueRepository, persistenceRepository, stats, parser)
	tcpHandler := handler.NewTCPHandler(commandHandler, parser)
	container := NewContainer(keyValueRepository, persistenceRepository, parser, commandHandler, tcpHandler)
	return container, func() {
		keyValueRepository.StopCleanup()
		container.Close()
	}, nil
}
