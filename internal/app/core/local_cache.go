package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalCache implements CacheBackend using local filesystem
type LocalCache struct {
	basePath   string
	defaultTTL time.Duration
}

// NewLocalCache creates a new local filesystem cache backend
func NewLocalCache(cfg CacheConfig) (*LocalCache, error) {
	path := cfg.LocalPath
	if path == "" {
		path = "/tmp/hercules-cache"
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory %s: %w", path, err)
	}

	return &LocalCache{
		basePath:   path,
		defaultTTL: cfg.DefaultTTL,
	}, nil
}

// makePath creates a filesystem path for a cache key
func (l *LocalCache) makePath(key string) string {
	// Sanitize key for filesystem
	safeKey := strings.ReplaceAll(key, "/", "_")
	safeKey = strings.ReplaceAll(safeKey, "\\", "_")
	return filepath.Join(l.basePath, safeKey)
}

// Get retrieves a value from local cache
func (l *LocalCache) Get(ctx context.Context, key string) ([]byte, error) {
	path := l.makePath(key)

	// Check if file exists and is not expired
	if exists, err := l.Exists(ctx, key); err != nil || !exists {
		return nil, fmt.Errorf("cache key not found or expired")
	}

	return os.ReadFile(path)
}

// Set stores a value in local cache
func (l *LocalCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	path := l.makePath(key)

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write file with TTL metadata
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	// Write TTL header
	expiry := time.Now().Add(ttl).Unix()
	header := fmt.Sprintf("TTL:%d\n", expiry)
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write TTL header: %w", err)
	}

	// Write data
	if _, err := file.Write(value); err != nil {
		return fmt.Errorf("failed to write cache data: %w", err)
	}

	return nil
}

// Delete removes a value from local cache
func (l *LocalCache) Delete(ctx context.Context, key string) error {
	path := l.makePath(key)
	return os.Remove(path)
}

// Exists checks if a key exists in local cache and is not expired
func (l *LocalCache) Exists(ctx context.Context, key string) (bool, error) {
	path := l.makePath(key)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	// Check if file is expired
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read TTL header
	var ttl int64
	if _, err := fmt.Fscanf(file, "TTL:%d\n", &ttl); err != nil {
		// If no TTL header, assume expired
		return false, nil
	}

	// Check if expired
	if time.Now().Unix() > ttl {
		// Remove expired file
		os.Remove(path)
		return false, nil
	}

	return true, nil
}

// GetReader returns an io.ReadCloser for streaming large values
func (l *LocalCache) GetReader(ctx context.Context, key string) (io.ReadCloser, error) {
	path := l.makePath(key)

	// Check if file exists and is not expired
	if exists, err := l.Exists(ctx, key); err != nil || !exists {
		return nil, fmt.Errorf("cache key not found or expired")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Skip TTL header
	var ttl int64
	if _, err := fmt.Fscanf(file, "TTL:%d\n", &ttl); err != nil {
		file.Close()
		return nil, fmt.Errorf("invalid cache file format")
	}

	return file, nil
}

// SetReader stores data from an io.Reader with optional TTL
func (l *LocalCache) SetReader(ctx context.Context, key string, reader io.Reader, ttl time.Duration) error {
	path := l.makePath(key)

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	// Write TTL header
	expiry := time.Now().Add(ttl).Unix()
	header := fmt.Sprintf("TTL:%d\n", expiry)
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write TTL header: %w", err)
	}

	// Copy data from reader
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write cache data: %w", err)
	}

	return nil
}

// Close performs any cleanup operations
func (l *LocalCache) Close() error {
	// No cleanup needed for local cache
	return nil
}

// Cleanup removes expired cache entries
func (l *LocalCache) Cleanup() error {
	return filepath.Walk(l.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file is expired
		file, err := os.Open(path)
		if err != nil {
			return nil // Skip files we can't read
		}
		defer file.Close()

		var ttl int64
		if _, err := fmt.Fscanf(file, "TTL:%d\n", &ttl); err != nil {
			// Remove files without TTL header
			os.Remove(path)
			return nil
		}

		if time.Now().Unix() > ttl {
			os.Remove(path)
		}

		return nil
	})
}
