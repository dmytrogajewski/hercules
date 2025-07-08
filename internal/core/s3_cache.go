package core

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Cache implements CacheBackend using AWS S3
type S3Cache struct {
	client     *s3.Client
	bucket     string
	prefix     string
	defaultTTL time.Duration
}

// NewS3Cache creates a new S3 cache backend
func NewS3Cache(cfg CacheConfig) (*S3Cache, error) {
	if cfg.S3Bucket == "" {
		return nil, fmt.Errorf("s3_bucket is required for S3 cache backend")
	}

	// Build AWS config
	awsCfg, err := buildAWSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg)

	// Test bucket access
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.S3Bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to access S3 bucket %s: %w", cfg.S3Bucket, err)
	}

	return &S3Cache{
		client:     client,
		bucket:     cfg.S3Bucket,
		prefix:     strings.Trim(cfg.S3Prefix, "/"),
		defaultTTL: cfg.DefaultTTL,
	}, nil
}

// buildAWSConfig creates AWS configuration with optional credentials
func buildAWSConfig(cfg CacheConfig) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	// Set region
	if cfg.S3Region != "" {
		opts = append(opts, config.WithRegion(cfg.S3Region))
	}

	// Set custom endpoint (for MinIO, etc.)
	if cfg.S3Endpoint != "" {
		opts = append(opts, config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               cfg.S3Endpoint,
					HostnameImmutable: true,
					PartitionID:       "aws",
				}, nil
			}),
		))
	}

	// Set credentials if provided
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     cfg.AWSAccessKeyID,
				SecretAccessKey: cfg.AWSSecretAccessKey,
			},
		}))
	}

	return config.LoadDefaultConfig(context.Background(), opts...)
}

// makeKey creates a full S3 key with prefix
func (s *S3Cache) makeKey(key string) string {
	if s.prefix == "" {
		return key
	}
	return path.Join(s.prefix, key)
}

// Get retrieves a value from S3 cache
func (s *S3Cache) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.makeKey(key)),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// Set stores a value in S3 cache
func (s *S3Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.makeKey(key)),
		Body:   strings.NewReader(string(value)),
		Metadata: map[string]string{
			"ttl": fmt.Sprintf("%d", time.Now().Add(ttl).Unix()),
		},
	})
	return err
}

// Delete removes a value from S3 cache
func (s *S3Cache) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.makeKey(key)),
	})
	return err
}

// Exists checks if a key exists in S3 cache
func (s *S3Cache) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.makeKey(key)),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetReader returns an io.ReadCloser for streaming large values
func (s *S3Cache) GetReader(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.makeKey(key)),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

// SetReader stores data from an io.Reader with optional TTL
func (s *S3Cache) SetReader(ctx context.Context, key string, reader io.Reader, ttl time.Duration) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.makeKey(key)),
		Body:   reader,
		Metadata: map[string]string{
			"ttl": fmt.Sprintf("%d", time.Now().Add(ttl).Unix()),
		},
	})
	return err
}

// Close performs any cleanup operations
func (s *S3Cache) Close() error {
	// S3 client doesn't need explicit cleanup
	return nil
}

// GenerateCacheKey creates a deterministic cache key for repository analysis
func GenerateCacheKey(repoURL, branch, commit string, analysisType string) string {
	// Create a hash of the repository URL and branch for consistent keys
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", repoURL, branch, commit)))
	return fmt.Sprintf("%s/%s/%s", analysisType, hex.EncodeToString(hash[:]), commit)
}
