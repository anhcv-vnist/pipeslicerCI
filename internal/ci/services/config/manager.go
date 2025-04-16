package config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConfigValue represents a configuration value
type ConfigValue struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	Service     string    `gorm:"not null"`
	Environment string    `gorm:"not null"`
	Key         string    `gorm:"not null"`
	Value       string    `gorm:"not null"`
	IsSecret    bool      `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// ConfigManager manages configuration values
type ConfigManager struct {
	db *gorm.DB
}

// NewConfigManager creates a new ConfigManager instance
func NewConfigManager(dbPath string) (*ConfigManager, error) {
	db, err := gorm.Open(postgres.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&ConfigValue{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &ConfigManager{db: db}, nil
}

// Close closes the database connection
func (m *ConfigManager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// SetValue sets a configuration value
func (m *ConfigManager) SetValue(ctx context.Context, service, environment, key, value string, isSecret bool) error {
	now := time.Now()
	configValue := &ConfigValue{
		Service:     service,
		Environment: environment,
		Key:         key,
		Value:       value,
		IsSecret:    isSecret,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := m.db.WithContext(ctx).
		Where(ConfigValue{
			Service:     service,
			Environment: environment,
			Key:         key,
		}).
		Assign(map[string]interface{}{
			"value":      value,
			"is_secret":  isSecret,
			"updated_at": now,
		}).
		FirstOrCreate(configValue)

	return result.Error
}

// GetValue gets a configuration value
func (m *ConfigManager) GetValue(ctx context.Context, service, environment, key string) (*ConfigValue, error) {
	var value ConfigValue
	result := m.db.WithContext(ctx).
		Where("service = ? AND environment = ? AND key = ?", service, environment, key).
		First(&value)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no value found for %s/%s/%s", service, environment, key)
		}
		return nil, fmt.Errorf("failed to get value: %w", result.Error)
	}

	return &value, nil
}

// DeleteValue deletes a configuration value
func (m *ConfigManager) DeleteValue(ctx context.Context, service, environment, key string) error {
	result := m.db.WithContext(ctx).
		Where("service = ? AND environment = ? AND key = ?", service, environment, key).
		Delete(&ConfigValue{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete value: %w", result.Error)
	}

	return nil
}

// GetServiceConfig gets all configuration values for a service and environment
func (m *ConfigManager) GetServiceConfig(ctx context.Context, service, environment string) (map[string]string, error) {
	var values []ConfigValue
	result := m.db.WithContext(ctx).
		Where("service = ? AND environment = ? AND is_secret = ?", service, environment, false).
		Order("key").
		Find(&values)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get service config: %w", result.Error)
	}

	config := make(map[string]string)
	for _, value := range values {
		config[value.Key] = value.Value
	}

	return config, nil
}

// GetServiceSecrets gets all secret values for a service and environment
func (m *ConfigManager) GetServiceSecrets(ctx context.Context, service, environment string) (map[string]string, error) {
	var values []ConfigValue
	result := m.db.WithContext(ctx).
		Where("service = ? AND environment = ? AND is_secret = ?", service, environment, true).
		Order("key").
		Find(&values)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get service secrets: %w", result.Error)
	}

	secrets := make(map[string]string)
	for _, value := range values {
		secrets[value.Key] = value.Value
	}

	return secrets, nil
}

// GetEnvironments gets all environments
func (m *ConfigManager) GetEnvironments(ctx context.Context) ([]string, error) {
	var environments []string
	result := m.db.WithContext(ctx).
		Model(&ConfigValue{}).
		Distinct().
		Pluck("environment", &environments)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get environments: %w", result.Error)
	}

	return environments, nil
}

// GetServices gets all services
func (m *ConfigManager) GetServices(ctx context.Context) ([]string, error) {
	var services []string
	result := m.db.WithContext(ctx).
		Model(&ConfigValue{}).
		Distinct().
		Pluck("service", &services)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get services: %w", result.Error)
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

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

					// Insert or update the value
					now := time.Now()
					configValue := &ConfigValue{
						Service:     service,
						Environment: env,
						Key:         key,
						Value:       strValue,
						IsSecret:    isSecret,
						CreatedAt:   now,
						UpdatedAt:   now,
					}

					result := tx.Where(ConfigValue{
						Service:     service,
						Environment: env,
						Key:         key,
					}).Assign(map[string]interface{}{
						"value":      strValue,
						"is_secret":  isSecret,
						"updated_at": now,
					}).FirstOrCreate(configValue)

					if result.Error != nil {
						return fmt.Errorf("failed to import value for %s/%s/%s: %w", service, env, key, result.Error)
					}
				}
			}
		}
		return nil
	})
}

// ExportConfig exports configuration values to a JSON string
func (m *ConfigManager) ExportConfig(ctx context.Context) (string, error) {
	var values []ConfigValue
	result := m.db.WithContext(ctx).
		Order("service, environment, key").
		Find(&values)

	if result.Error != nil {
		return "", fmt.Errorf("failed to export config: %w", result.Error)
	}

	config := make(map[string]map[string]map[string]interface{})
	for _, value := range values {
		// Initialize maps if they don't exist
		if _, ok := config[value.Service]; !ok {
			config[value.Service] = make(map[string]map[string]interface{})
		}
		if _, ok := config[value.Service][value.Environment]; !ok {
			config[value.Service][value.Environment] = make(map[string]interface{})
		}

		// Store value or secret
		if value.IsSecret {
			config[value.Service][value.Environment][value.Key] = map[string]interface{}{
				"value":    value.Value,
				"isSecret": true,
			}
		} else {
			config[value.Service][value.Environment][value.Key] = value.Value
		}
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// GenerateEnvFile generates a .env file for a service and environment
func (m *ConfigManager) GenerateEnvFile(ctx context.Context, service, environment string) (string, error) {
	var values []ConfigValue
	result := m.db.WithContext(ctx).
		Where("service = ? AND environment = ?", service, environment).
		Order("key").
		Find(&values)

	if result.Error != nil {
		return "", fmt.Errorf("failed to generate .env file: %w", result.Error)
	}

	var envContent string
	for _, value := range values {
		envContent += fmt.Sprintf("%s=%s\n", value.Key, value.Value)
	}

	return envContent, nil
}
