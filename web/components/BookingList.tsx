"use client";

import { Paper, Text, Group, Button, Stack } from '@mantine/core';
import { Booking } from '@/lib/api';

interface BookingListProps {
  bookings: Booking[];
  onCancel: (id: string) => void;
}

export default function BookingList({ bookings, onCancel }: BookingListProps) {
  const active = bookings.filter((b) => b.status === 'active');

  if (active.length === 0) {
    return <Text c="dimmed">Нет предстоящих встреч</Text>;
  }

  return (
    <Stack gap="sm">
      {active.map((booking) => (
        <Paper key={booking.id} p="md" withBorder>
          <Group justify="space-between">
            <div>
              <Text fw={500}>{booking.name}</Text>
              <Text size="sm" c="dimmed">
                {booking.slotDate} в {booking.slotStartTime}
                {booking.recurrence && booking.recurrence !== 'none'
                  ? ` (${booking.recurrence})`
                  : ''}
              </Text>
              <Text size="sm" c="dimmed">{booking.email}</Text>
            </div>
            <Button variant="light" color="red" size="xs" onClick={() => onCancel(booking.id)}>
              Отменить
            </Button>
          </Group>
        </Paper>
      ))}
    </Stack>
  );
}
