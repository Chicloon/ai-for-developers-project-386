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
