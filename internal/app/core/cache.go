package core

import (
	"context"
	"io"
	"time"
)

// CacheBackend defines the interface for different cache implementations
type CacheBackend interface {
	// Get retrieves a value from cache by key
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with optional TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// GetReader returns an io.ReadCloser for streaming large values
	GetReader(ctx context.Context, key string) (io.ReadCloser, error)

	// SetReader stores data from an io.Reader with optional TTL
	SetReader(ctx context.Context, key string, reader io.Reader, ttl time.Duration) error

	// Close performs any cleanup operations
	Close() error
}

// CacheConfig holds configuration for cache backends
type CacheConfig struct {
	// Backend type: "local", "s3", "memory"
	Backend string `mapstructure:"backend" yaml:"backend"`

	// Local cache settings
	LocalPath string `mapstructure:"local_path" yaml:"local_path"`

	// S3 cache settings
	S3Bucket   string `mapstructure:"s3_bucket" yaml:"s3_bucket"`
	S3Region   string `mapstructure:"s3_region" yaml:"s3_region"`
	S3Endpoint string `mapstructure:"s3_endpoint" yaml:"s3_endpoint"`
	S3Prefix   string `mapstructure:"s3_prefix" yaml:"s3_prefix"`

	// AWS credentials (optional, can use IAM roles)
	AWSAccessKeyID     string `mapstructure:"aws_access_key_id" yaml:"aws_access_key_id"`
	AWSSecretAccessKey string `mapstructure:"aws_secret_access_key" yaml:"aws_secret_access_key"`

	// Cache settings
	DefaultTTL time.Duration `mapstructure:"default_ttl" yaml:"default_ttl"`
	MaxSize    int64         `mapstructure:"max_size" yaml:"max_size"` // in bytes
}
