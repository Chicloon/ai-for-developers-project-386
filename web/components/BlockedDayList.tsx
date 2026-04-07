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
