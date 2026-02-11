package model

import "time"

// Task — одна задача-напоминание
type Task struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	RemindAt  time.Time `json:"remind_at"`
	CreatedAt time.Time `json:"created_at"`
	Completed bool      `json:"completed,omitempty"`
}
