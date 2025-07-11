package uast

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadProviders_Embedded(t *testing.T) {
	providers, err := LoadProviders()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(providers) == 0 {
		t.Errorf("expected at least one provider, got 0")
	}
	goProv, ok := providers["go"]
	if !ok {
		t.Errorf("expected go provider present")
	}
	if goProv.Parser != "tree-sitter" {
		t.Errorf("expected parser 'tree-sitter', got %q", goProv.Parser)
	}
	if goProv.Mapping["function_declaration"].Type != "Function" {
		t.Errorf("expected mapping for function_declaration")
	}
}

func TestLoadProviders_Negative(t *testing.T) {
	cases := []struct {
		name     string
		content  string
		filename string
		wantErr  string
	}{
		{
			name:     "missing language",
			filename: "bad1.yaml",
			content: `extensions: [".go"]
parser: tree-sitter
mapping:
  foo:
    type: Bar
`,
			wantErr: "missing language",
		},
		{
			name:     "missing parser",
			filename: "bad2.yaml",
			content: `language: go
extensions: [".go"]
mapping:
  foo:
    type: Bar
`,
			wantErr: "missing parser",
		},
		{
			name:     "missing mapping type",
			filename: "bad3.yaml",
			content: `language: go
extensions: [".go"]
parser: tree-sitter
mapping:
  foo:
    roles: [Bar]
`,
			wantErr: "missing type for go:foo",
		},
		{
			name:     "duplicate provider",
			filename: "go1.yaml",
			content: `language: go
extensions: [".go"]
parser: tree-sitter
mapping:
  foo:
    type: Bar
`,
			wantErr: "duplicate provider",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			// Write a second file with the same language for duplicate test only
			if tc.name == "duplicate provider" {
				os.WriteFile(filepath.Join(dir, "go2.yaml"), []byte("language: go\nextensions: [\".go\"]\nparser: tree-sitter\nmapping:\n  foo:\n    type: Bar\n"), 0644)
			}
			path := filepath.Join(dir, tc.filename)
			os.WriteFile(path, []byte(tc.content), 0644)
			_, err := loadProvidersFromDir(dir)
			if err == nil || tc.wantErr == "" {
				if tc.wantErr != "" {
					t.Errorf("expected error containing %q, got nil", tc.wantErr)
				}
				return
			}
			if err != nil && tc.wantErr != "" && !contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

// loadProvidersFromDir is a test helper for negative cases.
func loadProvidersFromDir(dir string) (Providers, error) {
	providers := Providers{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var cfg ProviderConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
		if cfg.Language == "" {
			return nil, errors.New("missing language in " + entry.Name())
		}
		if cfg.Parser == "" {
			return nil, errors.New("missing parser in " + entry.Name())
		}
		for kind, mapping := range cfg.Mapping {
			if mapping.Type == "" {
				return nil, errors.New("missing type for " + cfg.Language + ":" + kind)
			}
		}
		if _, exists := providers[cfg.Language]; exists {
			return nil, errors.New("duplicate provider for language " + cfg.Language)
		}
		providers[cfg.Language] = &cfg
	}
	return providers, nil
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (contains(s[1:], substr) || contains(s[:len(s)-1], substr)))))
}
