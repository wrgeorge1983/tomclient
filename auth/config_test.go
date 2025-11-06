package auth

import (
	"os"
	"testing"
)

func TestCacheDefaults(t *testing.T) {
	// Test default cache settings
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !cfg.CacheEnabled {
		t.Errorf("Expected CacheEnabled to be true by default, got false")
	}

	if cfg.CacheTTL != 300 {
		t.Errorf("Expected CacheTTL to be 300 by default, got %d", cfg.CacheTTL)
	}
}

func TestCacheEnvironmentVariables(t *testing.T) {
	// Test TOM_CACHE_ENABLED environment variable
	os.Setenv("TOM_CACHE_ENABLED", "false")
	defer os.Unsetenv("TOM_CACHE_ENABLED")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.CacheEnabled {
		t.Errorf("Expected CacheEnabled to be false from env var, got true")
	}

	// Test TOM_CACHE_TTL environment variable
	os.Setenv("TOM_CACHE_TTL", "600")
	defer os.Unsetenv("TOM_CACHE_TTL")

	cfg2, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg2.CacheTTL != 600 {
		t.Errorf("Expected CacheTTL to be 600 from env var, got %d", cfg2.CacheTTL)
	}
}
