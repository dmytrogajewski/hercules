package uast

import (
	"testing"
)

type mockProvider struct {
	lang      string
	parseErr  error
	parseNode *Node
}

func (m *mockProvider) Parse(filename string, content []byte) (*Node, error) {
	return m.parseNode, m.parseErr
}
func (m *mockProvider) Language() string { return m.lang }

func TestNewParser_CreatesProviders(t *testing.T) {
	cfg := &ProviderConfig{
		Language:   "go",
		Extensions: []string{".go"},
		Parser:     string(ParserTreeSitter),
		Mapping:    map[string]Mapping{"foo": {Type: "Bar"}},
	}
	providers := Providers{"go": cfg}
	// Use NewParser and inject configs for test
	p := &Parser{
		providers: make(map[string]LanguageProvider),
		configs:   providers,
	}
	for language, providerConfig := range providers {
		provider := NewProviderForLanguage(language, providerConfig)
		if provider != nil {
			p.providers[language] = provider
		}
	}
	langs := p.SupportedLanguages()
	if len(langs) == 0 {
		t.Errorf("expected at least one language, got 0")
	}
}

func TestParser_Parse(t *testing.T) {
	p := &Parser{
		providers: map[string]LanguageProvider{
			"go": &mockProvider{lang: "go", parseNode: &Node{Type: "Root"}},
		},
		configs: Providers{"go": &ProviderConfig{Language: "go", Extensions: []string{".go"}}},
	}
	node, err := p.Parse("foo.go", []byte(""))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if node == nil || node.Type != "Root" {
		t.Errorf("expected Root node, got %+v", node)
	}
	_, err = p.Parse("", []byte(""))
	if err == nil {
		t.Errorf("expected error for empty filename")
	}
	_, err = p.Parse("foo.py", []byte(""))
	if err == nil {
		t.Errorf("expected error for unsupported language")
	}
}

func TestParser_detectLanguage(t *testing.T) {
	cfg := &ProviderConfig{Language: "go", Extensions: []string{".go"}}
	p := &Parser{configs: Providers{"go": cfg}}
	lang := p.detectLanguage("foo.go")
	if lang != "go" {
		t.Errorf("expected go, got %s", lang)
	}
	lang = p.detectLanguage("foo.py")
	if lang != "" {
		t.Errorf("expected empty, got %s", lang)
	}
}

func TestParser_SupportedLanguages(t *testing.T) {
	p := &Parser{providers: map[string]LanguageProvider{"go": &mockProvider{lang: "go"}}}
	langs := p.SupportedLanguages()
	if len(langs) != 1 || langs[0] != "go" {
		t.Errorf("expected [go], got %v", langs)
	}
}

func TestParser_IsSupported(t *testing.T) {
	cfg := &ProviderConfig{Language: "go", Extensions: []string{".go"}}
	p := &Parser{
		providers: map[string]LanguageProvider{"go": &mockProvider{lang: "go"}},
		configs:   Providers{"go": cfg},
	}
	if !p.IsSupported("foo.go") {
		t.Errorf("expected true for .go")
	}
	if p.IsSupported("foo.py") {
		t.Errorf("expected false for .py")
	}
}
