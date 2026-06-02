package persistence

import "github.com/vituu69/tiketis/internal/like-redis/infrastructure/domain/repository"

type AOFProviderOption struct {
	EnableAOF bool
	Filepath  string
}

// NewAOFProvider creates an AOF repository if enabled, otherwise returns nil
func NewAOFProvider(opt AOFProviderOption) (repository.PersistenceRepository, error) {
	if !opt.EnableAOF {
		return nil, nil
	}
	return NewAOF(opt.Filepath)
}
