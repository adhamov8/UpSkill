package seed

import (
	"time"

	"github.com/upskill/progressservice/internal/models"
	"gorm.io/gorm"
)

// кладём 30 «занятий» по умолчанию (одна тема в день)
func Plan(db *gorm.DB) {
	var cnt int64
	db.Model(&models.PlanItem{}).Count(&cnt)
	if cnt > 0 {
		return
	}
	start := time.Now().Truncate(24 * time.Hour)
	topics := []string{
		"Введение в Go", "Основы синтаксиса", "Функции и методы",
		"Структуры и интерфейсы", "Пакеты", "Тестирование",
		"Concurrency: goroutines", "Channels", "Context", "Модули",
		"Работа с БД", "SQLx / GORM", "HTTP-сервер net/http",
		"Chi-router", "Middleware", "JWT-аутентификация",
		"Докеризация", "CI/CD", "OpenAI API", "Кеш Redis",
		// … что нибудь добавим может быть или нет
	}
	for i, t := range topics {
		item := models.PlanItem{
			Title:       t,
			ResourceURL: "https://go.dev/",
			StartAt:     start.AddDate(0, 0, i),
			DueAt:       start.AddDate(0, 0, i+1),
		}
		db.Create(&item)
	}
}
