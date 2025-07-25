#  EffectiveMobile_TTask
#  📦 Subscription Tracker API

REST API-сервис для учёта онлайн-подписок пользователей.

##  📑 Описание

Сервис позволяет:

- Создавать, просматривать, обновлять и удалять записи о подписках пользователей
- Фильтровать подписки по дате, пользователю и названию сервиса
- Считать общую стоимость подписок за период
- Документировать API через Swagger

##  🧱 Стек технологий

- **Golang**
- **Gin** — HTTP-фреймворк
- **PostgreSQL** — база данных
- **Docker + Docker Compose** — для контейнеризации
- **Swaggo/swag** — автогенерация Swagger-документации

---

##  📁 Структура проекта
srcs/
├── config/ # Docker Compose и .env

├── docs/ # Автогенерированная Swagger-документация

├── internal/

│ ├── handler/ # HTTP-обработчики

│ ├── model/ # Структуры данных

│ ├── repository/ # Доступ к данным (PostgreSQL)

│ └── storage/ # Инициализация хранилища

├── cmd/

│ └── app/ # Точка входа (main.go)

└── migrations/ # SQL-миграции базы данных


---

##  🚀 Быстрый старт

### 🔧 Предварительные требования

- [Docker](https://www.docker.com/)
- [Make](https://www.gnu.org/software/make/)

1.  **🏗 Сборка и запуск проекта**
    ```bash
    make
    ```
    
2.  **🛑 Остановка контейнеров**
    ```bash
    make down
    ```

3.  **🔁 Пересборка проекта**
    ```bash
    make re
    ```

4.  **📚 Генерация Swagger-документации**
    ```bash
    make docs
    ```
    **Swagger будет доступен по адресу:**
    
    http://localhost:8080/swagger/index.html

5.  **🧼 Очистка Docker-среды**
    Полная очистка (включая volumes и образы):
    ```bash
    make fclean
    ```

##  Примеры API запросов

1.  **Получить все подписки**
    ```http
    GET /subscriptions
    ```

2.  **Получить подписки по фильтру**
    ```http
    GET /subscriptions?user_id={uuid}&service_name={string}
    ```

3.  **Создать подписку**
    ```http
    POST /subscriptions
    Content-Type: application/json 
    {
      "user_id": "uuid",
      "service_name": "Netflix",
      "price": 500,
      "start_date": "2024-01",
      "end_date": "2024-12"
    }
    ```

4.  **Посчитать суммарную стоимость**
    ```http
    GET /subscriptions/total_price?from_date=01-2024&to_date=12-2024&user_id={uuid}&service_name={string}
    ```

##  ⚙️ Переменные окружения

Настраиваются в файле srcs/config/.env:
    ```env

    DB_USER=postgres
    DB_PASS=your_password
    DB_NAME=subscriptions
    DB_HOST=db
    ```
