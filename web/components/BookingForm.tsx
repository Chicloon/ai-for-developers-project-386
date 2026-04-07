"use client";

import { useState } from 'react';
import { TextInput, Button, Group, Stack, Text } from '@mantine/core';
import { Slot } from '@/lib/api';

interface BookingFormProps {
  slot: Slot;
  onSubmit: (name: string, email: string) => void;
  onCancel: () => void;
}

export default function BookingForm({ slot, onSubmit, onCancel }: BookingFormProps) {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (name && email) {
      onSubmit(name, email);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <Stack>
        <Text fw={500}>
          Запись на {slot.date} в {slot.startTime}
        </Text>
        <TextInput
          label="Имя"
          value={name}
          onChange={(e) => setName(e.currentTarget.value)}
          required
        />
        <TextInput
          label="Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.currentTarget.value)}
          required
        />
        <Group>
          <Button type="submit">Записаться</Button>
          <Button variant="default" onClick={onCancel}>Отмена</Button>
        </Group>
      </Stack>
    </form>
  );
}
