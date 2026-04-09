package utility

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName      string     `json:"app_name"`
	Port		 string     `json:"port"`
	DatabaseURL  string     `json:"database_url"`

	// Optional fields
	Debug        bool       `json:"debug,omitempty"`
	LogLevel     string     `json:"log_level,omitempty"`
}

const configFile = "config.json"

// Ensure config file exists or create a default one
func InitConfig() (*Config, error) {
	// 1. Check if file exists
	if _, err := os.Stat(configFile); err == nil {
		return loadConfig()
	}

	// 2. If not, load env file
	_ = godotenv.Load()

	// 3. Validate config values
	required := []string{
		"APP_NAME",
		"PORT",
		"DATABASE_URL",
	}

	envValues := make(map[string]string)
	for _, key := range required {
		val := os.Getenv(key)
		if val == "" {
			return  nil, fmt.Errorf("missing required env variables: %s", key)
		}
		envValues[key] = val
	}

	// 4. Create config struct
	cfg := Config{
		AppName:     envValues["APP_NAME"],
		Port:        envValues["PORT"],
		DatabaseURL: envValues["DATABASE_URL"],

		// Optional values
		Debug:       getEnvOrDefault("DEBUG", "false") == "true",
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
	}

	// 5. Save to file
	if err := saveConfig(&cfg); err != nil {
		return nil, fmt.Errorf("failed to save config: %v", err)
	}

	return &cfg, nil
}

func loadConfig() (*Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %v", err)
	}

	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	file, err := os.Create(configFile)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(cfg)
}

func getEnvOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}