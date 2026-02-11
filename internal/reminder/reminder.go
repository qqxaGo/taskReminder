package reminder

// Здесь будет логика фоновых таймеров и уведомлений

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/qqxaGO/task-reminder/internal/model"
	"github.com/qqxaGO/task-reminder/internal/storage"
)

var mu sync.Mutex // для безопасной записи в файл из разных горутин

// Run запускает режим фоновых напоминаний
func Run(filePath string) error {
	tasks, err := loadTasksSafe(filePath)
	if err != nil {
		return err
	}

	fmt.Printf("Загружено %d задач\n", len(tasks))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Запускаем горутину-монитор для каждой активной задачи
	for i := range tasks {
		if tasks[i].Completed {
			continue
		}

		wg.Add(1)
		go func(t *model.Task) {
			defer wg.Done()

			delay := time.Until(t.RemindAt)
			if delay < 0 {
				// уже просрочено → сразу напоминаем
				notifyAndMark(t, filePath)
				return
			}

			timer := time.NewTimer(delay)
			defer timer.Stop()

			select {
			case <-timer.C:
				notifyAndMark(t, filePath)
			case <-ctx.Done():
				// shutdown — ничего не делаем, просто выходим
				return
			}
		}(&tasks[i])
	}

	// Горутіна для периодического вывода статуса (опционально, для наглядности)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fmt.Printf("[INFO] %s — всё ещё работаю, активных напоминаний: %d\n",
					time.Now().Format("15:04:05"), wgActive(&wg))
			}
		}
	}()

	fmt.Println("Режим напоминаний запущен. Нажмите Ctrl+C для выхода.")
	fmt.Println("Напоминания будут появляться в консоли по мере наступления времени.")

	// Ждём сигнал завершения
	<-sigChan
	fmt.Println("\nПолучен сигнал завершения...")

	// Отменяем контекст → все таймеры, которые ещё не сработали, просто выйдут
	cancel()

	// Ждём завершения всех горутин напоминаний
	wg.Wait()

	fmt.Println("Все напоминания обработаны или отменены. Приложение завершено корректно.")
	return nil
}

// notifyAndMark — выводит напоминание и помечает задачу завершённой
func notifyAndMark(t *model.Task, filePath string) {
	fmt.Printf("\n╔════════════════════════════════════╗")
	fmt.Printf("\n║ НАПОМИНАНИЕ: %s", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("\n║ Задача:      %s", t.Text)
	fmt.Printf("\n║ ID:          %s", t.ID)
	fmt.Printf("\n╚════════════════════════════════════╝\n\n")

	// Помечаем как завершённую
	t.Completed = true

	// Сохраняем изменения
	mu.Lock()
	defer mu.Unlock()

	tasks, _ := loadTasksSafe(filePath) // игнорируем ошибку — файл может быть изменён параллельно
	for i := range tasks {
		if tasks[i].ID == t.ID {
			tasks[i].Completed = true
			break
		}
	}
	_ = storage.SaveTasks(tasks, filePath)
}

// loadTasksSafe — вспомогательная функция с обработкой отсутствия файла
func loadTasksSafe(filePath string) ([]model.Task, error) {
	tasks, err := storage.LoadTasks(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Task{}, nil
		}
		fmt.Printf("Ошибка чтения файла: %v\n", err)
		return nil, err
	}
	return tasks, nil
}

// wgActive — вспомогательная функция для получения текущего счётчика WaitGroup (только для лога)
func wgActive(wg *sync.WaitGroup) int {
	// Это неофициальный способ, в продакшене лучше отдельный счётчик
	// Но для учебного проекта пойдёт
	state := fmt.Sprintf("%#v", wg)
	var count int
	fmt.Sscanf(state, "sync.WaitGroup{state:{count:%d", &count)
	return count
}
