package uast

import (
	"embed"
	"errors"
	"io/fs"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed providers/*.yaml
var providerFS embed.FS

type ProviderConfig struct {
	Language   string             `yaml:"language"`
	Extensions []string           `yaml:"extensions"`
	Parser     string             `yaml:"parser"`
	Mapping    map[string]Mapping `yaml:"mapping"`
}

type Mapping struct {
	Type        string                 `yaml:"type"`
	Roles       []string               `yaml:"roles,omitempty"`
	Props       map[string]interface{} `yaml:"props,omitempty"`
	SkipIfEmpty bool                   `yaml:"skip_if_empty,omitempty"`
}

type Providers map[string]*ProviderConfig // language -> config

// LoadProviders loads all provider YAMLs from the embedded FS (pkg/uast/providers/).
func LoadProviders() (Providers, error) {
	providers := Providers{}
	entries, err := fs.ReadDir(providerFS, "providers")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		data, err := providerFS.ReadFile("providers/" + entry.Name())
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
