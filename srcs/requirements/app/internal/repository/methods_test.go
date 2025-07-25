package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"subscription_service/internal/model"
	"subscription_service/internal/storage"

	"github.com/google/uuid"
)

func TestCreateGetUpdateDeleteSubscription(t *testing.T) {
    ctx := context.Background()

    dsn := os.Getenv("DSN")
    if dsn == "" {
        t.Fatal("Переменная окружения DSN не установлена")
    }

    db, err := storage.NewPostgres(ctx, dsn)
    if err != nil {
        t.Fatalf("Не удалось подключиться к базе данных: %v", err)
    }
    defer db.Close()

    repo := new(SubRepository).NewSubRepository(db)

    userID := uuid.New()
    serviceName := "Netflix"
    startDate := model.MonthYear(time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC))
    endDate := model.MonthYear(time.Date(2023, time.December, 1, 0, 0, 0, 0, time.UTC))

    sub := &model.Subscription{
        ServiceName: serviceName,
        Price:       999,
        UserID:      userID,
        StartDate:   startDate,
        EndDate:     &endDate,
    }

    // CREATE
    if err := repo.CreateSubscription(ctx, sub); err != nil {
        t.Fatalf("Создание подписки завершилось ошибкой: %v", err)
    }

    // GET
    gotSub, err := repo.GetSubscription(ctx, userID, serviceName, startDate)
    if err != nil {
        t.Fatalf("Получение подписки завершилось ошибкой: %v", err)
    }
    if gotSub == nil {
        t.Fatal("Получена пустая подписка, ожидалась подписка")
    }
    if gotSub.Price != sub.Price {
        t.Errorf("Получена цена %d, ожидалась %d", gotSub.Price, sub.Price)
    }

    // UPDATE
    newPrice := 1099
    newEndDate := model.MonthYear(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC))
    if err := repo.UpdateSubscription(ctx, userID, serviceName, startDate, newPrice, &newEndDate); err != nil {
        t.Fatalf("Обновление подписки завершилось ошибкой: %v", err)
    }

    updatedSub, err := repo.GetSubscription(ctx, userID, serviceName, startDate)
    if err != nil {
        t.Fatalf("Получение подписки после обновления завершилось ошибкой: %v", err)
    }
    if updatedSub.Price != newPrice {
        t.Errorf("После обновления получена цена %d, ожидалась %d", updatedSub.Price, newPrice)
    }
    if updatedSub.EndDate == nil || !updatedSub.EndDate.ToTime().Equal(newEndDate.ToTime()) {
        t.Errorf("После обновления получена дата окончания %v, ожидалась %v", updatedSub.EndDate, newEndDate)
    }

    // LIST (по userID)
    subs, err := repo.ListSubscriptions(ctx, &userID, nil, nil, nil)
    if err != nil {
        t.Fatalf("Получение списка подписок завершилось ошибкой: %v", err)
    }
    if len(subs) == 0 {
        t.Error("Получен пустой список подписок, ожидался хотя бы один элемент")
    }

    // DELETE
    if err := repo.DeleteSubscription(ctx, userID, serviceName, startDate); err != nil {
        t.Fatalf("Удаление подписки завершилось ошибкой: %v", err)
    }

    deletedSub, err := repo.GetSubscription(ctx, userID, serviceName, startDate)
    if err != nil {
        t.Fatalf("Получение подписки после удаления завершилось ошибкой: %v", err)
    }
    if deletedSub != nil {
        t.Error("Подписка не удалена, ожидалось nil")
    }
}

