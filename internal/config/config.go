package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrinjamul/flareship/internal/utils"
	"github.com/mrinjamul/flareship/pkg/schema"
)

const DefaultConfigFile = "flareship.json"

// InitConfig writes the provided AppConfig to ./flareship.json after validation.
// Returns error if validation fails or file write fails.
// Skips writing if file already exists.
func InitConfig(cfg *schema.AppConfig) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	path := filepath.Join(cwd, DefaultConfigFile)

	// Skip writing if file already exists
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Config file already exists at '%s'. Skipping write.\n", path)
		return nil
	} else if !os.IsNotExist(err) {
		// An unexpected error occurred when checking file existence
		return fmt.Errorf("failed to check if config file exists: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file at %s: %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config to file: %w", err)
	}

	fmt.Printf("Config successfully written to '%s'\n", path)
	return nil
}

// LoadConfig loads config from the given path or default if empty
func LoadConfig(path string) (*schema.AppConfig, error) {
	var config schema.AppConfig = schema.AppConfig{}

	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		path = filepath.Join(cwd, DefaultConfigFile)
	}

	_, present := os.LookupEnv("FLARESHIP_DOMAINS")
	if present {
		config, err := loadFromEnv()
		if err == nil {
			return config, nil
		}
		fmt.Println(err)
	}

	homeDir := utils.HomeDir()
	configPath := filepath.Join(homeDir, ".config", "flareship.json")

	if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
		path = configPath // File exists and is not a directory
	}

	bytes, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("warning: no config found!")
		return &config, nil
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("invalid config format: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// loadFromEnv supports environment variable configuration
func loadFromEnv() (*schema.AppConfig, error) {
	// Example:
	// FLARESHIP_DOMAINS="example.com,myapp.io"
	// FLARESHIP_CF_TOKENS="token1,token2"
	// FLARESHIP_ZONE_IDS="zone1,zone2"
	// FLARESHIP_RECORD_FILES="records.json,"
	// FLARESHIP_RESTRICTED_FILES="restricted.json,"
	// FLARESHIP_ALLOWED_TYPES="A,CNAME;A,CNAME"

	domainNames := strings.Split(os.Getenv("FLARESHIP_DOMAINS"), ",")
	tokens := strings.Split(os.Getenv("FLARESHIP_CF_TOKENS"), ",")
	zones := strings.Split(os.Getenv("FLARESHIP_ZONE_IDS"), ",")
	recordFiles := strings.Split(os.Getenv("FLARESHIP_RECORD_FILES"), ",")
	restrictedFiles := strings.Split(os.Getenv("FLARESHIP_RESTRICTED_FILES"), ",")
	allowedTypesRaw := os.Getenv("FLARESHIP_ALLOWED_TYPES")

	// Ensure none are empty
	if os.Getenv("FLARESHIP_DOMAINS") == "" ||
		os.Getenv("FLARESHIP_CF_TOKENS") == "" ||
		os.Getenv("FLARESHIP_ZONE_IDS") == "" ||
		os.Getenv("FLARESHIP_RECORD_FILES") == "" ||
		os.Getenv("FLARESHIP_RESTRICTED_FILES") == "" ||
		allowedTypesRaw == "" {
		return nil, errors.New("one or more required FLARESHIP_* environment variables are missing or empty")
	}

	if len(domainNames) == 0 || domainNames[0] == "" {
		return nil, errors.New("no config file found and FLARESHIP_DOMAINS is empty")
	}

	var allowedTypesList [][]string
	if allowedTypesRaw != "" {
		// Split by `;` → per domain
		perDomain := strings.Split(allowedTypesRaw, ";")
		if len(perDomain) != len(domainNames) {
			return nil, errors.New("mismatch in number of domains and allowed types entries")
		}
		for _, domainTypes := range perDomain {
			types := strings.Split(domainTypes, ",")
			for i := range types {
				types[i] = strings.TrimSpace(types[i])
			}
			allowedTypesList = append(allowedTypesList, types)
		}
	}

	if len(domainNames) != len(tokens) ||
		len(domainNames) != len(zones) ||
		len(domainNames) != len(recordFiles) ||
		len(domainNames) != len(restrictedFiles) {
		return nil, errors.New("mismatch in number of domains vs tokens, zones, files, or allowed types")
	}

	cfg := schema.AppConfig{}
	for i := range domainNames {
		d := schema.DomainConfig{
			CFToken: strings.TrimSpace(tokens[i]),
			ZoneID:  strings.TrimSpace(zones[i]),
			Name:    strings.TrimSpace(domainNames[i]),
		}
		if i < len(recordFiles) && strings.TrimSpace(recordFiles[i]) != "" {
			d.RecordFile = strings.TrimSpace(recordFiles[i])
		}
		if i < len(restrictedFiles) && strings.TrimSpace(restrictedFiles[i]) != "" {
			d.RestrictedFile = strings.TrimSpace(restrictedFiles[i])
		}
		if i < len(allowedTypesList) {
			d.RecordTypes = allowedTypesList[i]
		}
		cfg.Domains = append(cfg.Domains, d)
	}

	return &cfg, nil
}
