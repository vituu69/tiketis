package service

import "github.com/vituu69/tiketis/internal/db"

type IngressService struct {
	store *db.Store
}

func NewIngressService(store *db.Store) *IngressService {
	return &IngressService{store: store}
}
