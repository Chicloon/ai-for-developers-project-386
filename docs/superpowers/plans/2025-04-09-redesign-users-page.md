# Редизайн страницы /users — Select с поиском и слоты

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Переработать страницу `/users`, заменив список карточек на Select с поиском. При выборе пользователя отображать его профиль, календарь и доступные слоты для бронирования.

**Architecture:** Страница загружает список пользователей один раз при монтировании. При выборе пользователя из Select отображается его профиль + DatePicker. При изменении даты загружаются слоты и отображаются кнопками для бронирования.

**Tech Stack:** Next.js (App Router), Mantine UI (`@mantine/core`, `@mantine/dates`), dayjs, TypeScript

---

## File Structure

| File | Responsibility |
|------|----------------|
| `/web/app/(app)/users/page.tsx` | Основной компонент страницы. Содержит Select с поиском, логику выбора пользователя, отображение профиля, DatePicker, список слотов |

---

## Task 1: Обновить импорты и типы

**Files:**
- Modify: `/web/app/(app)/users/page.tsx`

- [ ] **Step 1: Заменить импорты**

Добавить `Select` из `@mantine/core`, `DatePickerInput` из `@mantine/dates`, `dayjs` и дополнительные типы из API.

```typescript
"use client";

import { useEffect, useState } from "react";
import {
  Paper,
  Title,
  Stack,
  Card,
  Group,
  Text,
  Button,
  Loader,
  Center,
  Select,
  Alert,
} from "@mantine/core";
import { DatePickerInput } from "@mantine/dates";
import {
  User,
  Slot,
  getUsers,
  getUser,
  getUserSlots,
  createBooking,
} from "@/lib/api";
import { useAuth } from "@/components/auth/AuthProvider";
import dayjs from "dayjs";
import customParseFormat from "dayjs/plugin/customParseFormat";

dayjs.extend(customParseFormat);
```

- [ ] **Step 2: Добавить хелпер для форматирования времени**

```typescript
// Format time from HH:mm:ss or HH:mm:ss.ssssss or HH:mm to HH:mm
function formatTime(timeStr: string): string {
  const cleanStr = timeStr.split('.')[0];
  if (cleanStr.length === 5 && cleanStr.includes(':')) {
    return cleanStr;
  }
  const parsed = dayjs(cleanStr, "HH:mm:ss");
  return parsed.isValid() ? parsed.format("HH:mm") : cleanStr;
}
```

---

## Task 2: Обновить стейт компонента

**Files:**
- Modify: `/web/app/(app)/users/page.tsx`

- [ ] **Step 1: Заменить стейт в компоненте UsersCatalogPage**

Удалить старый стейт (users, loading) и добавить новый:

```typescript
export default function UsersCatalogPage() {
  const { user: currentUser } = useAuth();
  
  // Список всех пользователей для Select
  const [users, setUsers] = useState<User[]>([]);
  const [usersLoading, setUsersLoading] = useState(true);
  
  // Выбранный пользователь
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [userLoading, setUserLoading] = useState(false);
  
  // Дата и слоты
  const [selectedDate, setSelectedDate] = useState<Date | null>(new Date());
  const [slots, setSlots] = useState<Slot[] | null>(null);
  const [slotsLoading, setSlotsLoading] = useState(false);
  
  // Бронирование
  const [bookingInProgress, setBookingInProgress] = useState<string | null>(null);
  
  // Уведомления
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
```

---

## Task 3: Добавить функции загрузки данных

**Files:**
- Modify: `/web/app/(app)/users/page.tsx`

- [ ] **Step 1: Добавить useEffect для загрузки списка пользователей**

```typescript
  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      setUsersLoading(true);
      const data = await getUsers();
      setUsers(data.users);
    } catch (e) {
      console.error(e);
      setError("Не удалось загрузить список пользователей");
    } finally {
      setUsersLoading(false);
    }
  };
```

- [ ] **Step 2: Добавить useEffect для загрузки выбранного пользователя**

```typescript
  useEffect(() => {
    if (selectedUserId) {
      loadSelectedUser();
    } else {
      setSelectedUser(null);
      setSlots(null);
    }
  }, [selectedUserId]);

  const loadSelectedUser = async () => {
    if (!selectedUserId) return;
    try {
      setUserLoading(true);
      const data = await getUser(selectedUserId);
      setSelectedUser(data);
    } catch (e) {
      console.error(e);
      setError("Не удалось загрузить профиль пользователя");
    } finally {
      setUserLoading(false);
    }
  };
```

- [ ] **Step 3: Добавить useEffect для загрузки слотов**

```typescript
  useEffect(() => {
    if (selectedUserId && selectedDate) {
      loadSlots();
    }
  }, [selectedUserId, selectedDate]);

  const loadSlots = async () => {
    if (!selectedUserId || !selectedDate) return;
    try {
      setSlotsLoading(true);
      const dateStr = selectedDate.toISOString().split("T")[0];
      const data = await getUserSlots(selectedUserId, dateStr);
      setSlots(data.slots ?? []);
    } catch (e) {
      console.error(e);
      setError("Не удалось загрузить слоты");
    } finally {
      setSlotsLoading(false);
    }
  };
```

- [ ] **Step 4: Добавить функцию бронирования**

```typescript
  const handleBooking = async (slot: Slot) => {
    if (!currentUser || !selectedUserId) return;
    try {
      setBookingInProgress(slot.id);
      setError(null);
      setSuccess(null);

      const [scheduleId, slotStartTime] = slot.id.split("_");

      await createBooking({
        ownerId: selectedUserId,
        scheduleId: scheduleId,
        slotStartTime: slotStartTime,
        slotDate: slot.date,
      });

      setSuccess("Запись успешно создана!");
      await loadSlots();
    } catch (e: any) {
      console.error(e);
      setError(e.message || "Не удалось создать запись");
    } finally {
      setBookingInProgress(null);
    }
  };
```

---

## Task 4: Обновить UI — Select и условный рендеринг

**Files:**
- Modify: `/web/app/(app)/users/page.tsx`

- [ ] **Step 1: Подготовить данные для Select**

```typescript
  // Формирование данных для Select
  const selectData = users.map((user) => ({
    value: user.id,
    label: `${user.name} (${user.email})`,
  }));

  const availableSlots = slots?.filter((s) => !s.isBooked) ?? [];
```

- [ ] **Step 2: Обновить return JSX**

```typescript
  return (
    <Stack gap="md" data-testid="users-page">
      <Title order={2} data-testid="users-title">
        Каталог пользователей
      </Title>

      {error && (
        <Alert color="red" onClose={() => setError(null)} withCloseButton>
          {error}
        </Alert>
      )}

      {success && (
        <Alert color="green" onClose={() => setSuccess(null)} withCloseButton>
          {success}
        </Alert>
      )}

      <Select
        label="Выберите пользователя"
        placeholder="Начните вводить имя или email"
        data={selectData}
        value={selectedUserId}
        onChange={setSelectedUserId}
        searchable
        clearable
        disabled={usersLoading}
        data-testid="users-select"
      />

      {userLoading && (
        <Center h="200px">
          <Loader />
        </Center>
      )}

      {selectedUser && !userLoading && (
        <>
          <Card withBorder data-testid={`user-card-${selectedUser.id}`}>
            <Text fw={500} size="lg" data-testid={`user-name-${selectedUser.id}`}>
              {selectedUser.name}
            </Text>
            <Text c="dimmed" data-testid={`user-email-${selectedUser.id}`}>
              {selectedUser.email}
            </Text>
          </Card>

          <Paper p="md" withBorder>
            <Title order={4} mb="md">
              Выберите дату
            </Title>
            <DatePickerInput
              value={selectedDate}
              onChange={(value) => setSelectedDate(value ? new Date(value) : null)}
              locale="ru"
              minDate={new Date()}
            />
          </Paper>

          <Paper p="md" withBorder>
            <Title order={4} mb="md">
              Доступное время
            </Title>
            {slotsLoading ? (
              <Center>
                <Loader />
              </Center>
            ) : availableSlots.length === 0 ? (
              <Text c="dimmed">Нет доступного времени на выбранную дату</Text>
            ) : (
              <Stack gap="xs">
                {availableSlots.map((slot) => (
                  <Card key={slot.id} withBorder padding="sm">
                    <Group justify="space-between">
                      <Text>
                        {formatTime(slot.startTime)} - {formatTime(slot.endTime)}
                      </Text>
                      <Button
                        size="sm"
                        onClick={() => handleBooking(slot)}
                        loading={bookingInProgress === slot.id}
                        disabled={selectedUserId === currentUser?.id}
                      >
                        {selectedUserId === currentUser?.id ? "Это вы" : "Записаться"}
                      </Button>
                    </Group>
                  </Card>
                ))}
              </Stack>
            )}
          </Paper>
        </>
      )}
    </Stack>
  );
}
```

---

## Task 5: Проверка и тестирование

**Files:**
- Modify: `/web/app/(app)/users/page.tsx`

- [ ] **Step 1: Запустить dev-сервер и проверить в браузере**

```bash
cd web && npm run dev
```

Открыть http://localhost:3000/users и проверить:
1. Select отображается с placeholder «Выберите пользователя»
2. При клике на Select открывается список пользователей с именами и email
3. Поиск работает (фильтрует список)
4. При выборе пользователя появляется его карточка
5. При выборе даты загружаются слоты
6. Кнопки «Записаться» работают и создают бронирование
7. Сообщения об успехе/ошибке отображаются

- [ ] **Step 2: Проверить TypeScript компиляцию**

```bash
cd web && npx tsc --noEmit
```

Ожидаемый результат: нет ошибок компиляции.

- [ ] **Step 3: Закоммитить изменения**

```bash
git add web/app/(app)/users/page.tsx
git commit -m "feat: redesign /users page with Select and inline slot booking"
```

---

## Self-Review Checklist

**Spec coverage:**
- [x] Select с поиском для выбора пользователя — Task 4
- [x] Отображение профиля при выборе — Task 4
- [x] DatePicker для выбора даты — Task 4
- [x] Список доступных слотов — Task 4
- [x] Кнопки бронирования — Task 3

**Placeholder scan:**
- [x] Нет TBD/TODO
- [x] Код полный во всех шагах
- [x] Пути файлов указаны точно

**Type consistency:**
- [x] Используются типы `User`, `Slot` из `@/lib/api`
- [x] Сигнатуры функций совпадают с API
