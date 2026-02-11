package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// timeNow — для подмены в тестах
var timeNow = time.Now

// ParseRemindTime парсит строку в time.Time
func ParseRemindTime(s string) (time.Time, error) {
	now := timeNow()
	s = strings.ToLower(strings.TrimSpace(s))

	// 1. Пробуем как "голые" минуты (самый частый случай)
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return now.Add(time.Duration(n) * time.Minute), nil
	}

	// 2. Пробуем как Duration (2h30m, 1h20m45s и т.д.)
	d, err := time.ParseDuration(s)
	if err == nil {
		return now.Add(d), nil
	}

	// 3. Пробуем как "HH:MM" → сегодня или завтра
	if t, err := time.Parse("15:04", s); err == nil {
		year, month, day := now.Date()
		candidate := time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, now.Location())
		if candidate.After(now) || candidate.Equal(now) {
			return candidate, nil
		}
		// если уже прошло — на завтра
		return candidate.Add(24 * time.Hour), nil
	}

	// 4. Пробуем "завтра HH:MM", "послезавтра HH:MM"
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
		case "послезавтра", "aftertomorrow", "dayaftertomorrow":
			days = 2
		default:
			return time.Time{}, fmt.Errorf("неизвестное ключевое слово: %s (поддерживаются: завтра, послезавтра, tomorrow)", when)
		}

		year, month, day := now.AddDate(0, 0, days).Date()
		return time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, now.Location()), nil
	}

	return time.Time{}, fmt.Errorf("не удалось распознать формат: %q. Примеры: 20, 2h30m, 14:30, завтра 07:30", s)
}
