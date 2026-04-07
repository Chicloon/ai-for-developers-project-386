# API Design: Запись на звонок

## Overview

Сервис бронирования времени для звонков. Владелец публикует доступное время через правила доступности, клиент выбирает свободный слот и записывается. Без авторизации, личных кабинетов и внешних интеграций.

## Stack

- **Frontend:** Next.js (App Router) + Tailwind CSS
- **Backend:** Go + `chi` роутер + `pgx` драйвер
- **API Contract:** TypeSpec → OpenAPI генерация
- **Database:** PostgreSQL
- **Repo:** Монорепозиторий

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Next.js    │────▶│   Go API     │────▶│  PostgreSQL  │
│  (Frontend)  │     │  (Backend)   │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
```

- Фронтенд на Vercel, бэкенд на Render/Fly.io
- REST API, JSON request/response
- Слоты вычисляются на лету из правил, не хранятся в БД

## Models

### TimeRange

| Поле        | Тип     | Описание                    |
|-------------|---------|-----------------------------|
| `startTime` | string  | Начало диапазона, `HH:MM`   |
| `endTime`   | string  | Конец диапазона, `HH:MM`    |

### AvailabilityRule

| Поле         | Тип          | Описание                                    |
|--------------|--------------|---------------------------------------------|
| `id`         | string       | UUID                                        |
| `type`       | string       | `recurring` или `one-time`                  |
| `dayOfWeek`  | number?      | 0=Sun..6=Sat (только для recurring)         |
| `date`       | string?      | `YYYY-MM-DD` (только для one-time)          |
| `timeRanges` | TimeRange[]  | Массив диапазонов доступности               |

### BlockedDay

| Поле   | Тип    | Описание              |
|--------|--------|-----------------------|
| `id`   | string | UUID                  |
| `date` | string | `YYYY-MM-DD`          |

### Slot (вычисляемый)

| Поле        | Тип      | Описание                          |
|-------------|----------|-----------------------------------|
| `id`        | string   | Формат: `{date}_{startTime}`      |
| `date`      | string   | `YYYY-MM-DD`                      |
| `startTime` | string   | `HH:MM`                           |
| `endTime`   | string   | `HH:MM`                           |
| `isBooked`  | boolean  | Занят ли слот                     |

### Booking

| Поле            | Тип      | Описание                                                |
|-----------------|----------|---------------------------------------------------------|
| `id`            | string   | UUID                                                    |
| `slotDate`      | string   | `YYYY-MM-DD`                                            |
| `slotStartTime` | string   | `HH:MM`                                                 |
| `name`          | string   | Имя клиента                                             |
| `email`         | string   | Email клиента                                           |
| `status`        | string   | `active` или `cancelled`                                |
| `recurrence`    | string?  | `none`, `daily`, `weekly`, `yearly`                     |
| `dayOfWeek`     | number?  | 0=Sun..6=Sat (только для weekly)                        |
| `endDate`       | string?  | `YYYY-MM-DD` — дата окончания повторений (опционально)  |

## API Endpoints

### Availability Rules

| Метод    | Путь                          | Описание           |
|----------|-------------------------------|--------------------|
| `GET`    | `/api/availability-rules`     | Список всех правил |
| `POST`   | `/api/availability-rules`     | Создать правило    |
| `PUT`    | `/api/availability-rules/{id}`| Обновить правило   |
| `DELETE` | `/api/availability-rules/{id}`| Удалить правило    |

### Blocked Days

| Метод    | Путь                          | Описание           |
|----------|-------------------------------|--------------------|
| `POST`   | `/api/blocked-days`           | Заблокировать день |
| `DELETE` | `/api/blocked-days/{id}`      | Разблокировать день|

### Slots

| Метод | Путь                         | Описание                                          |
|-------|------------------------------|---------------------------------------------------|
| `GET` | `/api/slots?date=YYYY-MM-DD` | Слоты на дату (генерируются из правил + bookings) |

### Bookings

| Метод    | Путь                  | Описание                |
|----------|-----------------------|-------------------------|
| `POST`   | `/api/bookings`       | Создать бронирование    |
| `GET`    | `/api/bookings`       | Список бронирований     |
| `DELETE` | `/api/bookings/{id}`  | Отменить бронирование   |

## Slot Generation Logic

При `GET /api/slots?date=YYYY-MM-DD`:

1. Определить день недели для запрошенной даты
2. Найти все recurring правила для этого дня недели
3. Найти one-time правила для этой даты
4. Исключить дату, если она в BlockedDays
5. Для каждого timeRange сгенерировать слоты с шагом 30 минут
6. Отметить слоты как `isBooked: true`, если они попадают под active booking (включая recurring)
7. Вернуть объединённый список

## Recurrence Logic

При проверке занятости слота:

- `none` — бронирование только на `slotDate`
- `daily` — каждый день начиная с `slotDate` до `endDate` (или бессрочно)
- `weekly` — каждый `dayOfWeek` начиная с `slotDate` до `endDate` (или бессрочно)
- `yearly` — каждый год в ту же дату начиная с `slotDate` до `endDate` (или бессрочно)

## Error Responses

Все ошибки возвращают JSON:

```json
{
  "error": "string"
}
```

HTTP статусы:
- `200` — успешный GET
- `201` — успешный POST (создание ресурса)
- `204` — успешный DELETE
- `400` — невалидный запрос
- `404` — ресурс не найден
- `409` — конфликт (слот уже занят)
- `500` — внутренняя ошибка сервера

## Database Schema

```sql
CREATE TABLE availability_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL CHECK (type IN ('recurring', 'one-time')),
    day_of_week INT CHECK (type != 'recurring' OR day_of_week IS NOT NULL),
    date DATE CHECK (type != 'one-time' OR date IS NOT NULL),
    time_ranges JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE blocked_days (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_date DATE NOT NULL,
    slot_start_time TIME NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
    recurrence VARCHAR(20) DEFAULT 'none' CHECK (recurrence IN ('none', 'daily', 'weekly', 'yearly')),
    day_of_week INT CHECK (recurrence != 'weekly' OR day_of_week IS NOT NULL),
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## TypeSpec Project Structure

```
typespec/
  main.tspec          # Entry point, imports
  models.tspec        # All model definitions
  rules.tspec         # Availability rules endpoints
  blocked-days.tspec  # Blocked days endpoints
  slots.tspec         # Slots endpoint
  bookings.tspec      # Bookings endpoints
```
