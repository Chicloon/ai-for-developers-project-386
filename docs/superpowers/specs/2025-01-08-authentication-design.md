# Дизайн: Система авторизации и multi-user бронирования

**Дата:** 2025-01-08  
**Статус:** Утверждено  
**Автор:** AI Assistant  
**Приоритет:** Высокий

## 1. Обзор

### 1.1 Цель
Добавить полноценную систему авторизации с JWT, превратив сервис из single-user в multi-user платформу, где каждый пользователь может:
- Настраивать своё расписание доступности
- Управлять видимостью через группы (семья, работа, друзья, все)
- Записываться к другим пользователям
- Отменять бронирования (как свои, так и созданные к нему)

### 1.2 Контекст
Существующий Call Booking сервис использует:
- Go 1.22+ с chi router
- PostgreSQL 16
- Next.js App Router с Mantine UI
- Без авторизации (переключатель режимов через localStorage)

## 2. Архитектура базы данных

### 2.1 Таблицы

#### users
Хранение пользователей и хешей паролей.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### visibility_groups
Группы видимости создаются пользователями для управления доступом к их расписанию.

```sql
CREATE TABLE visibility_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    visibility_level VARCHAR(20) NOT NULL CHECK (visibility_level IN ('family', 'work', 'friends', 'public')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### group_members
Связь групп с членами (кто видит расписание владельца группы).

```sql
CREATE TABLE group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES visibility_groups(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_by UUID NOT NULL REFERENCES users(id),
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(group_id, member_id)
);
```

#### schedules
Расписания пользователей — заменяет `availability_rules` и `blocked_days`.

```sql
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('recurring', 'one-time')),
    day_of_week INT CHECK (day_of_week BETWEEN 0 AND 6),
    date DATE,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_blocked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (
        (type = 'recurring' AND day_of_week IS NOT NULL AND date IS NULL) OR
        (type = 'one-time' AND date IS NOT NULL AND day_of_week IS NULL)
    )
);
```

#### bookings
Бронирования слотов.

```sql
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id),
    booker_id UUID NOT NULL REFERENCES users(id),
    owner_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    cancelled_at TIMESTAMP,
    cancelled_by UUID REFERENCES users(id)
);
```

### 2.2 Удалённые/изменённые таблицы

- **Удалены:** `availability_rules`, `blocked_days`
- **Изменена:** `bookings` — добавлены `booker_id`, `owner_id`, `cancelled_at`, `cancelled_by`

### 2.3 Индексы

```sql
CREATE INDEX idx_schedules_user_id ON schedules(user_id);
CREATE INDEX idx_schedules_user_date ON schedules(user_id, date);
CREATE INDEX idx_bookings_booker_id ON bookings(booker_id);
CREATE INDEX idx_bookings_owner_id ON bookings(owner_id);
CREATE INDEX idx_bookings_schedule_id ON bookings(schedule_id);
CREATE INDEX idx_group_members_member_id ON group_members(member_id);
CREATE INDEX idx_visibility_groups_owner_id ON visibility_groups(owner_id);
```

## 3. API Endpoints

### 3.1 Аутентификация (без авторизации)

| Method | Endpoint | Описание |
|--------|----------|----------|
| POST | /api/auth/register | Регистрация нового пользователя |
| POST | /api/auth/login | Вход, получение JWT |
| GET | /api/auth/me | Информация о текущем пользователе |

**POST /api/auth/register**
```json
// Request
{
  "email": "user@example.com",
  "password": "securepassword",
  "name": "Иван Петров"
}

// Response 201
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "Иван Петров"
  }
}
```

**POST /api/auth/login**
```json
// Request
{
  "email": "user@example.com",
  "password": "securepassword"
}

// Response 200
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "Иван Петров"
  }
}
```

### 3.2 Пользователи (требует JWT)

| Method | Endpoint | Описание |
|--------|----------|----------|
| GET | /api/users | Каталог видимых пользователей |
| GET | /api/users/:id | Профиль пользователя |
| GET | /api/users/:id/slots | Доступные слоты пользователя |

**GET /api/users**
```json
// Response 200
{
  "users": [
    {
      "id": "uuid",
      "name": "Иван Петров",
      "email": "ivan@example.com"
    }
  ]
}
```

**GET /api/users/:id/slots**
```json
// Query params: ?date=2025-01-15
// Response 200
{
  "slots": [
    {
      "id": "uuid",
      "date": "2025-01-15",
      "startTime": "10:00",
      "endTime": "10:30",
      "isBooked": false
    }
  ]
}
```

### 3.3 Моё расписание (требует JWT)

| Method | Endpoint | Описание |
|--------|----------|----------|
| GET | /api/my/schedules | Все мои правила |
| POST | /api/my/schedules | Создать правило |
| PUT | /api/my/schedules/:id | Обновить правило |
| DELETE | /api/my/schedules/:id | Удалить правило |

**POST /api/my/schedules**
```json
// Request (recurring)
{
  "type": "recurring",
  "dayOfWeek": 1,
  "startTime": "09:00",
  "endTime": "18:00"
}

// Request (one-time)
{
  "type": "one-time",
  "date": "2025-01-15",
  "startTime": "10:00",
  "endTime": "11:00"
}

// Request (blocked day)
{
  "type": "one-time",
  "date": "2025-01-20",
  "startTime": "00:00",
  "endTime": "23:59",
  "isBlocked": true
}
```

### 3.4 Группы видимости (требует JWT)

| Method | Endpoint | Описание |
|--------|----------|----------|
| GET | /api/my/groups | Мои группы |
| POST | /api/my/groups | Создать группу |
| PUT | /api/my/groups/:id | Обновить группу |
| DELETE | /api/my/groups/:id | Удалить группу |
| POST | /api/my/groups/:id/members | Добавить члена |
| DELETE | /api/my/groups/:id/members/:userId | Удалить члена |

**POST /api/my/groups**
```json
// Request
{
  "name": "Команда разработки",
  "visibilityLevel": "work"
}

// Response 201
{
  "id": "uuid",
  "name": "Команда разработки",
  "visibilityLevel": "work",
  "createdAt": "2025-01-08T10:00:00Z"
}
```

**POST /api/my/groups/:id/members**
```json
// Request (по email)
{
  "email": "colleague@example.com"
}

// Request (по user_id из публичного каталога)
{
  "userId": "uuid"
}
```

### 3.5 Бронирования (требует JWT)

| Method | Endpoint | Описание |
|--------|----------|----------|
| GET | /api/my/bookings | Мои бронирования |
| POST | /api/my/bookings | Создать бронирование |
| DELETE | /api/my/bookings/:id | Отменить бронирование |

**GET /api/my/bookings**
```json
// Response 200
{
  "bookings": [
    {
      "id": "uuid",
      "scheduleId": "uuid",
      "booker": {
        "id": "uuid",
        "name": "Иван Петров"
      },
      "owner": {
        "id": "uuid",
        "name": "Мария Сидорова"
      },
      "date": "2025-01-15",
      "startTime": "10:00",
      "endTime": "10:30",
      "status": "active"
    }
  ]
}
```

**POST /api/my/bookings**
```json
// Request
{
  "ownerId": "uuid",
  "scheduleId": "uuid"
}

// Response 201
{
  "id": "uuid",
  "scheduleId": "uuid",
  "bookerId": "uuid",
  "ownerId": "uuid",
  "status": "active",
  "createdAt": "2025-01-08T10:00:00Z"
}
```

## 4. JWT Middleware и Авторизация

### 4.1 JWT Middleware

```go
// Проверяет заголовок Authorization: Bearer <token>
// Валидирует токен (подпись, expiration)
// Добавляет user_id в context.Context
// Возвращает 401 если токен отсутствует или невалиден
```

### 4.2 Права доступа

**Логика видимости пользователей:**
Пользователь A видит пользователя B, если выполняется любое условие:
1. B имеет группу с `visibility_level = 'public'`
2. A состоит в любой группе B (в таблице `group_members`)
3. B явно добавил A в одну из своих групп

**SQL для получения видимых пользователей:**
```sql
SELECT DISTINCT u.* FROM users u
LEFT JOIN visibility_groups vg ON vg.owner_id = u.id
LEFT JOIN group_members gm ON gm.group_id = vg.id AND gm.member_id = $1
WHERE u.id != $1
  AND (
    EXISTS (SELECT 1 FROM visibility_groups WHERE owner_id = u.id AND visibility_level = 'public')
    OR gm.member_id IS NOT NULL
  );
```

**Права на действия:**
- Свои данные (`/api/my/*`): полный доступ
- Чужие слоты (`/api/users/:id/slots`): только GET, если пользователь видим
- Бронирования: создание только в видимом слоте, отмена своих или созданных к текущему пользователю

## 5. Frontend Архитектура

### 5.1 Структура страниц

```
app/
├── (auth)/                    # Группа без layout с навигацией
│   ├── login/
│   │   └── page.tsx          # Форма входа
│   └── register/
│       └── page.tsx          # Форма регистрации
├── (app)/                     # Группа с авторизацией
│   ├── layout.tsx            # Layout с меню навигации
│   ├── page.tsx              # Каталог пользователей (главная)
│   ├── users/
│   │   └── [id]/
│   │       └── page.tsx      # Профиль + слоты пользователя
│   └── my/
│       ├── page.tsx          # Редирект на /my/schedule
│       ├── schedule/
│       │   └── page.tsx      # Управление расписанием
│       ├── groups/
│       │   └── page.tsx      # Управление группами
│       └── bookings/
│           └── page.tsx      # Мои бронирования
```

### 5.2 Компоненты

#### AuthProvider (context)
Хранит состояние авторизации:
- `user: User | null`
- `token: string | null`
- `isLoading: boolean`
- `login(email, password): Promise<void>`
- `register(email, password, name): Promise<void>`
- `logout(): void`

При загрузке проверяет токен из localStorage и валидирует через `/api/auth/me`.

#### ProtectedRoute
Компонент-обёртка для защищённых страниц:
- Если нет пользователя → редирект на /login
- Если идёт загрузка → показывает Loader

#### Navigation (AppShell)
Верхнее меню:
- Логотип → /
- Каталог → /
- Моё расписание → /my/schedule
- Мои группы → /my/groups
- Мои бронирования → /my/bookings
- Имя пользователя + Выйти

### 5.3 API Client

Обновлённый `lib/api.ts`:
```typescript
// Автоматически добавляет Authorization: Bearer <token>
// Обрабатывает 401 → автоматический logout

export interface User { id: string; email: string; name: string; }
export interface Schedule { ... }
export interface Booking { ... }
export interface VisibilityGroup { ... }

// Auth
export async function register(data: RegisterData): Promise<AuthResponse>
export async function login(data: LoginData): Promise<AuthResponse>
export async function getMe(): Promise<User>

// Users
export async function getUsers(): Promise<User[]>
export async function getUser(id: string): Promise<User>
export async function getUserSlots(id: string, date: string): Promise<Slot[]>

// My Schedule
export async function getMySchedules(): Promise<Schedule[]>
export async function createSchedule(data: CreateScheduleData): Promise<Schedule>
export async function updateSchedule(id: string, data: UpdateScheduleData): Promise<Schedule>
export async function deleteSchedule(id: string): Promise<void>

// Groups
export async function getMyGroups(): Promise<VisibilityGroup[]>
export async function createGroup(data: CreateGroupData): Promise<VisibilityGroup>
export async function updateGroup(id: string, data: UpdateGroupData): Promise<VisibilityGroup>
export async function deleteGroup(id: string): Promise<void>
export async function addGroupMember(groupId: string, data: AddMemberData): Promise<void>
export async function removeGroupMember(groupId: string, userId: string): Promise<void>

// Bookings
export async function getMyBookings(): Promise<Booking[]>
export async function createBooking(data: CreateBookingData): Promise<Booking>
export async function cancelBooking(id: string): Promise<void>
```

## 6. Потоки данных

### 6.1 Регистрация → Вход → Каталог

```
1. Пользователь открывает /register
2. Заполняет форму (email, password, name)
3. POST /api/auth/register
4. Сохраняем {token, user} в AuthProvider
5. Редирект на /
6. Загружаем каталог: GET /api/users
7. Отображаем только видимых пользователей
```

### 6.2 Запись к пользователю

```
1. Пользователь выбирает пользователя из каталога
2. Переход на /users/:id
3. GET /api/users/:id/slots?date=YYYY-MM-DD
4. Пользователь выбирает слот
5. POST /api/my/bookings {ownerId, scheduleId}
6. Backend проверяет:
   - Токен валиден
   - Текущий пользователь видит target пользователя
   - Слот свободен (нет active booking)
7. Создаётся бронирование
8. Редирект на /my/bookings
```

### 6.3 Отмена бронирования

```
1. Пользователь нажимает "Отменить" на /my/bookings
2. DELETE /api/my/bookings/:id
3. Backend проверяет:
   - booking.booker_id == current_user.id ИЛИ
   - booking.owner_id == current_user.id
4. Обновляем status = 'cancelled', cancelled_by, cancelled_at
5. Слот снова доступен
```

## 7. Обработка ошибок

| HTTP | Сценарий | Ответ пользователю |
|------|----------|-------------------|
| 400 | Невалидный JSON | "Проверьте данные формы" |
| 401 | Отсутствует/невалидный токен | Редирект на /login |
| 403 | Нет доступа к пользователю | "У вас нет доступа к этому пользователю" |
| 404 | Ресурс не найден | "Не найдено" |
| 409 | Слот уже занят | "Этот слот уже занят" |
| 409 | Дублирующий email при регистрации | "Пользователь с таким email уже существует" |
| 422 | Ошибка валидации | Показать ошибки полей |
| 500 | Внутренняя ошибка | "Что-то пошло не так" |

## 8. Технические детали

### 8.1 Безопасность

- **Пароли:** bcrypt с cost=10
- **JWT:** HS256, expiration=24h
- **CORS:** настроен для localhost:3000 (dev) и продакшена
- **SQL-инъекции:** использование параметризованных запросов pgx

### 8.2 Генерация слотов

Существующая логика слотов адаптируется:
1. Получаем все `schedules` пользователя на дату
2. Генерируем 30-минутные слоты из `start_time` → `end_time`
3. Исключаем слоты, где есть `is_blocked=true`
4. Помечаем занятые слоты (есть active booking)

### 8.3 Миграция данных

Существующие данные не мигрируются (таблицы заменяются).
Начальная миграция создаёт пустые таблицы.

## 9. Тестирование

### 9.1 Backend тесты

- Unit-тесты для хеширования паролей
- Unit-тесты для JWT middleware
- Integration-тесты для всех API endpoints
- Тесты логики видимости пользователей

### 9.2 Frontend тесты

- Компоненты форм (валидация)
- AuthProvider (состояния, logout)
- API client (обработка ошибок)

## 10. Будущие расширения

Дизайн поддерживает:
- Email-уведомления о бронированиях
- Повторяющиеся бронирования (уже есть поля в bookings)
- Роли администратора платформы
- OAuth (Google, GitHub)
- Загрузка аватаров
- Календарь вместо списка слотов

---

**Следующий шаг:** Создание плана реализации
