"use client";

import { Group, Button, Text } from '@mantine/core';
import { Slot } from '@/lib/api';

interface SlotPickerProps {
  slots: Slot[];
  selectedSlot: Slot | null;
  onSelect: (slot: Slot) => void;
}

export default function SlotPicker({ slots, selectedSlot, onSelect }: SlotPickerProps) {
  if (slots.length === 0) {
    return <Text c="dimmed">Нет доступных слотов на эту дату</Text>;
  }

  return (
    <Group gap="xs" wrap="wrap">
      {slots.map((slot) => (
        <Button
          key={slot.id}
          variant={selectedSlot?.id === slot.id ? 'filled' : 'outline'}
          disabled={slot.isBooked}
          onClick={() => !slot.isBooked && onSelect(slot)}
          size="sm"
        >
          {slot.startTime}
        </Button>
      ))}
    </Group>
  );
}
