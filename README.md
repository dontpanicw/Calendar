# Calendar Service

Сервис календаря с поддержкой событий и фоновых воркеров для обработки задач.

## Архитектура

Проект построен на основе Clean Architecture с разделением на слои:

- `cmd/` - точка входа приложения
- `internal/domain/` - доменные модели (Event)
- `internal/port/` - интерфейсы (порты) для репозиториев и use cases
- `internal/usecases/` - бизнес-логика приложения
- `internal/adapter/repository/` - реализация репозиториев
  - `postgres/` - PostgreSQL репозиторий
  - `cache/` - In-memory кеш репозиторий
- `internal/input/http/` - HTTP handlers и типы запросов/ответов
- `pkg/migrations/` - миграции базы данных (goose)
- `cleaning_worker/` - воркер архивации событий
- `log_worker/` - асинхронный логгер
- `notify_worker/` - воркер уведомлений

## Воркеры

Приложение использует три фоновых воркера для асинхронной обработки задач:

### 1. Cleaning Worker (`cleaning_worker/`)
Периодически архивирует устаревшие события. Запускается каждые X минут (по умолчанию 10) и переносит в архив все события, дата которых уже прошла.

**Основные функции:**
- Автоматическая архивация прошедших событий
- Настраиваемый период запуска
- Graceful shutdown при остановке приложения

### 2. Log Worker (`log_worker/`)
Асинхронный логгер для обработки логов через канал. HTTP-хендлеры не пишут в stdout напрямую, а отправляют записи в канал, который обрабатывает воркер.

**Основные функции:**
- Неблокирующая запись логов из хендлеров
- Буферизованный канал (100 сообщений)
- Fallback при переполнении канала

### 3. Notify Worker (`notify_worker/`)
Фоновый воркер для отправки уведомлений о событиях. При создании события с напоминанием задача помещается в канал, воркер следит за временем и отправляет напоминания.

**Основные функции:**
- Обработка событий через канал
- Планирование уведомлений (в разработке)
- Отправка напоминаний за час до события

## Установка зависимостей

```bash
go mod download
```

## Запуск

```bash
go run cmd/main.go
```

## Конфигурация

Настройки приложения задаются через переменные окружения (см. `.env.example`).

## API Endpoints

Все эндпоинты принимают JSON или form-data. Дата передается в формате `YYYY-MM-DD`.

### Создать событие
```bash
POST /create_event

# JSON
curl -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "date": "2026-03-15",
    "event": "Встреча с командой"
  }'

# Form-data
curl -X POST http://localhost:8080/create_event \
  -d "user_id=1" \
  -d "date=2026-03-15" \
  -d "event=Встреча с командой"

# Ответ
{"result": "event created"}
```

### Обновить событие
```bash
POST /update_event

# JSON
curl -X POST http://localhost:8080/update_event \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": 1,
    "user_id": 1,
    "date": "2026-03-16",
    "event": "Встреча перенесена"
  }'

# Ответ
{"result": "event updated"}
```

### Удалить событие
```bash
POST /delete_event

# JSON
curl -X POST http://localhost:8080/delete_event \
  -H "Content-Type: application/json" \
  -d '{"event_id": 1}'

# Form-data
curl -X POST http://localhost:8080/delete_event \
  -d "event_id=1"

# Ответ
{"result": "event deleted"}
```

### Получить события за день
```bash
GET /events_for_day?user_id=1&date=2026-03-15

curl "http://localhost:8080/events_for_day?user_id=1&date=2026-03-15"

# Ответ
{
  "result": [
    {
      "event_id": 1,
      "user_id": 1,
      "date": "2026-03-15T10:00:00Z",
      "is_archived": false,
      "description": "Встреча с командой"
    }
  ]
}
```

### Получить события за неделю
```bash
GET /events_for_week?user_id=1&date=2026-03-15

curl "http://localhost:8080/events_for_week?user_id=1&date=2026-03-15"

# Возвращает события с 2026-03-15 по 2026-03-21 (7 дней)
```

### Получить события за месяц
```bash
GET /events_for_month?user_id=1&date=2026-03-15

curl "http://localhost:8080/events_for_month?user_id=1&date=2026-03-15"

# Возвращает все события за март 2026
```

## Тестирование

```bash
# Запустить все тесты
go test ./...

# Запустить тесты репозитория (требуется PostgreSQL)
go test ./internal/adapter/repository/postgres/

# Запустить тесты с покрытием
go test -cover ./...
```

Для тестов репозитория требуется тестовая база данных:
```bash
createdb calendar_test
```

Покрытие тестами:
- `internal/adapter/repository/cache` - 100%
- `internal/usecases` - 56%
- `internal/input/http/handlers` - 37%
- `internal/adapter/repository/postgres` - интеграционные тесты (требуют БД)
