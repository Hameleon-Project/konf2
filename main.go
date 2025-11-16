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

// Тестовый граф: A: [B, C], C: [D, E], ...
type Graph map[string][]string

// Проверка корректности конфигурации
func (c *Config) validate() error {
	if c.PackageName == "" {
		return errors.New("package_name не может быть пустым")
	}
	if c.RepoURL == "" {
		return errors.New("repo_url не может быть пустым")
	}
	if c.RepoMode != "local" && c.RepoMode != "remote" && c.RepoMode != "test" {
		return fmt.Errorf("repo_mode должен быть 'local', 'remote' или 'test'")
	}
	return nil
}

// Загружаем конфиг из файла
func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Загружаем тестовый граф из JSON файла
func loadGraph(path string) (Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла графа: %v", err)
	}

	var g Graph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("ошибка парсинга графа: %v", err)
	}
	return g, nil
}

// DFS без рекурсии — главный алгоритм Этапа 3
func dfsIterative(graph Graph, start string) []string {
	visited := make(map[string]bool) // уже посещённые узлы
	stack := []string{start}         // стек вместо рекурсии
	order := []string{}              // порядок обхода

	for len(stack) > 0 {
		// снимаем последний элемент
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[node] {
			continue
		}

		visited[node] = true
		order = append(order, node)

		// добавляем всех зависимых пакетов
		for _, dep := range graph[node] {
			if !visited[dep] {
				stack = append(stack, dep)
			}
		}
	}

	return order
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("Ошибка загрузки конфигурации:", err)
		os.Exit(1)
	}

	if err := cfg.validate(); err != nil {
		fmt.Println("Ошибка проверки конфигурации:", err)
		os.Exit(1)
	}

	if cfg.RepoMode != "test" {
		fmt.Println("Этап 3: включи режим 'test' в config.json")
		os.Exit(0)
	}

	// Загружаем тестовый граф
	graph, err := loadGraph(cfg.RepoURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Выводим граф
	fmt.Println("Граф зависимостей:")
	for pkg, deps := range graph {
		fmt.Printf("%s → %v\n", pkg, deps)
	}

	// Обход DFS без рекурсии
	fmt.Printf("\nDFS начиная с %s:\n", cfg.PackageName)
	order := dfsIterative(graph, cfg.PackageName)

	fmt.Println(order)
}
