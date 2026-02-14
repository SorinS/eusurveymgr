package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	BaseURL        string `json:"base_url"`
	WebUser        string `json:"web_user"`
	WebPassword    string `json:"web_password"`
	DBHost         string `json:"db_host"`
	DBPort         int    `json:"db_port"`
	DBName         string `json:"db_name"`
	DBUser         string `json:"db_user"`
	DBPassword     string `json:"db_password"`
	OutputDir      string `json:"output_dir"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	InsecureTLS    bool   `json:"insecure_tls"`
}

func LoadFromFile(filePath string) (*Configuration, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var c Configuration
	err = json.Unmarshal(content, &c)
	if err != nil {
		return nil, err
	}
	if c.TimeoutSeconds == 0 {
		c.TimeoutSeconds = 30
	}
	if c.OutputDir == "" {
		c.OutputDir = "."
	}
	applyEnvOverrides(&c)
	return &c, nil
}

// applyEnvOverrides overrides config fields from EUSURVEYMGR_* environment
// variables. This avoids exposing credentials on the command line.
func applyEnvOverrides(c *Configuration) {
	if v := os.Getenv("EUSURVEYMGR_WEB_USER"); v != "" {
		c.WebUser = v
	}
	if v := os.Getenv("EUSURVEYMGR_WEB_PASSWORD"); v != "" {
		c.WebPassword = v
	}
	if v := os.Getenv("EUSURVEYMGR_DB_HOST"); v != "" {
		c.DBHost = v
	}
	if v := os.Getenv("EUSURVEYMGR_DB_NAME"); v != "" {
		c.DBName = v
	}
	if v := os.Getenv("EUSURVEYMGR_DB_USER"); v != "" {
		c.DBUser = v
	}
	if v := os.Getenv("EUSURVEYMGR_DB_PASSWORD"); v != "" {
		c.DBPassword = v
	}
}

func PrintConfig(cfg *Configuration) {
	safe := *cfg
	safe.WebPassword = "***"
	safe.DBPassword = "***"
	c, err := json.MarshalIndent(safe, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling config: %v\n", err)
		return
	}
	fmt.Printf("Configuration:\n%s\n", string(c))
}