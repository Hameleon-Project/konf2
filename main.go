package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	PackageName string `json:"package_name"`
	RepoURL     string `json:"repo_url"`
	RepoMode    string `json:"repo_mode"`
}

// проверяем корректность параметров
func (c *Config) validate() error {
	if c.PackageName == "" {
		return errors.New("package_name не может быть пустым")
	}
	if c.RepoURL == "" {
		return errors.New("repo_url не может быть пустым")
	}
	if c.RepoMode != "local" && c.RepoMode != "remote" {
		return fmt.Errorf("repo_mode должен быть 'local' или 'remote', получено: %s", c.RepoMode)
	}
	return nil
}

func main() {
	// путь к конфигу
	const configPath = "config.json"

	file, err := os.Open(configPath)
	if err != nil {
		fmt.Printf("Ошибка при открытии %s: %v\n", configPath, err)
		os.Exit(1)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		fmt.Printf("Ошибка чтения JSON: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.validate(); err != nil {
		fmt.Printf("Ошибка в конфигурации: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Конфигурация успешно загружена:")
	fmt.Printf("package_name = %s\n", cfg.PackageName)
	fmt.Printf("repo_url = %s\n", cfg.RepoURL)
	fmt.Printf("repo_mode = %s\n", cfg.RepoMode)
}
