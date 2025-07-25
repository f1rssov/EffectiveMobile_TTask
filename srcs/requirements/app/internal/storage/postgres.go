package storage

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5/pgxpool"
    "log"
    "time"
)

// NewPostgres создает и возвращает пул подключений к PostgreSQL базе данных.
// Пытается подключиться несколько раз с задержкой между попытками (retry).
// Возвращает ошибку, если не удалось установить соединение.
func NewPostgres(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
    const maxAttempts = 5                // Максимальное количество попыток подключения
    const retryInterval = 2 * time.Second // Интервал между попытками

    var pool *pgxpool.Pool
    var err error

    // Парсим строку подключения в конфигурацию pgxpool
    cfg, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("ошибка парсинга строки подключения: %w", err)
    }
    log.Printf("Строка подключения успешно распознана")

    // Цикл повторных попыток подключения к базе
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        log.Printf("Попытка подключения к базе данных #%d из %d", attempt, maxAttempts)

        // Создаем новый пул подключений с конфигурацией
        pool, err = pgxpool.NewWithConfig(ctx, cfg)
        if err != nil {
            log.Printf("Попытка %d: не удалось создать пул подключений: %v", attempt, err)
        } else {
            // Проверяем соединение с базой, выполняя PING
            pingErr := pool.Ping(ctx)
            if pingErr != nil {
                log.Printf("Попытка %d: не удалось подключиться к базе данных: %v", attempt, pingErr)
                err = pingErr
                pool.Close() // Закрываем пул при ошибке подключения
            } else {
                // Успешное подключение — возвращаем пул
                log.Printf("Успешное подключение к базе данных на попытке %d", attempt)
                return pool, nil
            }
        }

        // Если не последняя попытка — ждем перед следующей
        if attempt < maxAttempts {
            log.Printf("Повторная попытка подключения через %v...", retryInterval)
            time.Sleep(retryInterval)
        }
    }

    // Если все попытки не удались — возвращаем ошибку
    return nil, fmt.Errorf("не удалось подключиться к базе данных после %d попыток: %w", maxAttempts, err)
}
