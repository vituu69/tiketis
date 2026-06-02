package storage

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	domain "github.com/vituu69/tiketis/internal/like-redis/infrastructure/domain/entity"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/domain/repository"
)

type StoreRedis struct {
	data        map[string]*domain.Item
	mu          sync.RWMutex
	stopCleanup chan struct{}
	stopOnce    sync.Once
}

func NewStore() repository.KeyValueRepository {
	return &StoreRedis{
		data:        make(map[string]*domain.Item),
		stopCleanup: make(chan struct{}),
	}
}

func (s *StoreRedis) Set(ctx context.Context, key, value string) {
	if ctx.Err() != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &domain.Item{Value: value, ExpiresAt: nil}
}

func (s *StoreRedis) SetEX(ctx context.Context, key, value string, ttl time.Duration) {
	if ctx.Err() != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &domain.Item{Value: value, ExpiresAt: expiresAt(ttl)}
}

func (s *StoreRedis) SetNX(ctx context.Context, key, value string, ttl time.Duration) bool {
	if ctx.Err() != nil {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	if item, exists := s.data[key]; exists && !item.IsExpired(now) {
		return false
	}

	s.data[key] = &domain.Item{Value: value, ExpiresAt: expiresAt(ttl)}
	return true
}

func (s *StoreRedis) Get(ctx context.Context, key string) (string, bool) {
	if ctx.Err() != nil {
		return "", false
	}

	s.mu.RLock()
	item, exists := s.data[key]
	if !exists {
		s.mu.RUnlock()
		return "", false
	}
	if !item.IsExpired(time.Now().Unix()) {
		value := item.Value
		s.mu.RUnlock()
		return value, true
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if item, exists := s.data[key]; exists && item.IsExpired(time.Now().Unix()) {
		delete(s.data, key)
	}
	return "", false
}

func (s *StoreRedis) GetDel(ctx context.Context, key string) (string, bool) {
	if ctx.Err() != nil {
		return "", false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return "", false
	}
	if item.IsExpired(time.Now().Unix()) {
		delete(s.data, key)
		return "", false
	}

	delete(s.data, key)
	return item.Value, true
}

func (s *StoreRedis) Del(ctx context.Context, key string) int {
	if ctx.Err() != nil {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; exists {
		delete(s.data, key)
		return 1
	}
	return 0
}

func (s *StoreRedis) Expire(ctx context.Context, key string, seconds int) bool {
	if ctx.Err() != nil {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}
	if item.IsExpired(time.Now().Unix()) {
		delete(s.data, key)
		return false
	}
	if seconds <= 0 {
		delete(s.data, key)
		return true
	}

	expiresAt := time.Now().Unix() + int64(seconds)
	item.ExpiresAt = &expiresAt
	return true
}

func (s *StoreRedis) TTL(ctx context.Context, key string) int64 {
	if ctx.Err() != nil {
		return -1
	}
	s.mu.RLock()

	item, exists := s.data[key]
	if !exists {
		s.mu.RUnlock()
		return -1
	}

	if item.ExpiresAt == nil {
		s.mu.RUnlock()
		return -1
	}

	now := time.Now().Unix()
	remaining := *item.ExpiresAt - now
	s.mu.RUnlock()

	if remaining <= 0 {
		s.mu.Lock()
		if item, exists := s.data[key]; exists && item.IsExpired(time.Now().Unix()) {
			delete(s.data, key)
		}
		s.mu.Unlock()
		return -1
	}
	return remaining
}

func (s *StoreRedis) Persist(ctx context.Context, key string) bool {
	if ctx.Err() != nil {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}
	if item.IsExpired(time.Now().Unix()) {
		delete(s.data, key)
		return false
	}

	item.ExpiresAt = nil
	return true
}

func (s *StoreRedis) Keys(ctx context.Context, pattern string) []string {
	if ctx.Err() != nil {
		return []string{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().Unix()
	var matches []string

	for key, item := range s.data {
		if item.IsExpired(now) {
			continue
		}

		if matchPattern(key, pattern) {
			matches = append(matches, key)
		}
	}
	return matches
}

func (s *StoreRedis) Exists(ctx context.Context, key string) bool {
	if ctx.Err() != nil {
		return false
	}
	s.mu.RLock()
	item, exists := s.data[key]
	if !exists {
		s.mu.RUnlock()
		return false
	}

	if !item.IsExpired(time.Now().Unix()) {
		s.mu.RUnlock()
		return true
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if item, exists := s.data[key]; exists && item.IsExpired(time.Now().Unix()) {
		delete(s.data, key)
	}
	return false
}

func (s *StoreRedis) Incr(ctx context.Context, key string, ttl time.Duration) (int64, int64, bool) {
	if ctx.Err() != nil {
		return 0, -1, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	item, exists := s.data[key]
	if !exists || item.IsExpired(now) {
		s.data[key] = &domain.Item{Value: "1", ExpiresAt: expiresAt(ttl)}
		return 1, ttlSeconds(ttl), true
	}

	current, err := strconv.ParseInt(item.Value, 10, 64)
	if err != nil {
		return 0, -1, false
	}
	current++
	item.Value = strconv.FormatInt(current, 10)

	return current, remainingTTL(item, now), true
}

func (s *StoreRedis) AcquireLock(ctx context.Context, key, owner string, ttl time.Duration) bool {
	if ttl <= 0 {
		return false
	}
	return s.SetNX(ctx, key, owner, ttl)
}

func (s *StoreRedis) ReleaseLock(ctx context.Context, key, owner string) bool {
	if ctx.Err() != nil {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}
	if item.IsExpired(time.Now().Unix()) {
		delete(s.data, key)
		return false
	}
	if item.Value != owner {
		return false
	}

	delete(s.data, key)
	return true
}

func (s *StoreRedis) Touch(ctx context.Context, key string, ttl time.Duration) bool {
	if ctx.Err() != nil {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}
	if item.IsExpired(time.Now().Unix()) {
		delete(s.data, key)
		return false
	}
	item.ExpiresAt = expiresAt(ttl)
	return true
}

func (s *StoreRedis) Size(ctx context.Context) int {
	if ctx.Err() != nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().Unix()
	count := 0
	for _, item := range s.data {
		if !item.IsExpired(now) {
			count++
		}
	}
	return count
}

func (s *StoreRedis) StartCleanup(intervalMs int64) {
	if intervalMs <= 0 {
		intervalMs = 1000
	}
	interval := time.Duration(intervalMs) * time.Millisecond
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanupExpired()
			case <-s.stopCleanup:
				return
			}
		}
	}()
}

func (s *StoreRedis) StopCleanup() {
	s.stopOnce.Do(func() {
		close(s.stopCleanup)
	})
}

func (s *StoreRedis) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	for key, item := range s.data {
		if item.IsExpired(now) {
			delete(s.data, key)
		}
	}
}

func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}

	matched, err := filepath.Match(pattern, key)
	if err != nil {
		return key == pattern
	}

	return matched
}

func expiresAt(ttl time.Duration) *int64 {
	if ttl <= 0 {
		return nil
	}
	expiresAt := time.Now().Add(ttl).Unix()
	return &expiresAt
}

func ttlSeconds(ttl time.Duration) int64 {
	if ttl <= 0 {
		return -1
	}
	seconds := int64(ttl.Seconds())
	if seconds <= 0 {
		return 1
	}
	return seconds
}

func remainingTTL(item *domain.Item, now int64) int64 {
	if item.ExpiresAt == nil {
		return -1
	}
	ttl := *item.ExpiresAt - now
	if ttl <= 0 {
		return -1
	}
	return ttl
}
