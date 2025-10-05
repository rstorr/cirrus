package config

import (
	"cirrus/internal/services/dynamo/filter"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	DynamoDB DynamoDBConfig `json:"dynamodb"`
}

type DynamoDBConfig struct {
	TableColumnPreferences     map[string][]string                 `json:"table_column_preferences"`
	FilterConditionPreferences map[string][]filter.FilterCondition `json:"filter_condition_preferences"`
}

func NewConfig() *Config {
	return &Config{
		DynamoDB: DynamoDBConfig{
			TableColumnPreferences: make(map[string][]string),
		},
	}
}

func (c *Config) GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".aws-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

func (c *Config) Save() error {
	path, err := c.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Printf("json marshal error %s", err)
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func LoadConfig() (*Config, error) {
	config := NewConfig()

	path, err := config.GetConfigPath()
	if err != nil {
		return config, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist yet, return new config
			return config, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) GetTableColumns(tableName string) []string {
	if cols, ok := c.DynamoDB.TableColumnPreferences[tableName]; ok {
		return cols
	}
	return nil
}

func (c *Config) SetTableColumns(tableName string, columns []string) {
	if c.DynamoDB.TableColumnPreferences == nil {
		c.DynamoDB.TableColumnPreferences = make(map[string][]string)
	}
	c.DynamoDB.TableColumnPreferences[tableName] = columns
}

func (c *Config) SetFilterConditions(tableName string, conditions []filter.FilterCondition) {
	if c.DynamoDB.FilterConditionPreferences == nil {
		c.DynamoDB.FilterConditionPreferences = make(map[string][]filter.FilterCondition)
	}
	c.DynamoDB.FilterConditionPreferences[tableName] = conditions

	log.Printf("conds %v", conditions)
}

func (c *Config) GetFilterConditions(tableName string) []filter.FilterCondition {
	if filters, ok := c.DynamoDB.FilterConditionPreferences[tableName]; ok {
		return filters
	}
	return nil
}
