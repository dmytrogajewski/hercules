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

type Providers map[string]*ProviderConfig

func LoadProviders() (Providers, error) {
	providers := Providers{}
	entries, err := readProviderEntries()
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if shouldSkipEntry(entry) {
			continue
		}
		cfg, err := loadProviderConfig(entry)
		if err != nil {
			return nil, err
		}
		if isDuplicateProvider(providers, cfg.Language) {
			return nil, createDuplicateProviderError(cfg.Language)
		}
		providers[cfg.Language] = cfg
	}
	return providers, nil
}

func readProviderEntries() ([]fs.DirEntry, error) {
	return fs.ReadDir(providerFS, "providers")
}

func shouldSkipEntry(entry fs.DirEntry) bool {
	return entry.IsDir() || !isYamlFile(entry.Name())
}

func isYamlFile(name string) bool {
	return filepath.Ext(name) == ".yaml"
}

func loadProviderConfig(entry fs.DirEntry) (*ProviderConfig, error) {
	data, err := readProviderFile(entry.Name())
	if err != nil {
		return nil, err
	}
	cfg, err := parseProviderConfig(data)
	if err != nil {
		return nil, err
	}
	if isMissingLanguage(cfg) {
		return nil, createMissingLanguageError(entry.Name())
	}
	if isMissingParser(cfg) {
		return nil, createMissingParserError(entry.Name())
	}
	if hasInvalidMapping(cfg) {
		return nil, createInvalidMappingError(cfg.Language)
	}
	return cfg, nil
}

func readProviderFile(name string) ([]byte, error) {
	return providerFS.ReadFile("providers/" + name)
}

func parseProviderConfig(data []byte) (*ProviderConfig, error) {
	var cfg ProviderConfig
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func isMissingLanguage(cfg *ProviderConfig) bool {
	return cfg.Language == ""
}

func createMissingLanguageError(filename string) error {
	return errors.New("missing language in " + filename)
}

func isMissingParser(cfg *ProviderConfig) bool {
	return cfg.Parser == ""
}

func createMissingParserError(filename string) error {
	return errors.New("missing parser in " + filename)
}

func hasInvalidMapping(cfg *ProviderConfig) bool {
	for _, mapping := range cfg.Mapping {
		if isMissingMappingType(mapping) {
			return true
		}
	}
	return false
}

func isMissingMappingType(mapping Mapping) bool {
	return mapping.Type == ""
}

func createInvalidMappingError(language string) error {
	return errors.New("missing type for " + language + ":kind")
}

func isDuplicateProvider(providers Providers, language string) bool {
	_, exists := providers[language]
	return exists
}

func createDuplicateProviderError(language string) error {
	return errors.New("duplicate provider for language " + language)
}
