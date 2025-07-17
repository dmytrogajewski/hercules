package uast

import (
	"testing"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader(nil)
	if loader == nil {
		t.Errorf("expected non-nil loader")
	}
}

func TestLoader_LoadProvider(t *testing.T) {
	loader := NewLoader(nil)

	// Test loading a provider (this will fail since we don't have actual embed.FS)
	_, err := loader.LoadProvider("go")
	if err == nil {
		t.Errorf("expected error when loading provider without embed.FS")
	}
}

func TestLoader_LoadAllProviders(t *testing.T) {
	loader := NewLoader(nil)

	// Test loading all providers (this will fail since we don't have actual embed.FS)
	_, err := loader.LoadAllProviders()
	if err == nil {
		t.Errorf("expected error when loading providers without embed.FS")
	}
}

func TestLoader_loadDSLMapping(t *testing.T) {
	loader := NewLoader(nil)

	// Test loading DSL mapping (this will fail since we don't have actual embed.FS)
	_, err := loader.loadDSLMapping("go")
	if err == nil {
		t.Errorf("expected error when loading DSL mapping without embed.FS")
	}
}
