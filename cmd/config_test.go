package cmd

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear any existing environment variables
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("PORT")
	
	cfg := LoadConfig()
	
	if cfg.DatabaseURL != "" {
		t.Errorf("Expected empty DATABASE_URL by default, got '%s'", cfg.DatabaseURL)
	}
	
	if cfg.Port != "8080" {
		t.Errorf("Expected default port '8080', got '%s'", cfg.Port)
	}
}

func TestLoadConfig_FromEnvironment(t *testing.T) {
	// Set environment variables
	testDBURL := "postgres://test:test@localhost/testdb"
	testPort := "3000"
	
	os.Setenv("DATABASE_URL", testDBURL)
	os.Setenv("PORT", testPort)
	
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("PORT")
	}()
	
	cfg := LoadConfig()
	
	if cfg.DatabaseURL != testDBURL {
		t.Errorf("Expected DATABASE_URL '%s', got '%s'", testDBURL, cfg.DatabaseURL)
	}
	
	if cfg.Port != testPort {
		t.Errorf("Expected PORT '%s', got '%s'", testPort, cfg.Port)
	}
}

func TestLoadConfig_PartialEnvironment(t *testing.T) {
	// Set only one environment variable
	testDBURL := "postgres://test:test@localhost/testdb"
	
	os.Setenv("DATABASE_URL", testDBURL)
	os.Unsetenv("PORT") // Ensure PORT is not set
	
	defer func() {
		os.Unsetenv("DATABASE_URL")
	}()
	
	cfg := LoadConfig()
	
	if cfg.DatabaseURL != testDBURL {
		t.Errorf("Expected DATABASE_URL '%s', got '%s'", testDBURL, cfg.DatabaseURL)
	}
	
	if cfg.Port != "8080" {
		t.Errorf("Expected default PORT '8080', got '%s'", cfg.Port)
	}
}