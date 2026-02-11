package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/qqxaGO/task-reminder/internal/model"
	"github.com/qqxaGO/task-reminder/internal/reminder"
	"github.com/qqxaGO/task-reminder/internal/storage"
)

const defaultDataFile = "tasks.json"

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		text := addCmd.String("text", "", "Текст напоминания (обязательно)")
		in := addCmd.String("in", "", "Когда напомнить: минуты, \"2h 30m\", \"14:30\", \"tomorrow 09:00\", \"послезавтра 18:00\"")
		file := addCmd.String("file", defaultDataFile, "Путь к файлу с задачами")
		addCmd.Usage = func() {
			fmt.Println("Использование add:")
			fmt.Println("  --text \"Текст задачи\" (обязательно)")
			fmt.Println("  --in значение (обязательно). Примеры:")
			fmt.Println("     --in 15          → через 15 минут")
			fmt.Println("     --in 2h30m       → через 2 часа 30 минут")
			fmt.Println("     --in 14:30       → сегодня в 14:30 (или завтра, если время прошло)")
			fmt.Println("     --in tomorrow 09:00 → завтра в 09:00")
			fmt.Println("     --in послезавтра 18:00 → послезавтра в 18:00")
			addCmd.PrintDefaults()
		}
		addCmd.Parse(args)

		if *text == "" || *in == "" {
			fmt.Println("Ошибка: нужны оба флага --text и --in")
			addCmd.Usage()
			os.Exit(1)
		}

		remindAt, err := parseRemindTime(*in)
		if err != nil {
			fmt.Printf("Ошибка в формате времени (--in): %v\n", err)
			addCmd.Usage()
			os.Exit(1)
		}

		err = addTask(*text, remindAt, *file)
		if err != nil {
			fmt.Printf("Ошибка при добавлении: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Задача добавлена. Напоминание: %s\n", remindAt.Format("2006-01-02 15:04"))

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		file := listCmd.String("file", defaultDataFile, "Путь к файлу с задачами")
		all := listCmd.Bool("all", false, "Показать также завершённые задачи")
		listCmd.Parse(args)

		err := listTasks(*file, *all)
		if err != nil {
			fmt.Printf("Ошибка при выводе списка: %v\n", err)
			os.Exit(1)
		}

	case "delete":
		deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
		id := deleteCmd.String("id", "", "ID задачи для удаления (обязательно)")
		file := deleteCmd.String("file", defaultDataFile, "Путь к файлу с задачами")
		deleteCmd.Parse(args)

		if *id == "" {
			fmt.Println("Ошибка: укажите --id")
			deleteCmd.Usage()
			os.Exit(1)
		}

		err := deleteTask(*id, *file)
		if err != nil {
			fmt.Printf("Ошибка при удалении: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Задача удалена.")

	case "complete":
		completeCmd := flag.NewFlagSet("complete", flag.ExitOnError)
		id := completeCmd.String("id", "", "ID задачи для отметки как выполненной (обязательно)")
		file := completeCmd.String("file", defaultDataFile, "Путь к файлу с задачами")
		completeCmd.Parse(args)

		if *id == "" {
			fmt.Println("Ошибка: укажите --id")
			completeCmd.Usage()
			os.Exit(1)
		}

		err := completeTask(*id, *file)
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Задача %s отмечена как выполненная.\n", *id)

	case "run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		file := runCmd.String("file", defaultDataFile, "Путь к файлу с задачами")
		runCmd.Parse(args)

		// Запуск режима фоновых напоминаний
		fmt.Println("Запуск режима напоминаний... (Ctrl+C для выхода)")
		err := reminder.Run(*file)
		if err != nil {
			fmt.Printf("Ошибка запуска: %v\n", err)
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Task Reminder CLI")
	fmt.Println("Использование:")
	fmt.Println("  taskreminder add --text \"Купить молоко\" --in 30 [--file tasks.json]")
	fmt.Println("  taskreminder list [--file tasks.json] [--all]")
	fmt.Println("  taskreminder complete --id <id> [--file tasks.json]")
	fmt.Println("  taskreminder delete --id <id> [--file tasks.json]")
	fmt.Println("  taskreminder run [--file tasks.json]          (запускает фоновые напоминания)")
	fmt.Println()
	fmt.Println("Доступные команды:")
	fmt.Println("  add       Добавить новую задачу")
	fmt.Println("  list      Показать задачи (по умолчанию только активные + просроченные)")
	fmt.Println("  complete  Отметить задачу выполненной")
	fmt.Println("  delete    Удалить задачу")
	fmt.Println("  run       Запустить режим напоминаний")
}

func addTask(text string, remindAt time.Time, filePath string) error {
	tasks, err := storage.LoadTasks(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	newTask := model.Task{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Text:      text,
		RemindAt:  remindAt,
		CreatedAt: time.Now(),
		Completed: false,
	}

	tasks = append(tasks, newTask)

	return storage.SaveTasks(tasks, filePath)
}

func listTasks(filePath string, showAll bool) error {
	tasks, err := storage.LoadTasks(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Задач пока нет.")
			return nil
		}
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("Задач пока нет.")
		return nil
	}

	now := time.Now()

	// Разделяем по статусам
	var active, overdue, completed []model.Task
	for _, t := range tasks {
		if t.Completed {
			completed = append(completed, t)
			continue
		}
		if t.RemindAt.Before(now) {
			overdue = append(overdue, t)
		} else {
			active = append(active, t)
		}
	}

	// Функция для вывода одной секции в таблице
	printTable := func(title string, ts []model.Task) {
		if len(ts) == 0 {
			return
		}

		fmt.Printf("\n%s:\n", title)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight|tabwriter.TabIndent)

		// Заголовок таблицы с цветом
		headerColor := colorBlue
		if title == "Просроченные" {
			headerColor = colorRed
		} else if title == "Завершённые" {
			headerColor = colorGray
		}

		fmt.Fprintf(w, "%sID\tВремя\tОсталось/Просрочено\tТекст%s\t\n", headerColor, colorReset)

		for _, t := range ts {
			idShort := t.ID
			if len(idShort) > 12 {
				idShort = idShort[:12] + "..."
			}

			timeStr := t.RemindAt.Format("2006-01-02 15:04")

			var delta string
			deltaColor := colorGreen
			if t.RemindAt.After(now) {
				delta = fmt.Sprintf("через %v", t.RemindAt.Sub(now).Round(time.Minute))
			} else {
				delta = fmt.Sprintf("просрочено %v назад", now.Sub(t.RemindAt).Round(time.Minute))
				deltaColor = colorRed
			}

			textColor := colorReset
			if t.Completed {
				textColor = colorGray
			}

			fmt.Fprintf(w, "%s%s\t%s%s\t%s%s\t%s%s\t%s\n",
				colorReset, idShort,
				colorReset, timeStr,
				deltaColor, delta,
				textColor, t.Text,
				colorReset)
		}

		w.Flush()
		fmt.Println() // пустая строка после таблицы
	}

	fmt.Println("Список задач:")

	printTable("Активные (ожидают)", active)
	printTable("Просроченные", overdue)

	if showAll {
		printTable("Завершённые", completed)
	}

	if len(active)+len(overdue) == 0 && !showAll {
		fmt.Println("(все задачи завершены — используйте --all для просмотра архива)")
	}

	return nil
}

func deleteTask(id string, filePath string) error {
	tasks, err := storage.LoadTasks(filePath)
	if err != nil {
		return err
	}

	newTasks := make([]model.Task, 0, len(tasks))
	found := false
	for _, t := range tasks {
		if t.ID != id {
			newTasks = append(newTasks, t)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("задача с ID %s не найдена", id)
	}

	return storage.SaveTasks(newTasks, filePath)
}

func completeTask(id string, filePath string) error {
	tasks, err := storage.LoadTasks(filePath)
	if err != nil {
		return err
	}

	found := false
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Completed = true
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("задача с ID %s не найдена", id)
	}

	return storage.SaveTasks(tasks, filePath)
}

func parseRemindTime(s string) (time.Time, error) {
	now := time.Now()
	s = strings.ToLower(strings.TrimSpace(s))

	// 1. Пробуем как Duration (15m, 2h30m, 1h20m45s и т.д.)
	d, err := time.ParseDuration(s)
	if err == nil {
		return now.Add(d), nil
	}

	// 2. Пробуем как "HH:MM" → сегодня или завтра
	if t, err := time.Parse("15:04", s); err == nil {
		year, month, day := now.Date()
		candidate := time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, now.Location())
		if candidate.After(now) {
			return candidate, nil
		}
		// если уже прошло — на завтра
		return candidate.Add(24 * time.Hour), nil
	}

	// 3. Пробуем "tomorrow HH:MM", "послезавтра HH:MM"
	parts := strings.Fields(s)
	if len(parts) >= 2 {
		when := parts[0]
		timeStr := parts[1]

		t, err := time.Parse("15:04", timeStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("неверный формат времени: %s", timeStr)
		}

		var days int
		switch when {
		case "завтра", "tomorrow":
			days = 1
		case "послезавтра", "aftertomorrow":
			days = 2
		default:
			return time.Time{}, fmt.Errorf("неизвестное ключевое слово: %s (поддерживаются: завтра, послезавтра, tomorrow)", when)
		}

		year, month, day := now.AddDate(0, 0, days).Date()
		return time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, now.Location()), nil
	}

	// 4. Можно добавить поддержку "в 15:30" или другие форматы позже

	return time.Time{}, fmt.Errorf("не удалось распознать формат: %s", s)
}
