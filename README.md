```markdown
# Subscription Service – тестовое задание Junior Golang Developer

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Быстрый старт

```bash
# Клонировать репозиторий
git clone https://github.com/твой-username/subscription-service.git
cd subscription-service

# Запустить сервис (требуется Docker)
docker compose up --build -d
```

После запуска API доступен на `http://localhost:8080`.

Swagger-документация: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## Стек

- **Язык**: Go 1.21+
- **СУБД**: PostgreSQL 16
- **Миграции**: golang-migrate
- **Роутер**: go-chi/chi
- **Логгер**: zap (uber-go)
- **Документация**: swaggo/swag
- **Контейнеризация**: Docker + Docker Compose

## Конфигурация

Настройки хранятся в `config.yaml`. При запуске через Docker Compose параметры базы данных переопределяются переменными окружения (см. `docker-compose.yml`).

Основные параметры:

| Параметр | Описание | По умолчанию |
|----------|----------|--------------|
| `server.port` | Порт HTTP-сервера | `8080` |
| `database.host` | Хост PostgreSQL | `localhost` |
| `database.port` | Порт PostgreSQL | `5432` |
| `database.user` | Пользователь | `postgres` |
| `database.password` | Пароль | `postgres` |
| `database.dbname` | Имя базы данных | `subscriptions` |

## 📡 API

Все эндпоинты доступны по базовому пути `/api/v1`.

### Создать подписку

```http
POST /api/v1/subscriptions
Content-Type: application/json

{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}
```

Ответ: `201 Created`

### Получить список подписок

```http
GET /api/v1/subscriptions
```

С фильтрацией:

```http
GET /api/v1/subscriptions?user_id=60601fee-...&service_name=Yandex Plus
```

Ответ: `200 OK` (массив объектов)

### Получить подписку по ID

```http
GET /api/v1/subscriptions/{id}
```

Ответ: `200 OK` или `404 Not Found`

### Обновить подписку

```http
PUT /api/v1/subscriptions/{id}
Content-Type: application/json

{
  "service_name": "Netflix",
  "price": 999,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "01-2026",
  "end_date": "12-2026"
}
```

Ответ: `200 OK`

### Удалить подписку

```http
DELETE /api/v1/subscriptions/{id}
```

Ответ: `204 No Content` или `404 Not Found`

### Суммарная стоимость за период

```http
GET /api/v1/subscriptions/summary?from=01-2025&to=12-2025
```

Дополнительные параметры: `user_id`, `service_name`.

Ответ: `200 OK`

```json
{
  "total_cost": 400
}
```

## Тестирование через curl (PowerShell)

```powershell
# Создать
Invoke-RestMethod -Uri http://localhost:8080/api/v1/subscriptions -Method Post -ContentType "application/json" -Body '{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}'

# Получить список
Invoke-RestMethod -Uri http://localhost:8080/api/v1/subscriptions

# Получить по ID (подставьте реальный id из ответа создания)
Invoke-RestMethod -Uri http://localhost:8080/api/v1/subscriptions/ВАШ-UUID

# Обновить
Invoke-RestMethod -Uri http://localhost:8080/api/v1/subscriptions/ВАШ-UUID -Method Put -ContentType "application/json" -Body '{"service_name":"Netflix","price":999,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"01-2026","end_date":"12-2026"}'

# Удалить
Invoke-RestMethod -Uri http://localhost:8080/api/v1/subscriptions/ВАШ-UUID -Method Delete

# Агрегация
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/subscriptions/summary?from=01-2025&to=12-2025"
```

## 📁 Структура проекта

```
.
├── cmd/server/          # Точка входа
├── internal/
│   ├── config/          # Чтение конфигурации
│   ├── database/        # Миграции
│   ├── handler/         # HTTP-обработчики
│   ├── logger/          # Инициализация логгера
│   ├── model/           # Модели данных
│   ├── repository/      # Работа с PostgreSQL
│   └── service/         # Бизнес-логика
├── migrations/          # SQL-файлы миграций
├── docs/                # Сгенерированная Swagger-документация
├── docker-compose.yml
├── Dockerfile
├── config.yaml
└── README.md
```
```

Замени `https://github.com/твой-username/subscription-service.git` на реальный URL твоего репозитория. После этого коммить:

```bash
git add README.md
git commit -m "Add README"
git push
```
