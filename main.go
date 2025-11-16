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
	Stage       int    `json:"stage"`
	PackageName string `json:"package_name"`
	RepoURL     string `json:"repo_url"`
	RepoMode    string `json:"repo_mode"`
	OutputMode  string `json:"output_mode"`
}

type Graph map[string][]string

// читаем JSON-файл конфигурации
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

func (c *Config) validate() error {
	if c.Stage < 1 || c.Stage > 4 {
		return errors.New("stage должен быть 1–4")
	}
	return nil
}

// отправляет http запрос по указанному url для этапа 2
func loadRemotePackageJSON(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func loadGraph(path string) (Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g Graph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return g, nil
}

// этап 3 — DFS

func dfsIterative(graph Graph, start string) []string {
	visited := make(map[string]bool)
	stack := []string{start}
	result := []string{}

	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[n] {
			continue
		}

		visited[n] = true
		result = append(result, n)

		for _, dep := range graph[n] {
			if !visited[dep] {
				stack = append(stack, dep)
			}
		}
	}

	return result
}

// этап 4 топологическая сортировка

func topologicalSort(graph Graph, start string) []string {
	visited := make(map[string]bool)
	result := []string{}

	type Frame struct {
		Node     string
		ChildIdx int
		Children []string
	}

	stack := []Frame{{Node: start, Children: graph[start], ChildIdx: 0}}

	for len(stack) > 0 {
		top := &stack[len(stack)-1]

		if !visited[top.Node] {
			visited[top.Node] = true
		}

		if top.ChildIdx < len(top.Children) {
			child := top.Children[top.ChildIdx]
			top.ChildIdx++

			if !visited[child] {
				stack = append(stack, Frame{
					Node:     child,
					Children: graph[child],
					ChildIdx: 0,
				})
			}

			continue
		}

		result = append(result, top.Node)
		stack = stack[:len(stack)-1]
	}

	return result
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Использование: go run main.go <config.json>")
		return
	}

	configPath := os.Args[1]

	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Println("Ошибка загрузки конфигурации:", err)
		return
	}

	if err := cfg.validate(); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	// этап 1
	if cfg.Stage == 1 {
		fmt.Println("Этап 1 — вывод параметров конфигурации:")
		fmt.Printf("%+v\n", cfg)
		return
	}

	// этап 2
	if cfg.Stage == 2 {
		fmt.Println("Этап 2 — прямые зависимости npm-пакета")

		data, err := loadRemotePackageJSON(cfg.RepoURL)
		if err != nil {
			fmt.Println("Ошибка загрузки:", err)
			return
		}

		var pkg struct {
			Dependencies map[string]string `json:"dependencies"`
		}
		json.Unmarshal(data, &pkg)

		fmt.Println("\nПрямые зависимости пакета:")
		for dep, ver := range pkg.Dependencies {
			fmt.Printf("- %s: %s\n", dep, ver)
		}

		return
	}

	// этап 3 и 4

	graph, err := loadGraph(cfg.RepoURL)
	if err != nil {
		fmt.Println("Ошибка загрузки графа:", err)
		return
	}

	fmt.Println("Граф зависимостей:")
	for k, v := range graph {
		fmt.Printf("%s → %v\n", k, v)
	}

	// Этап 3
	if cfg.Stage == 3 {
		fmt.Printf("\nDFS начиная с %s:\n", cfg.PackageName)
		order := dfsIterative(graph, cfg.PackageName)
		fmt.Println(order)
		return
	}

	// Этап 4
	if cfg.Stage == 4 {
		fmt.Printf("\nПорядок загрузки начиная с %s:\n", cfg.PackageName)
		order := topologicalSort(graph, cfg.PackageName)
		fmt.Println(order)
		return
	}
}
