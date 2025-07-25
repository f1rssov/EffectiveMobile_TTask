package model

import (
	"github.com/google/uuid"
	"time"
	"encoding/json"
	"fmt"
)

// MonthYear — кастомный тип для хранения даты в формате "месяц-год" (MM-YYYY).
// В основе лежит time.Time, но при сериализации в JSON форматируется только месяц и год.
type MonthYear time.Time

// MarshalJSON реализует кастомную сериализацию в JSON
// Преобразует дату в строку формата "MM-YYYY"
func (c MonthYear) MarshalJSON() ([]byte, error) {
	t := time.Time(c)
	s := fmt.Sprintf("\"%02d-%d\"", t.Month(), t.Year()) // Пример: "07-2025"
	return []byte(s), nil
}

// UnmarshalJSON реализует кастомную десериализацию из JSON
// Ожидает строку формата "MM-YYYY" и парсит её в time.Time
func (c *MonthYear) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err // Ошибка при распаковке JSON строки
	}
	t, err := time.Parse("01-2006", s) // Парсим "ММ-ГГГГ"
	if err != nil {
		return err // Ошибка парсинга даты
	}
	*c = MonthYear(t)
	return nil
}

// ToTime конвертирует MonthYear в стандартный time.Time
func (c MonthYear) ToTime() time.Time {
	return time.Time(c)
}

// Subscription — модель подписки пользователя на сервис
// @Description Подписка пользователя на онлайн-сервис. Используется для учёта затрат.
// @Example {
//   "service_name": "Netflix",
//   "price": 999,
//   "user_id": "4a79c82c-b09f-4cde-bf80-6edfd680793e",
//   "start_date": "07-2025",
//   "end_date": "12-2025"
// }
type Subscription struct {
	// Название сервиса, например "Netflix"
	ServiceName string `json:"service_name" example:"Netflix"`

	// Цена подписки в рублях
	Price int `json:"price" example:"999"`

	// UUID пользователя
	UserID uuid.UUID `json:"user_id" format:"uuid" example:"4a79c82c-b09f-4cde-bf80-6edfd680793e"`

	// Дата начала подписки (месяц и год)
	StartDate MonthYear `json:"start_date" format:"MM-YYYY" example:"07-2025"`

	// Опциональная дата окончания подписки (месяц и год)
	EndDate *MonthYear `json:"end_date,omitempty" format:"MM-YYYY" example:"12-2025"`
}