package memrate

import (
	"context"
	lru "github.com/hashicorp/golang-lru"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/youminxue/v2/framework/ratelimit"
	logger "github.com/youminxue/v2/toolkit/zlogger"
	"sync"
)

const defaultMaxKeys = 256

type LimiterFn func(ctx context.Context, store *MemoryStore, key string) ratelimit.Limiter

type MemoryStore struct {
	keys      *lru.Cache
	maxKeys   int
	onEvicted simplelru.EvictCallback
	limiterFn LimiterFn
	mu        sync.RWMutex
}

type MemoryStoreOption func(*MemoryStore)

// WithMaxKeys set maxKeys
func WithMaxKeys(maxKeys int) MemoryStoreOption {
	return func(ls *MemoryStore) {
		ls.maxKeys = maxKeys
	}
}

// WithOnEvicted set onEvicted
func WithOnEvicted(onEvicted func(key interface{}, value interface{})) MemoryStoreOption {
	return func(ls *MemoryStore) {
		ls.onEvicted = onEvicted
	}
}

func NewMemoryStore(fn LimiterFn, opts ...MemoryStoreOption) *MemoryStore {
	store := &MemoryStore{
		maxKeys:   defaultMaxKeys,
		limiterFn: fn,
	}

	for _, opt := range opts {
		opt(store)
	}

	if store.onEvicted != nil {
		store.keys, _ = lru.NewWithEvict(store.maxKeys, store.onEvicted)
	} else {
		store.keys, _ = lru.New(store.maxKeys)
	}

	return store
}

// GetLimiter returns the rate limiter for the provided key if it exists,
// otherwise calls addKey to add key to the map
func (store *MemoryStore) GetLimiter(key string) ratelimit.Limiter {
	return store.GetLimiterCtx(context.Background(), key)
}

func (store *MemoryStore) addKeyCtx(ctx context.Context, key string) ratelimit.Limiter {
	store.mu.Lock()
	defer store.mu.Unlock()

	limiter, exists := store.keys.Get(key)
	if exists {
		return limiter.(ratelimit.Limiter)
	}

	limiter = store.limiterFn(ctx, store, key)
	store.keys.Add(key, limiter)

	return limiter.(ratelimit.Limiter)
}

// GetLimiterCtx returns the rate limiter for the provided key if it exists,
// otherwise calls addKey to add key to the map
func (store *MemoryStore) GetLimiterCtx(ctx context.Context, key string) ratelimit.Limiter {
	store.mu.RLock()

	limiter, exists := store.keys.Get(key)
	if !exists {
		store.mu.RUnlock()
		return store.addKeyCtx(ctx, key)
	}

	store.mu.RUnlock()
	return limiter.(ratelimit.Limiter)
}

func (store *MemoryStore) DeleteKey(key string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.keys.Remove(key)
	logger.Debug().Msgf("[go-doudou] key %s is deleted from store", key)
}
