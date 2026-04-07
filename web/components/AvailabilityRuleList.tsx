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
