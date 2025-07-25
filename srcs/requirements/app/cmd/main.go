package main

import (
    "context"
    "log"
    "os"

    "github.com/gin-gonic/gin"

    "subscription_service/internal/storage"
    "subscription_service/internal/repository"
    "subscription_service/internal/handler"
    "github.com/swaggo/gin-swagger"
    "github.com/swaggo/files"
    "subscription_service/docs"
)

// main — основная функция запуска сервиса
func main() {
    ctx := context.Background()

    // Получаем строку подключения к базе данных из переменных окружения
    dsn := os.Getenv("DSN")
    if dsn == "" {
        log.Fatal("Переменная окружения DSN не установлена")
    }

    log.Println("Попытка подключения к базе данных...")
    // Создаем пул подключений к PostgreSQL
    db, err := storage.NewPostgres(ctx, dsn)
    if err != nil {
        log.Fatalf("Ошибка подключения к базе данных: %v", err)
    }
    // Закрываем пул при завершении работы программы
    defer func() {
        log.Println("Закрытие подключения к базе данных")
        db.Close()
    }()
    log.Println("Подключение к базе данных успешно")

    // Создаем репозиторий для работы с подписками
    repo := new(repository.SubRepository).NewSubRepository(db)
    log.Println("Репозиторий подписок создан")

    // Создаем HTTP-обработчики, передаем в них репозиторий
    subHandler := handler.NewSubscriptionHandler(repo)
    log.Println("HTTP-обработчики подписок созданы")

    // Создаем роутер Gin — HTTP сервер
    router := gin.Default()

    docs.SwaggerInfo.BasePath = "/"
    // подключаем Swagger UI по пути /swagger/index.html
    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // Регистрируем маршруты (HTTP эндпоинты) и связываем их с обработчиками
    router.POST("/subscriptions", subHandler.CreateSubscription)                     // Создать новую подписку
    router.GET("/subscriptions/:user_id/:service_name/:start_date", subHandler.GetSubscription) // Получить подписку по ключу
    router.PUT("/subscriptions/:user_id/:service_name/:start_date", subHandler.UpdateSubscription) // Обновить подписку
    router.DELETE("/subscriptions/:user_id/:service_name/:start_date", subHandler.DeleteSubscription) // Удалить подписку
    router.GET("/subscriptions", subHandler.ListSubscriptions)                       // Получить список подписок с фильтрацией
    router.GET("/subscriptions/total_price", subHandler.CalculateTotalPrice)         // Подсчитать общую стоимость подписок за период

    log.Println("Запуск сервера на порту :8080")
    // Запускаем HTTP сервер на порту 8080
    if err := router.Run(":8080"); err != nil {
        log.Fatalf("Ошибка при запуске сервера: %v", err)
    }
}
