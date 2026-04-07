# Админ-интерфейс Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Добавить админ-интерфейс для управления доступными слотами через правила доступности и заблокированные дни.

**Architecture:** Переключение режимов (клиент/админ) через кнопку в header. Админ-режим - отдельная страница с управлением availability rules и blocked days. Используется существующий API.

**Tech Stack:** Next.js (App Router), Mantine UI, localStorage для режимов.

---

### Task 1: Добавить API функции для availability rules и blocked days

**Files:**
- Modify: `web/lib/api.ts`

- [ ] **Step 1: Добавить функции для availability rules**

```typescript
export async function getAvailabilityRules(): Promise<AvailabilityRule[]> {
  const res = await fetch(`${API_URL}/api/availability-rules`);
  if (!res.ok) throw new Error("Failed to fetch availability rules");
  return res.json();
}

export async function createAvailabilityRule(data: CreateAvailabilityRule): Promise<AvailabilityRule> {
  const res = await fetch(`${API_URL}/api/availability-rules`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create availability rule");
  return res.json();
}

export async function deleteAvailabilityRule(id: string): Promise<void> {
  const res = await fetch(`${API_URL}/api/availability-rules/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Failed to delete availability rule");
}
```

- [ ] **Step 2: Добавить функции для blocked days**

```typescript
export async function getBlockedDays(): Promise<BlockedDay[]> {
  const res = await fetch(`${API_URL}/api/blocked-days`);
  if (!res.ok) throw new Error("Failed to fetch blocked days");
  return res.json();
}

export async function createBlockedDay(date: string): Promise<BlockedDay> {
  const res = await fetch(`${API_URL}/api/blocked-days`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ date }),
  });
  if (!res.ok) throw new Error("Failed to create blocked day");
  return res.json();
}

export async function deleteBlockedDay(id: string): Promise<void> {
  const res = await fetch(`${API_URL}/api/blocked-days/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Failed to delete blocked day");
}
```

- [ ] **Step 3: Commit**

```bash
git add web/lib/api.ts
git commit -m "feat: add API functions for availability rules and blocked days"
```

---

### Task 2: Создать компонент ModeToggle

**Files:**
- Create: `web/components/ModeToggle.tsx`

- [ ] **Step 1: Написать компонент ModeToggle**

```tsx
"use client";

import { Menu, ActionIcon, Text } from '@mantine/core';
import { useState, useEffect } from 'react';

type Mode = 'client' | 'admin';

export default function ModeToggle() {
  const [mode, setMode] = useState<Mode>('client');

  useEffect(() => {
    const saved = localStorage.getItem('booking-mode') as Mode;
    if (saved) setMode(saved);
  }, []);

  const handleSetMode = (newMode: Mode) => {
    setMode(newMode);
    localStorage.setItem('booking-mode', newMode);
    window.location.reload();
  };

  return (
    <Menu shadow="md" width={200}>
      <Menu.Target>
        <ActionIcon variant="subtle" size="lg" title="Режим">
          ⚙️
        </ActionIcon>
      </Menu.Target>
      <Menu.Dropdown>
        <Menu.Label>Режим</Menu.Label>
        <Menu.Item 
          onClick={() => handleSetMode('client')}
          style={{ backgroundColor: mode === 'client' ? '#f0f0f0' : undefined }}
        >
          Режим записи
        </Menu.Item>
        <Menu.Item 
          onClick={() => handleSetMode('admin')}
          style={{ backgroundColor: mode === 'admin' ? '#f0f0f0' : undefined }}
        >
          Админ-режим
        </Menu.Item>
      </Menu.Dropdown>
    </Menu>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add web/components/ModeToggle.tsx
git commit -m "feat: add ModeToggle component"
```

---

### Task 3: Добавить ModeToggle в Layout

**Files:**
- Modify: `web/app/layout.tsx`

- [ ] **Step 1: Добавить ModeToggle в header**

```tsx
import ModeToggle from '@/components/ModeToggle';
import { Group, Container, Title } from '@mantine/core';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body>
        <header style={{ borderBottom: '1px solid #eee', padding: '12px 0' }}>
          <Container size="sm">
            <Group justify="space-between">
              <Title order={3}>Запись на звонок</Title>
              <ModeToggle />
            </Group>
          </Container>
        </header>
        <main>
          <Container size="sm" py="xl">
            {children}
          </Container>
        </main>
      </body>
    </html>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add web/app/layout.tsx
git commit -m "feat: add ModeToggle to layout header"
```

---

### Task 4: Создать страницу AdminPage

**Files:**
- Create: `web/app/admin/page.tsx`

- [ ] **Step 1: Написать AdminPage**

```tsx
"use client";

import { useState, useEffect } from 'react';
import { Paper, Title, Stack, Button, Group, Text, Loader, Center, SimpleGrid } from '@mantine/core';
import { getAvailabilityRules, deleteAvailabilityRule, getBlockedDays, deleteBlockedDay, createBlockedDay, AvailabilityRule, BlockedDay } from '@/lib/api';
import AvailabilityRuleList from '@/components/AvailabilityRuleList';
import BlockedDayList from '@/components/BlockedDayList';
import AvailabilityRuleForm from '@/components/AvailabilityRuleForm';
import { DatePickerInput } from '@mantine/dates';

export default function AdminPage() {
  const [rules, setRules] = useState<AvailabilityRule[]>([]);
  const [blockedDays, setBlockedDays] = useState<BlockedDay[]>([]);
  const [loading, setLoading] = useState(true);
  const [showRuleForm, setShowRuleForm] = useState(false);
  const [showBlockDayForm, setShowBlockDayForm] = useState(false);
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [rulesData, daysData] = await Promise.all([
        getAvailabilityRules(),
        getBlockedDays(),
      ]);
      setRules(rulesData);
      setBlockedDays(daysData);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteRule = async (id: string) => {
    await deleteAvailabilityRule(id);
    await loadData();
  };

  const handleCreateRule = async () => {
    setShowRuleForm(false);
    await loadData();
  };

  const handleBlockDay = async () => {
    if (!selectedDate) return;
    const dateStr = selectedDate.toISOString().split('T')[0];
    await createBlockedDay(dateStr);
    setShowBlockDayForm(false);
    setSelectedDate(null);
    await loadData();
  };

  const handleDeleteBlockedDay = async (id: string) => {
    await deleteBlockedDay(id);
    await loadData();
  };

  if (loading) {
    return (
      <Center>
        <Loader />
      </Center>
    );
  }

  return (
    <Stack gap="xl">
      <Paper p="md" withBorder>
        <Group justify="space-between" mb="md">
          <Title order={4}>Правила доступности</Title>
          <Button onClick={() => setShowRuleForm(true)}>Добавить правило</Button>
        </Group>
        <AvailabilityRuleList rules={rules} onDelete={handleDeleteRule} />
      </Paper>

      <Paper p="md" withBorder>
        <Group justify="space-between" mb="md">
          <Title order={4}>Заблокированные дни</Title>
          <Button onClick={() => setShowBlockDayForm(true)}>Заблокировать день</Button>
        </Group>
        <BlockedDayList days={blockedDays} onDelete={handleDeleteBlockedDay} />
      </Paper>

      {showRuleForm && (
        <AvailabilityRuleForm
          onClose={() => setShowRuleForm(false)}
          onSuccess={handleCreateRule}
        />
      )}

      {showBlockDayForm && (
        <Paper p="md" withBorder style={{ position: 'fixed', top: '50%', left: '50%', transform: 'translate(-50%, -50%)', zIndex: 1000 }}>
          <Title order={4} mb="md">Заблокировать день</Title>
          <DatePickerInput
            label="Дата"
            placeholder="Выберите дату"
            value={selectedDate}
            onChange={setSelectedDate}
            locale="ru"
            minDate={new Date()}
          />
          <Group mt="md" justify="flex-end">
            <Button variant="outline" onClick={() => setShowBlockDayForm(false)}>Отмена</Button>
            <Button onClick={handleBlockDay}>Заблокировать</Button>
          </Group>
        </Paper>
      )}
    </Stack>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add web/app/admin/page.tsx
git commit -m "feat: add AdminPage component"
```

---

### Task 5: Создать AvailabilityRuleList

**Files:**
- Create: `web/components/AvailabilityRuleList.tsx`

- [ ] **Step 1: Написать компонент**

```tsx
"use client";

import { Stack, Text, Button, Paper, Group } from '@mantine/core';
import { AvailabilityRule } from '@/lib/api';

const DAY_NAMES = ['Воскресенье', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота'];

interface Props {
  rules: AvailabilityRule[];
  onDelete: (id: string) => void;
}

export default function AvailabilityRuleList({ rules, onDelete }: Props) {
  if (rules.length === 0) {
    return <Text c="dimmed">Нет правил доступности</Text>;
  }

  const formatRule = (rule: AvailabilityRule) => {
    const times = rule.timeRanges
      .map(t => `${t.startTime} - ${t.endTime}`)
      .join(', ');
    
    if (rule.type === 'recurring' && rule.dayOfWeek !== undefined) {
      return `Каждый ${DAY_NAMES[rule.dayOfWeek]} ${times}`;
    } else if (rule.type === 'one-time' && rule.date) {
      return `${rule.date}: ${times}`;
    }
    return times;
  };

  return (
    <Stack gap="sm">
      {rules.map((rule) => (
        <Paper key={rule.id} p="sm" withBorder>
          <Group justify="space-between">
            <Text>{formatRule(rule)}</Text>
            <Button 
              variant="outline" 
              color="red" 
              size="xs"
              onClick={() => onDelete(rule.id)}
            >
              Удалить
            </Button>
          </Group>
        </Paper>
      ))}
    </Stack>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add web/components/AvailabilityRuleList.tsx
git commit -m "feat: add AvailabilityRuleList component"
```

---

### Task 6: Создать AvailabilityRuleForm

**Files:**
- Create: `web/components/AvailabilityRuleForm.tsx`

- [ ] **Step 1: Написать компонент формы**

```tsx
"use client";

import { useState } from 'react';
import { Modal, Text, Button, Group, Checkbox, TextInput, Stack } from '@mantine/core';
import { createAvailabilityRule } from '@/lib/api';

interface Props {
  onClose: () => void;
  onSuccess: () => void;
}

const DAYS = [
  { value: 1, label: 'Пн' },
  { value: 2, label: 'Вт' },
  { value: 3, label: 'Ср' },
  { value: 4, label: 'Чт' },
  { value: 5, label: 'Пт' },
  { value: 6, label: 'Сб' },
  { value: 0, label: 'Вс' },
];

export default function AvailabilityRuleForm({ onClose, onSuccess }: Props) {
  const [selectedDays, setSelectedDays] = useState<number[]>([]);
  const [startTime, setStartTime] = useState('09:00');
  const [endTime, setEndTime] = useState('17:00');
  const [saving, setSaving] = useState(false);

  const toggleDay = (day: number) => {
    setSelectedDays(prev => 
      prev.includes(day) 
        ? prev.filter(d => d !== day)
        : [...prev, day]
    );
  };

  const handleSave = async () => {
    if (selectedDays.length === 0) return;
    setSaving(true);
    try {
      for (const day of selectedDays) {
        await createAvailabilityRule({
          type: 'recurring',
          dayOfWeek: day,
          timeRanges: [{ startTime, endTime }],
        });
      }
      onSuccess();
    } catch (e) {
      console.error(e);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal opened onClose={onClose} title="Новое правило доступности" size="md">
      <Stack gap="md">
        <Text>Дни недели</Text>
        <Group>
          {DAYS.map(day => (
            <Checkbox
              key={day.value}
              label={day.label}
              checked={selectedDays.includes(day.value)}
              onChange={() => toggleDay(day.value)}
            />
          ))}
        </Group>

        <Group grow>
          <TextInput
            label="С"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            placeholder="09:00"
          />
          <TextInput
            label="До"
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            placeholder="17:00"
          />
        </Group>

        <Group justify="flex-end" mt="md">
          <Button variant="outline" onClick={onClose}>Отмена</Button>
          <Button onClick={handleSave} loading={saving} disabled={selectedDays.length === 0}>
            Сохранить
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add web/components/AvailabilityRuleForm.tsx
git commit -m "feat: add AvailabilityRuleForm component"
```

---

### Task 7: Создать BlockedDayList

**Files:**
- Create: `web/components/BlockedDayList.tsx`

- [ ] **Step 1: Написать компонент**

```tsx
"use client";

import { Stack, Text, Button, Paper, Group } from '@mantine/core';
import { BlockedDay } from '@/lib/api';

interface Props {
  days: BlockedDay[];
  onDelete: (id: string) => void;
}

export default function BlockedDayList({ days, onDelete }: Props) {
  if (days.length === 0) {
    return <Text c="dimmed">Нет заблокированных дней</Text>;
  }

  return (
    <Stack gap="sm">
      {days.map((day) => (
        <Paper key={day.id} p="sm" withBorder>
          <Group justify="space-between">
            <Text>{day.date}</Text>
            <Button 
              variant="outline" 
              color="red" 
              size="xs"
              onClick={() => onDelete(day.id)}
            >
              Удалить
            </Button>
          </Group>
        </Paper>
      ))}
    </Stack>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add web/components/BlockedDayList.tsx
git commit -m "feat: add BlockedDayList component"
```

---

### Task 8: Обновить page.tsx для переключения режимов

**Files:**
- Modify: `web/app/page.tsx`

- [ ] **Step 1: Добавить переключение режимов**

```tsx
"use client";

import { useState, useEffect } from 'react';
import { Paper, Text, Stack, Title, Loader, Center, Button, Group } from '@mantine/core';
import { DatePickerInput } from '@mantine/dates';
import { getSlots, createBooking, getBookings, cancelBooking, Slot, Booking } from '@/lib/api';
import SlotPicker from '@/components/SlotPicker';
import BookingForm from '@/components/BookingForm';
import BookingList from '@/components/BookingList';
import { useRouter } from 'next/navigation';

export default function Home() {
  const router = useRouter();
  const [mode, setMode] = useState<'client' | 'admin'>('client');
  const [selectedDate, setSelectedDate] = useState<string | null>(
    new Date().toISOString().split('T')[0]
  );
  const [slots, setSlots] = useState<Slot[]>([]);
  const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null);
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const saved = localStorage.getItem('booking-mode') as 'client' | 'admin';
    if (saved === 'admin') {
      router.push('/admin');
      return;
    }
    setMode('client');
    setLoading(false);
  }, [router]);

  useEffect(() => {
    if (mode !== 'client' || !selectedDate) return;
    loadSlots();
    loadBookings();
  }, [mode, selectedDate]);

  const loadSlots = async () => {
    try {
      setLoading(true);
      const data = await getSlots(selectedDate);
      setSlots(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const loadBookings = async () => {
    try {
      const data = await getBookings();
      setBookings(data);
    } catch (e) {
      console.error(e);
    }
  };

  const handleBooking = async (name: string, email: string) => {
    if (!selectedSlot) return;
    await createBooking({
      slotDate: selectedSlot.date,
      slotStartTime: selectedSlot.startTime,
      name,
      email,
    });
    setSelectedSlot(null);
    await loadSlots();
    await loadBookings();
  };

  const handleCancel = async (id: string) => {
    await cancelBooking(id);
    await loadBookings();
    await loadSlots();
  };

  if (mode !== 'client') {
    return null;
  }

  const dateStr = selectedDate || '';

  return (
    <Stack gap="xl">
      <Paper p="md" withBorder>
        <Title order={4} mb="md">Выберите дату</Title>
        <DatePickerInput
          label="Дата"
          placeholder="Выберите дату"
          value={selectedDate}
          onChange={setSelectedDate}
          locale="ru"
          weekendDays={[]}
          minDate={new Date()}
        />
      </Paper>

      <Paper p="md" withBorder>
        <Title order={4} mb="md">
          Доступное время на {dateStr}
        </Title>
        {loading ? (
          <Center>
            <Loader />
          </Center>
        ) : selectedSlot ? (
          <BookingForm
            slot={selectedSlot}
            onSubmit={handleBooking}
            onCancel={() => setSelectedSlot(null)}
          />
        ) : (
          <SlotPicker
            slots={slots}
            selectedSlot={selectedSlot}
            onSelect={setSelectedSlot}
          />
        )}
      </Paper>

      <Paper p="md" withBorder>
        <Title order={4} mb="md">Предстоящие встречи</Title>
        <BookingList bookings={bookings} onCancel={handleCancel} />
      </Paper>
    </Stack>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add web/app/page.tsx
git commit -m "feat: add mode switching in page.tsx"
```

---

### Task 9: Проверить работу

**Files:**
- Run: `npm run dev` in web directory

- [ ] **Step 1: Запустить dev server**

```bash
cd web && npm run dev
```

- [ ] **Step 2: Проверить переключение режимов**
- Открыть http://localhost:3000
- Нажать на шестерёнку в углу
- Выбрать "Админ-режим"
- Должна открыться страница /admin

- [ ] **Step 3: Проверить функционал админки**
- Добавить правило доступности (выбрать несколько дней)
- Удалить правило
- Заблокировать день
- Удалить заблокированный день

- [ ] **Step 4: Проверить режим записи**
- Вернуться в "Режим записи"
- Проверить отображение слотов

- [ ] **Step 5: Commit финальный**

```bash
git add .
git commit -m "feat: complete admin interface implementation"
```
