# Task Reminder CLI

Простое, но качественное консольное приложение на Go для управления напоминаниями о задачах.

## Возможности

- Добавление задач с гибким указанием времени напоминания  
  (`--in 15`, `--in 2h30m`, `--in "14:30"`, `--in "завтра 07:30"`, `--in "послезавтра 18:00"`)
- Просмотр списка задач с разделением на активные / просроченные / завершённые (`list` + `--all`)
- Отметка задачи выполненной (`complete --id <id>`)
- Удаление задачи (`delete --id <id>`)
- Фоновый режим напоминаний с выводом в консоль по наступлению времени (`run`)
- Graceful shutdown по Ctrl+C
- Сохранение задач в JSON-файл
- Table-driven тесты на парсинг времени
- Цветной табличный вывод в терминале

## Структура проекта 

task-reminder/
├── cmd/taskreminder/     ← точка входа, CLI-логика
├── internal/
│   ├── model/            ← структуры (Task)
│   ├── storage/          ← сохранение/загрузка JSON
│   ├── reminder/         ← concurrency, таймеры, graceful shutdown
│   └── util/             ← вспомогательные функции + тесты (парсинг времени)
├── go.mod
└── README.md

## Установка и запуск

```bash
git clone https://github.com/qqxaGO/task-reminder.git
cd task-reminder
go mod tidy

# Добавить задачу
go run ./cmd/taskreminder add --text "Утренняя пробежка" --in "завтра 07:30"

# Посмотреть список
go run ./cmd/taskreminder list
go run ./cmd/taskreminder list --all

# Отметить выполненной
go run ./cmd/taskreminder complete --id 1234567890123456789

# Запустить режим напоминаний
go run ./cmd/taskreminder run

## Установка и запуск

task-reminder/
├── cmd/taskreminder/     ← точка входа, CLI-логика
├── internal/
│   ├── model/            ← структуры (Task)
│   ├── storage/          ← сохранение/загрузка JSON
│   ├── reminder/         ← concurrency, таймеры, graceful shutdown
│   └── util/             ← вспомогательные функции + тесты (парсинг времени)
├── go.mod
└── README.md