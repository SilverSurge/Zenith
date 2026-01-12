package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"zenith/internal/config"
	"zenith/internal/model"
)

func EnsureDir() {
	if _, err := os.Stat(config.PersistenceDir); os.IsNotExist(err) {
		_ = os.MkdirAll(config.PersistenceDir, 0755)
	}
}

func GetFilename(d time.Time) string {
	return filepath.Join(config.PersistenceDir, fmt.Sprintf("tasks_%s.json", d.Format("2006-01-02")))
}

func LoadTasks(d time.Time) []model.Task {
	EnsureDir()
	data, err := os.ReadFile(GetFilename(d))
	if err != nil {
		return []model.Task{}
	}
	var tasks []model.Task
	_ = json.Unmarshal(data, &tasks)
	return tasks
}

func SaveTasks(d time.Time, tasks []model.Task) {
	EnsureDir()
	data, _ := json.MarshalIndent(tasks, "", "  ")
	_ = os.WriteFile(GetFilename(d), data, 0644)
}

func LoadScripts() []model.Script {
	EnsureDir()
	data, err := os.ReadFile(filepath.Join(config.PersistenceDir, "scripts.json"))
	if err != nil {
		// Return default scripts if none exist
		return []model.Script{
			{Name: "hello-world", Command: "echo 'hello world'", Description: "prints hello world"},
		}
	}
	var scripts []model.Script
	_ = json.Unmarshal(data, &scripts)
	return scripts
}

func SaveScripts(scripts []model.Script) {
	EnsureDir()
	data, _ := json.MarshalIndent(scripts, "", "  ")
	_ = os.WriteFile(filepath.Join(config.PersistenceDir, "scripts.json"), data, 0644)
}
