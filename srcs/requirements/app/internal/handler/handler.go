package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"subscription_service/internal/model"
	"subscription_service/internal/repository"
	"log"
)

// SubscriptionHandler — структура с зависимостью репозитория подписок
type SubscriptionHandler struct {
	repo *repository.SubRepository
}

// NewSubscriptionHandler — конструктор для SubscriptionHandler
func NewSubscriptionHandler(repo *repository.SubRepository) *SubscriptionHandler {
	return &SubscriptionHandler{repo: repo}
}

// parseMonthYear — парсит строку формата "MM-YYYY"
func parseMonthYear(s string) (model.MonthYear, error) {
	t, err := time.Parse("01-2006", s)
	return model.MonthYear(t), err
}

// CreateSubscription godoc
// @Summary Создать подписку
// @Description Создает новую запись о подпискена основе JSON-запроса. обработчик POST /subscriptions
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body model.Subscription true "Данные подписки"
// @Success 201 {string} string "Подписка успешно создана"
// @Failure 400 {string} string "Ошибка запроса"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	// Входящая структура для десериализации JSON тела запроса
	var input struct {
		ServiceName string  `json:"service_name" binding:"required"`      // Название сервиса (обязательное)
		Price       int     `json:"price" binding:"required"`             // Цена подписки (обязательное)
		UserID      string  `json:"user_id" binding:"required,uuid"`      // UUID пользователя (обязательное)
		StartDate   string  `json:"start_date" binding:"required"`        // Дата начала (формат MM-YYYY)
		EndDate     *string `json:"end_date"`                             // Опциональная дата окончания (формат MM-YYYY)
	}

	// Привязка JSON к структуре
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Ошибка парсинга тела запроса: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Парсим user_id из строки в UUID
	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Printf("Неверный формат user_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат user_id"})
		return
	}

	// Парсим дату начала подписки
	startDate, err := parseMonthYear(input.StartDate)
	if err != nil {
		log.Printf("Неверный формат start_date: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат start_date, ожидается MM-YYYY"})
		return
	}

	// Если дата окончания задана, парсим её
	var endDate *model.MonthYear
	if input.EndDate != nil {
		ed, err := parseMonthYear(*input.EndDate)
		if err != nil {
			log.Printf("Неверный формат end_date: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат end_date, ожидается MM-YYYY"})
			return
		}
		endDate = &ed
	}

	// Формируем структуру подписки
	sub := &model.Subscription{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      userUUID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Вызываем репозиторий для создания подписки в БД
	err = h.repo.CreateSubscription(c.Request.Context(), sub)
	if err != nil {
		log.Printf("Ошибка при создании подписки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать подписку"})
		return
	}

	log.Printf("Подписка создана: user_id=%s service=%s start_date=%s", userUUID, input.ServiceName, input.StartDate)
	c.JSON(http.StatusCreated, gin.H{"message": "подписка успешно создана"})
}

// GetSubscription godoc
// @Summary Получить подписку
// @Description Обработчик GET /subscriptions/:user_id/:service_name/:start_date. Получает подписку по user_id, service_name и start_date
// @Tags subscriptions
// @Produce json
// @Param user_id path string true "UUID пользователя"
// @Param service_name path string true "Название сервиса"
// @Param start_date path string true "Дата начала подписки (MM-YYYY)"
// @Success 200 {object} model.Subscription
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions/{user_id}/{service_name}/{start_date} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	userIDStr := c.Param("user_id")
	serviceName := c.Param("service_name")
	startDateStr := c.Param("start_date")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Неверный user_id в URL: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный user_id"})
		return
	}

	startDate, err := parseMonthYear(startDateStr)
	if err != nil {
		log.Printf("Неверный формат start_date в URL: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат start_date"})
		return
	}

	sub, err := h.repo.GetSubscription(c.Request.Context(), userID, serviceName, startDate)
	if err != nil {
		log.Printf("Ошибка получения подписки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить подписку"})
		return
	}
	if sub == nil {
		log.Printf("Подписка не найдена: user_id=%s service=%s start_date=%s", userID, serviceName, startDateStr)
		c.JSON(http.StatusNotFound, gin.H{"error": "подписка не найдена"})
		return
	}

	log.Printf("Подписка найдена: user_id=%s service=%s start_date=%s", userID, serviceName, startDateStr)
	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription godoc
// @Summary Обновить подписку
// @Description Обработчик PUT /subscriptions/:user_id/:service_name/:start_date Обновляет цену и дату окончания подписки по ключу user_id + service_name + start_date
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id path string true "UUID пользователя"
// @Param service_name path string true "Название сервиса"
// @Param start_date path string true "Дата начала подписки (MM-YYYY)"
// @Param subscription body model.Subscription true "Обновленные данные подписки"
// @Success 200 {string} string "Подписка успешно обновлена"
// @Failure 400 {string} string "Ошибка запроса"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions/{user_id}/{service_name}/{start_date} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	userIDStr := c.Param("user_id")
	serviceName := c.Param("service_name")
	startDateStr := c.Param("start_date") // MM-YYYY

	var input struct {
		Price   int     `json:"price" binding:"required"` // Новая цена подписки
		EndDate *string `json:"end_date"`                 // Новая дата окончания
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Ошибка парсинга тела запроса на обновление: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Неверный user_id для обновления: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный user_id"})
		return
	}

	startDate, err := parseMonthYear(startDateStr)
	if err != nil {
		log.Printf("Неверный формат start_date для обновления: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат start_date"})
		return
	}

	var endDate *model.MonthYear
	if input.EndDate != nil {
		ed, err := parseMonthYear(*input.EndDate)
		if err != nil {
			log.Printf("Неверный формат end_date для обновления: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат end_date"})
			return
		}
		endDate = &ed
	}

	err = h.repo.UpdateSubscription(c.Request.Context(), userID, serviceName, startDate, input.Price, endDate)
	if err != nil {
		log.Printf("Ошибка обновления подписки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить подписку"})
		return
	}

	log.Printf("Подписка обновлена: user_id=%s service=%s start_date=%s", userID, serviceName, startDateStr)
	c.JSON(http.StatusOK, gin.H{"message": "подписка успешно обновлена"})
}

// DeleteSubscription godoc
// @Summary Удалить подписку
// @Description Обработчик DELETE /subscriptions/:user_id/:service_name/:start_date.Удаляет подписку по user_id, service_name и start_date
// @Tags subscriptions
// @Produce json
// @Param user_id path string true "UUID пользователя"
// @Param service_name path string true "Название сервиса"
// @Param start_date path string true "Дата начала подписки (MM-YYYY)"
// @Success 200 {string} string "Подписка удалена"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions/{user_id}/{service_name}/{start_date} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	userIDStr := c.Param("user_id")
	serviceName := c.Param("service_name")
	startDateStr := c.Param("start_date")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Неверный user_id при удалении: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный user_id"})
		return
	}

	startDate, err := parseMonthYear(startDateStr)
	if err != nil {
		log.Printf("Неверный формат start_date при удалении: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный start_date"})
		return
	}

	err = h.repo.DeleteSubscription(c.Request.Context(), userID, serviceName, startDate)
	if err != nil {
		log.Printf("Ошибка удаления подписки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить подписку"})
		return
	}

	log.Printf("Подписка удалена: user_id=%s service=%s start_date=%s", userID, serviceName, startDateStr)
	c.JSON(http.StatusOK, gin.H{"message": "подписка успешно удалена"})
}

// ListSubscriptions godoc
// @Summary Получить список подписок
// @Description Обработчик GET /subscriptions с параметрами фильтрации. Возвращает список подписок с возможной фильтрацией по user_id, service_name, start_date и end_date
// @Tags subscriptions
// @Produce json
// @Success 200 {array} model.Subscription
// @Failure 400 {string} string "Ошибка валидации входных параметров (например, неверный UUID или формат даты)"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	var userID *uuid.UUID
	var serviceName *string
	var startDate, endDate *model.MonthYear

	// Получаем и парсим query параметры
	if u := c.Query("user_id"); u != "" {
		uid, err := uuid.Parse(u)
		if err == nil {
			userID = &uid
		} else {
			log.Printf("Неверный user_id в query: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный user_id"})
			return
		}
	}

	if s := c.Query("service_name"); s != "" {
		serviceName = &s
	}

	if sd := c.Query("start_date"); sd != "" {
		sdParsed, err := parseMonthYear(sd)
		if err != nil {
			log.Printf("Неверный start_date в query: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный start_date"})
			return
		}
		startDate = &sdParsed
	}

	if ed := c.Query("end_date"); ed != "" {
		edParsed, err := parseMonthYear(ed)
		if err != nil {
			log.Printf("Неверный end_date в query: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный end_date"})
			return
		}
		endDate = &edParsed
	}

	subs, err := h.repo.ListSubscriptions(c.Request.Context(), userID, serviceName, startDate, endDate)
	if err != nil {
		log.Printf("Ошибка получения списка подписок: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить список подписок"})
		return
	}

	log.Printf("Получен список подписок, кол-во: %d", len(subs))
	c.JSON(http.StatusOK, subs)
}

// CalculateTotalPrice godoc
// @Summary Посчитать суммарную стоимость подписок
// @Description Обработчик GET /subscriptions/total_price.Считает стоимость подписок за период с фильтрацией по id пользователя и названию сервиса.
// @Description /subscriptions/total_price?from_date={start_date}&to_date={end_date}&user_id={user_id}&service_name={service_name}
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "UUID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param start_date query string true "Начало периода (MM-YYYY)"
// @Param end_date query string true "Конец периода (MM-YYYY)"
// @Success 200 {object} TotalPriceResponse "Общая сумма"
// @Failure 400 {string} string "Ошибка запроса"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions/total_price [get]
func (h *SubscriptionHandler) CalculateTotalPrice(c *gin.Context) {
	var input struct {
		UserID      *string `form:"user_id"`      // Опциональный user_id для фильтрации
		ServiceName *string `form:"service_name"` // Опциональное имя сервиса для фильтрации
		FromDate    string  `form:"from_date" binding:"required"` // Начальная дата периода (MM-YYYY)
		ToDate      string  `form:"to_date" binding:"required"`   // Конечная дата периода (MM-YYYY)
	}

	// Парсим query параметры из запроса
	if err := c.ShouldBindQuery(&input); err != nil {
		log.Printf("Ошибка парсинга query параметров для подсчета стоимости: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uuid.UUID
	if input.UserID != nil {
		uid, err := uuid.Parse(*input.UserID)
		if err != nil {
			log.Printf("Неверный user_id для подсчета стоимости: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный user_id"})
			return
		}
		userID = &uid
	}

	fromDate, err := parseMonthYear(input.FromDate)
	if err != nil {
		log.Printf("Неверный from_date для подсчета стоимости: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный from_date"})
		return
	}

	toDate, err := parseMonthYear(input.ToDate)
	if err != nil {
		log.Printf("Неверный to_date для подсчета стоимости: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный to_date"})
		return
	}

	// Вызываем репозиторий для подсчета суммы
	total, err := h.repo.CalculateTotalPrice(c.Request.Context(), userID, input.ServiceName, fromDate, toDate)
	if err != nil {
		log.Printf("Ошибка подсчета общей стоимости подписок: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось подсчитать общую стоимость"})
		return
	}
	var totalP TotalPriceResponse
	totalP.TotalPrice = total
	log.Printf("Подсчитана общая стоимость подписок: %d", totalP.TotalPrice)
	c.JSON(http.StatusOK, gin.H{"total_price": totalP.TotalPrice})
}

type TotalPriceResponse struct {
    TotalPrice int `json:"total_price"`
}
