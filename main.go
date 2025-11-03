package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Config struct {
	PackageName string `json:"package_name"`
	RepoURL     string `json:"repo_url"`
	RepoMode    string `json:"repo_mode"`
}

type PackageJSON struct {
	Dependencies map[string]string `json:"dependencies"`
}

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

func loadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func getPackageJSON(cfg Config) ([]byte, error) {
	if cfg.RepoMode == "local" {
		data, err := os.ReadFile(cfg.RepoURL)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения локального файла: %v", err)
		}
		return data, nil
	}

	// remote
	resp, err := http.Get(cfg.RepoURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки по URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("не удалось получить файл, статус: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.validate(); err != nil {
		fmt.Printf("Ошибка проверки конфигурации: %v\n", err)
		os.Exit(1)
	}

	data, err := getPackageJSON(cfg)
	if err != nil {
		fmt.Printf("Ошибка получения package.json: %v\n", err)
		os.Exit(1)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		fmt.Printf("Ошибка парсинга JSON: %v\n", err)
		os.Exit(1)
	}

	if len(pkg.Dependencies) == 0 {
		fmt.Println("У пакета нет прямых зависимостей.")
		return
	}

	fmt.Printf("Прямые зависимости пакета %s:\n", cfg.PackageName)
	for dep, version := range pkg.Dependencies {
		fmt.Printf("- %s: %s\n", dep, version)
	}
}
