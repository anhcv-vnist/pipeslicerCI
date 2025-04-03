package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ConfigValue represents a configuration value
type ConfigValue struct {
	ID          int64
	Service     string
	Environment string
	Key         string
	Value       string
	IsSecret    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ConfigManager manages configuration values
type ConfigManager struct {
	db *sql.DB
}

// NewConfigManager creates a new ConfigManager instance
func NewConfigManager(dbPath string) (*ConfigManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS config_values (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			service TEXT NOT NULL,
			environment TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			is_secret BOOLEAN NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			UNIQUE(service, environment, key)
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &ConfigManager{db: db}, nil
}

// Close closes the database connection
func (m *ConfigManager) Close() error {
	return m.db.Close()
}

// SetValue sets a configuration value
func (m *ConfigManager) SetValue(ctx context.Context, service, environment, key, value string, isSecret bool) error {
	now := time.Now()
	_, err := m.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO config_values (service, environment, key, value, is_secret, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, service, environment, key, value, isSecret, now, now)
	
	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	
	return nil
}

// GetValue gets a configuration value
func (m *ConfigManager) GetValue(ctx context.Context, service, environment, key string) (*ConfigValue, error) {
	row := m.db.QueryRowContext(ctx, `
		SELECT id, service, environment, key, value, is_secret, created_at, updated_at
		FROM config_values
		WHERE service = ? AND environment = ? AND key = ?
		LIMIT 1
	`, service, environment, key)

	var value ConfigValue
	err := row.Scan(&value.ID, &value.Service, &value.Environment, &value.Key, &value.Value, &value.IsSecret, &value.CreatedAt, &value.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no value found for %s/%s/%s", service, environment, key)
		}
		return nil, fmt.Errorf("failed to get value: %w", err)
	}

	return &value, nil
}

// DeleteValue deletes a configuration value
func (m *ConfigManager) DeleteValue(ctx context.Context, service, environment, key string) error {
	_, err := m.db.ExecContext(ctx, `
		DELETE FROM config_values
		WHERE service = ? AND environment = ? AND key = ?
	`, service, environment, key)
	
	if err != nil {
		return fmt.Errorf("failed to delete value: %w", err)
	}
	
	return nil
}

// GetServiceConfig gets all configuration values for a service and environment
func (m *ConfigManager) GetServiceConfig(ctx context.Context, service, environment string) (map[string]string, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT key, value
		FROM config_values
		WHERE service = ? AND environment = ? AND is_secret = 0
		ORDER BY key
	`, service, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get service config: %w", err)
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config row: %w", err)
		}
		config[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating config rows: %w", err)
	}

	return config, nil
}

// GetServiceSecrets gets all secret values for a service and environment
func (m *ConfigManager) GetServiceSecrets(ctx context.Context, service, environment string) (map[string]string, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT key, value
		FROM config_values
		WHERE service = ? AND environment = ? AND is_secret = 1
		ORDER BY key
	`, service, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get service secrets: %w", err)
	}
	defer rows.Close()

	secrets := make(map[string]string)
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret row: %w", err)
		}
		secrets[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating secret rows: %w", err)
	}

	return secrets, nil
}

// GetEnvironments gets all environments
func (m *ConfigManager) GetEnvironments(ctx context.Context) ([]string, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT DISTINCT environment
		FROM config_values
		ORDER BY environment
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get environments: %w", err)
	}
	defer rows.Close()

	var environments []string
	for rows.Next() {
		var environment string
		err := rows.Scan(&environment)
		if err != nil {
			return nil, fmt.Errorf("failed to scan environment row: %w", err)
		}
		environments = append(environments, environment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environment rows: %w", err)
	}

	return environments, nil
}

// GetServices gets all services
func (m *ConfigManager) GetServices(ctx context.Context) ([]string, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT DISTINCT service
		FROM config_values
		ORDER BY service
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		err := rows.Scan(&service)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service row: %w", err)
		}
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating service rows: %w", err)
	}

	return services, nil
}

// ImportConfig imports configuration values from a JSON string
func (m *ConfigManager) ImportConfig(ctx context.Context, jsonStr string) error {
	var config map[string]map[string]map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &config)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for service, envs := range config {
		for env, values := range envs {
			for key, value := range values {
				// Check if the value is a secret
				isSecret := false
				if valueMap, ok := value.(map[string]interface{}); ok {
					if secretVal, ok := valueMap["value"]; ok {
						if secretBool, ok := valueMap["isSecret"].(bool); ok && secretBool {
							value = secretVal
							isSecret = true
						}
					}
				}

				// Convert value to string
				var strValue string
				switch v := value.(type) {
				case string:
					strValue = v
				case float64:
					strValue = fmt.Sprintf("%g", v)
				case bool:
					strValue = fmt.Sprintf("%t", v)
				case nil:
					strValue = ""
				default:
					// Try to marshal complex values to JSON
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return fmt.Errorf("failed to marshal value for %s/%s/%s: %w", service, env, key, err)
					}
					strValue = string(jsonBytes)
				}

				// Insert or replace the value
				now := time.Now()
				_, err := tx.ExecContext(ctx, `
					INSERT OR REPLACE INTO config_values (service, environment, key, value, is_secret, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?)
				`, service, env, key, strValue, isSecret, now, now)
				
				if err != nil {
					return fmt.Errorf("failed to import value for %s/%s/%s: %w", service, env, key, err)
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExportConfig exports configuration values to a JSON string
func (m *ConfigManager) ExportConfig(ctx context.Context) (string, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT service, environment, key, value, is_secret
		FROM config_values
		ORDER BY service, environment, key
	`)
	if err != nil {
		return "", fmt.Errorf("failed to export config: %w", err)
	}
	defer rows.Close()

	config := make(map[string]map[string]map[string]interface{})
	for rows.Next() {
		var service, environment, key, value string
		var isSecret bool
		err := rows.Scan(&service, &environment, &key, &value, &isSecret)
		if err != nil {
			return "", fmt.Errorf("failed to scan config row: %w", err)
		}

		// Initialize maps if they don't exist
		if _, ok := config[service]; !ok {
			config[service] = make(map[string]map[string]interface{})
		}
		if _, ok := config[service][environment]; !ok {
			config[service][environment] = make(map[string]interface{})
		}

		// Store value or secret
		if isSecret {
			config[service][environment][key] = map[string]interface{}{
				"value":    value,
				"isSecret": true,
			}
		} else {
			config[service][environment][key] = value
		}
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating config rows: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// GenerateEnvFile generates a .env file for a service and environment
func (m *ConfigManager) GenerateEnvFile(ctx context.Context, service, environment string) (string, error) {
	// Get regular config values
	config, err := m.GetServiceConfig(ctx, service, environment)
	if err != nil {
		return "", fmt.Errorf("failed to get service config: %w", err)
	}

	// Get secret values
	secrets, err := m.GetServiceSecrets(ctx, service, environment)
	if err != nil {
		return "", fmt.Errorf("failed to get service secrets: %w", err)
	}

	// Combine config and secrets
	allConfig := make(map[string]string)
	for k, v := range config {
		allConfig[k] = v
	}
	for k, v := range secrets {
		allConfig[k] = v
	}

	// Generate .env file content
	var envContent string
	for k, v := range allConfig {
		envContent += fmt.Sprintf("%s=%s\n", k, v)
	}

	return envContent, nil
}
