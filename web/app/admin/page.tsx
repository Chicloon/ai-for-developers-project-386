"use client";

import { useState, useEffect } from 'react';
import { Paper, Title, Stack, Button, Group, Loader, Center } from '@mantine/core';
import { DatePickerInput } from '@mantine/dates';
import { getAvailabilityRules, deleteAvailabilityRule, getBlockedDays, deleteBlockedDay, createBlockedDay, AvailabilityRule, BlockedDay } from '@/lib/api';
import AvailabilityRuleList from '@/components/AvailabilityRuleList';
import BlockedDayList from '@/components/BlockedDayList';
import AvailabilityRuleForm from '@/components/AvailabilityRuleForm';

export default function AdminPage() {
  const [rules, setRules] = useState<AvailabilityRule[]>([]);
  const [blockedDays, setBlockedDays] = useState<BlockedDay[]>([]);
  const [loading, setLoading] = useState(true);
  const [showRuleForm, setShowRuleForm] = useState(false);
  const [showBlockDayForm, setShowBlockDayForm] = useState(false);
  const [selectedDate, setSelectedDate] = useState<string | null>(null);

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
    await createBlockedDay(selectedDate);
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

  const handleDateChange = (value: string | null) => {
    setSelectedDate(value);
  };

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
            onChange={handleDateChange}
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
