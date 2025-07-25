package repository

import (
	"context"
	"log"
	"strconv"
	"time"

	"subscription_service/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SubRepository представляет собой структуру-репозиторий,
// содержащую подключение к базе данных и методы для работы с таблицей подписок.
type SubRepository struct {
	db *pgxpool.Pool
}

// NewSubRepository инициализирует новый экземпляр SubRepository с переданным пулом подключений к БД.
func (r *SubRepository) NewSubRepository(db *pgxpool.Pool) *SubRepository {
	log.Println("Создание экземпляра SubRepository")
	return &SubRepository{db: db}
}

// CreateSubscription добавляет новую запись о подписке в базу данных.
// Принимает структуру подписки и контекст выполнения.
// Возвращает ошибку, если произошёл сбой при выполнении запроса.
func (r *SubRepository) CreateSubscription(ctx context.Context, sub *model.Subscription) error {
	log.Printf("Создание подписки: %+v", sub)

	startDate := sub.StartDate.ToTime().Truncate(24 * time.Hour)

	var endDate *time.Time
	if sub.EndDate != nil {
		ed := sub.EndDate.ToTime().Truncate(24 * time.Hour)
		endDate = &ed
	}

	query := `
        INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.Exec(ctx, query, sub.ServiceName, sub.Price, sub.UserID, startDate, endDate)
	if err != nil {
		log.Printf("Ошибка при создании подписки: %v", err)
	}
	return err
}

// GetSubscription извлекает одну подписку по userID, имени сервиса и дате начала.
// Возвращает объект подписки или nil, если не найдено.
// Также возвращает ошибку при сбое запроса.
func (r *SubRepository) GetSubscription(ctx context.Context, userID uuid.UUID, serviceName string, startDate model.MonthYear) (*model.Subscription, error) {
	log.Printf("Получение подписки по userID=%s, serviceName=%s, startDate=%s", userID, serviceName, startDate.ToTime().Format("2006-01-02"))

	query := `
        SELECT service_name, price, user_id, start_date, end_date
        FROM subscriptions
        WHERE user_id = $1 AND service_name = $2 AND start_date = $3
    `

	var sub model.Subscription
	var start time.Time
	var end *time.Time

	err := r.db.QueryRow(ctx, query, userID, serviceName, startDate.ToTime()).Scan(
		&sub.ServiceName, &sub.Price, &sub.UserID, &start, &end,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println("Подписка не найдена")
			return nil, nil
		}
		log.Printf("Ошибка при получении подписки: %v", err)
		return nil, err
	}

	sub.StartDate = model.MonthYear(start)
	if end != nil {
		ym := model.MonthYear(*end)
		sub.EndDate = &ym
	}
	log.Printf("Подписка успешно найдена: %+v", sub)
	return &sub, nil
}

// UpdateSubscription обновляет цену и дату окончания подписки.
// Поиск выполняется по userID, имени сервиса и дате начала.
func (r *SubRepository) UpdateSubscription(ctx context.Context, userID uuid.UUID, serviceName string, startDate model.MonthYear, price int, endDate *model.MonthYear) error {
	log.Printf("Обновление подписки userID=%s, serviceName=%s, startDate=%s, новая цена=%d, новая дата окончания=%v", userID, serviceName, startDate.ToTime().Format("2006-01-02"), price, endDate)

	query := `
        UPDATE subscriptions
        SET price = $1, end_date = $2
        WHERE user_id = $3 AND service_name = $4 AND start_date = $5
    `

	var end interface{}
	if endDate != nil {
		end = endDate.ToTime()
	}

	_, err := r.db.Exec(ctx, query, price, end, userID, serviceName, startDate.ToTime())
	if err != nil {
		log.Printf("Ошибка при обновлении подписки: %v", err)
	}
	return err
}

// DeleteSubscription удаляет подписку по userID, имени сервиса и дате начала.
func (r *SubRepository) DeleteSubscription(ctx context.Context, userID uuid.UUID, serviceName string, startDate model.MonthYear) error {
	log.Printf("Удаление подписки userID=%s, serviceName=%s, startDate=%s", userID, serviceName, startDate.ToTime().Format("2006-01-02"))

	query := `
        DELETE FROM subscriptions
        WHERE user_id = $1 AND service_name = $2 AND start_date = $3
    `
	_, err := r.db.Exec(ctx, query, userID, serviceName, startDate.ToTime())
	if err != nil {
		log.Printf("Ошибка при удалении подписки: %v", err)
	}
	return err
}

// ListSubscriptions возвращает список подписок с возможностью фильтрации по userID, имени сервиса и диапазону дат.
// Если фильтры не заданы, возвращаются все записи.
func (r *SubRepository) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string, startDate, endDate *model.MonthYear) ([]model.Subscription, error) {
	log.Println("Получение списка подписок")

	query := `
        SELECT service_name, price, user_id, start_date, end_date
        FROM subscriptions
        WHERE 1=1
    `
	args := []interface{}{}
	i := 1

	if userID != nil {
		query += " AND user_id = $" + strconv.Itoa(i)
		args = append(args, *userID)
		i++
	}
	if serviceName != nil {
		query += " AND service_name ILIKE $" + strconv.Itoa(i)
		args = append(args, "%"+*serviceName+"%")
		i++
	}
	if startDate != nil {
		query += " AND start_date >= $" + strconv.Itoa(i)
		args = append(args, startDate.ToTime())
		i++
	}
	if endDate != nil {
		query += " AND start_date <= $" + strconv.Itoa(i)
		args = append(args, endDate.ToTime())
		i++
	}

	log.Printf("SQL-запрос: %s\nПараметры: %+v", query, args)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса: %v", err)
		return nil, err
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		var startTime time.Time
		var endTimePtr *time.Time

		err := rows.Scan(&sub.ServiceName, &sub.Price, &sub.UserID, &startTime, &endTimePtr)
		if err != nil {
			log.Printf("Ошибка при сканировании строки: %v", err)
			return nil, err
		}

		sub.StartDate = model.MonthYear(startTime)
		if endTimePtr != nil {
			ym := model.MonthYear(*endTimePtr)
			sub.EndDate = &ym
		}
		subs = append(subs, sub)
	}
	log.Printf("Найдено подписок: %d", len(subs))
	return subs, nil
}

// CalculateTotalPrice вычисляет общую сумму подписок в указанном диапазоне дат.
// Может фильтровать по userID и названию сервиса.
func (r *SubRepository) CalculateTotalPrice(ctx context.Context, userID *uuid.UUID, serviceName *string, fromDate, toDate model.MonthYear) (int, error) {
	log.Printf("Подсчёт общей стоимости подписок c %s по %s", fromDate.ToTime().Format("2006-01-02"), toDate.ToTime().Format("2006-01-02"))

	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE start_date BETWEEN $1 AND $2`
	args := []interface{}{fromDate.ToTime(), toDate.ToTime()}
	i := 3

	if userID != nil {
		query += " AND user_id = $" + strconv.Itoa(i)
		args = append(args, *userID)
		i++
	}
	if serviceName != nil {
		query += " AND service_name ILIKE $" + strconv.Itoa(i)
		args = append(args, "%"+*serviceName+"%")
	}

	log.Printf("SQL-запрос: %s\nПараметры: %+v", query, args)

	var total int
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		log.Printf("Ошибка при подсчёте общей стоимости: %v", err)
	}
	log.Printf("Общая сумма подписок: %d", total)
	return total, err
}

// timeMustParse — вспомогательная функция для преобразования строки в time.Time.
// Завершает выполнение программы, если строка некорректна.
func timeMustParse(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		log.Fatalf("Неверный формат даты в БД: %s", s)
	}
	return t
}
