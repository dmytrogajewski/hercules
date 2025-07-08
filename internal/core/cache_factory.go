package core

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// NewCacheBackend creates a cache backend based on configuration
func NewCacheBackend(cfg CacheConfig) (CacheBackend, error) {
	switch cfg.Backend {
	case "local":
		return NewLocalCache(cfg)
	case "s3":
		return NewS3Cache(cfg)
	case "memory":
		return NewMemoryCache(cfg)
	default:
		return nil, fmt.Errorf("unsupported cache backend: %s", cfg.Backend)
	}
}

// MemoryCache implements CacheBackend using in-memory storage
type MemoryCache struct {
	data       map[string][]byte
	expiry     map[string]time.Time
	defaultTTL time.Duration
}

// NewMemoryCache creates a new in-memory cache backend
func NewMemoryCache(cfg CacheConfig) (*MemoryCache, error) {
	return &MemoryCache{
		data:       make(map[string][]byte),
		expiry:     make(map[string]time.Time),
		defaultTTL: cfg.DefaultTTL,
	}, nil
}

// Get retrieves a value from memory cache
func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	if exists, err := m.Exists(ctx, key); err != nil || !exists {
		return nil, fmt.Errorf("cache key not found or expired")
	}
	return m.data[key], nil
}

// Set stores a value in memory cache
func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.data[key] = value
	m.expiry[key] = time.Now().Add(ttl)
	return nil
}

// Delete removes a value from memory cache
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	delete(m.expiry, key)
	return nil
}

// Exists checks if a key exists in memory cache and is not expired
func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	if _, exists := m.data[key]; !exists {
		return false, nil
	}

	if expiry, exists := m.expiry[key]; exists && time.Now().After(expiry) {
		// Remove expired entry
		delete(m.data, key)
		delete(m.expiry, key)
		return false, nil
	}

	return true, nil
}

// GetReader returns an io.ReadCloser for streaming large values
func (m *MemoryCache) GetReader(ctx context.Context, key string) (io.ReadCloser, error) {
	if exists, err := m.Exists(ctx, key); err != nil || !exists {
		return nil, fmt.Errorf("cache key not found or expired")
	}

	data := m.data[key]
	return io.NopCloser(strings.NewReader(string(data))), nil
}

// SetReader stores data from an io.Reader with optional TTL
func (m *MemoryCache) SetReader(ctx context.Context, key string, reader io.Reader, ttl time.Duration) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	m.data[key] = data
	m.expiry[key] = time.Now().Add(ttl)
	return nil
}

// Close performs any cleanup operations
func (m *MemoryCache) Close() error {
	// Clear memory cache
	m.data = make(map[string][]byte)
	m.expiry = make(map[string]time.Time)
	return nil
}
