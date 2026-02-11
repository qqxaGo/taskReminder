package storage

// Здесь позже будет логика сохранения/загрузки в JSON

import (
	"encoding/json"
	"os"

	"github.com/qqxaGO/task-reminder/internal/model"
)

func LoadTasks(filePath string) ([]model.Task, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tasks []model.Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func SaveTasks(tasks []model.Task, filePath string) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
