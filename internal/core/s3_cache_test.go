package core

import (
	"testing"
	"time"
)

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name         string
		repoURL      string
		branch       string
		commit       string
		analysisType string
		expected     string
	}{
		{
			name:         "basic cache key",
			repoURL:      "https://github.com/user/repo",
			branch:       "main",
			commit:       "abc123",
			analysisType: "burndown",
			expected:     "burndown/",
		},
		{
			name:         "multiple analyses",
			repoURL:      "https://github.com/user/repo",
			branch:       "develop",
			commit:       "def456",
			analysisType: "burndown,couples,devs",
			expected:     "burndown,couples,devs/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateCacheKey(tt.repoURL, tt.branch, tt.commit, tt.analysisType)

			// Check that the result starts with the expected analysis type
			if !contains(result, tt.expected) {
				t.Errorf("GenerateCacheKey() = %v, expected to contain %v", result, tt.expected)
			}

			// Check that the result contains the commit hash
			if !contains(result, tt.commit) {
				t.Errorf("GenerateCacheKey() = %v, expected to contain commit %v", result, tt.commit)
			}
		})
	}
}

func TestCacheConfig(t *testing.T) {
	cfg := CacheConfig{
		Backend:    "s3",
		S3Bucket:   "test-bucket",
		S3Region:   "us-east-1",
		S3Prefix:   "hercules",
		DefaultTTL: 24 * time.Hour,
	}

	if cfg.Backend != "s3" {
		t.Errorf("Expected backend s3, got %s", cfg.Backend)
	}

	if cfg.S3Bucket != "test-bucket" {
		t.Errorf("Expected bucket test-bucket, got %s", cfg.S3Bucket)
	}
}

func TestNewCacheBackend(t *testing.T) {
	tests := []struct {
		name    string
		cfg     CacheConfig
		wantErr bool
	}{
		{
			name: "local cache",
			cfg: CacheConfig{
				Backend:   "local",
				LocalPath: "/tmp/test-cache",
			},
			wantErr: false,
		},
		{
			name: "memory cache",
			cfg: CacheConfig{
				Backend: "memory",
			},
			wantErr: false,
		},
		{
			name: "invalid backend",
			cfg: CacheConfig{
				Backend: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCacheBackend(tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewCacheBackend() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewCacheBackend() unexpected error: %v", err)
				return
			}

			if cache == nil {
				t.Errorf("NewCacheBackend() returned nil cache")
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
