package util

import (
	"testing"
	"time"
)

func TestParseRemindTime(t *testing.T) {
	fixedNow := time.Date(2025, 10, 15, 10, 0, 0, 0, time.Local) // 10:00

	testCases := []struct {
		name     string
		input    string
		wantTime time.Time
		wantErr  bool
	}{
		{
			name:     "минуты как число",
			input:    "20",
			wantTime: fixedNow.Add(20 * time.Minute), // 10:20
			wantErr:  false,
		},
		{
			name:     "duration 2h30m",
			input:    "2h30m",
			wantTime: fixedNow.Add(2*time.Hour + 30*time.Minute), // 12:30
			wantErr:  false,
		},
		{
			name:     "сегодня в 14:30 (после сейчас)",
			input:    "14:30",
			wantTime: time.Date(2025, 10, 15, 14, 30, 0, 0, time.Local),
			wantErr:  false,
		},
		{
			name:     "сегодня в 09:00 (уже прошло → завтра)",
			input:    "09:00",
			wantTime: time.Date(2025, 10, 16, 9, 0, 0, 0, time.Local),
			wantErr:  false,
		},
		{
			name:     "завтра 07:30",
			input:    "завтра 07:30",
			wantTime: time.Date(2025, 10, 16, 7, 30, 0, 0, time.Local),
			wantErr:  false,
		},
		{
			name:     "послезавтра 18:00",
			input:    "послезавтра 18:00",
			wantTime: time.Date(2025, 10, 17, 18, 0, 0, 0, time.Local),
			wantErr:  false,
		},
		{
			name:    "некорректный ввод",
			input:   "абракадабра",
			wantErr: true,
		},
		{
			name:    "некорректное время",
			input:   "завтра 25:00",
			wantErr: true,
		},
		{
			name:    "некорректное ключевое слово",
			input:   "вчера 10:00",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			origTimeNow := timeNow
			timeNow = func() time.Time { return fixedNow }
			defer func() { timeNow = origTimeNow }()

			got, err := ParseRemindTime(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseRemindTime(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
				return
			}
			if err != nil {
				return // ошибка ожидаема
			}

			if !got.Equal(tc.wantTime) {
				t.Errorf("ParseRemindTime(%q) = %s, want %s",
					tc.input, got.Format(time.RFC3339), tc.wantTime.Format(time.RFC3339))
			}
		})
	}
}
