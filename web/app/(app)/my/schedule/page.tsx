"use client";

import { useEffect, useState } from "react";
import {
  Paper,
  Title,
  Stack,
  Table,
  Button,
  Loader,
  Center,
  Text,
  Modal,
  TextInput,
  Select,
  Group,
  ActionIcon,
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { TimeInput } from "@mantine/dates";
import { IconTrash } from "@tabler/icons-react";
import {
  Schedule,
  getMySchedules,
  createSchedule,
  deleteSchedule,
  TimeRange,
} from "@/lib/api";

const DAYS_OF_WEEK = [
  { value: "1", label: "Понедельник" },
  { value: "2", label: "Вторник" },
  { value: "3", label: "Среда" },
  { value: "4", label: "Четверг" },
  { value: "5", label: "Пятница" },
  { value: "6", label: "Суббота" },
  { value: "0", label: "Воскресенье" },
];

export default function MySchedulePage() {
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [loading, setLoading] = useState(true);
  const [opened, { open, close }] = useDisclosure(false);
  const [submitting, setSubmitting] = useState(false);

  // Form state
  const [type, setType] = useState<"recurring" | "one-time">("recurring");
  const [dayOfWeek, setDayOfWeek] = useState<string | null>("1");
  const [date, setDate] = useState<string>("");
  const [timeRanges, setTimeRanges] = useState<TimeRange[]>([
    { startTime: "09:00", endTime: "17:00" },
  ]);

  useEffect(() => {
    loadSchedules();
  }, []);

  const loadSchedules = async () => {
    try {
      setLoading(true);
      const data = await getMySchedules();
      setSchedules(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async () => {
    try {
      setSubmitting(true);
      await createSchedule({
        type,
        dayOfWeek: type === "recurring" ? parseInt(dayOfWeek || "1") : undefined,
        date: type === "one-time" ? date : undefined,
        timeRanges,
      });
      close();
      resetForm();
      await loadSchedules();
    } catch (e) {
      console.error(e);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteSchedule(id);
      await loadSchedules();
    } catch (e) {
      console.error(e);
    }
  };

  const resetForm = () => {
    setType("recurring");
    setDayOfWeek("1");
    setDate("");
    setTimeRanges([{ startTime: "09:00", endTime: "17:00" }]);
  };

  const addTimeRange = () => {
    setTimeRanges([...timeRanges, { startTime: "09:00", endTime: "17:00" }]);
  };

  const removeTimeRange = (index: number) => {
    setTimeRanges(timeRanges.filter((_, i) => i !== index));
  };

  const updateTimeRange = (
    index: number,
    field: keyof TimeRange,
    value: string
  ) => {
    const updated = [...timeRanges];
    updated[index] = { ...updated[index], [field]: value };
    setTimeRanges(updated);
  };

  if (loading) {
    return (
      <Center h="50vh">
        <Loader />
      </Center>
    );
  }

  return (
    <Stack gap="md">
      <Group justify="space-between">
        <Title order={2}>Моё расписание</Title>
        <Button onClick={open}>Добавить расписание</Button>
      </Group>

      {schedules.length === 0 ? (
        <Text c="dimmed">У вас пока нет настроенных расписаний</Text>
      ) : (
        <Paper withBorder>
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Тип</Table.Th>
                <Table.Th>День/дата</Table.Th>
                <Table.Th>Интервалы времени</Table.Th>
                <Table.Th></Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {schedules.map((schedule) => (
                <Table.Tr key={schedule.id}>
                  <Table.Td>
                    {schedule.type === "recurring" ? "Повторяющееся" : "Разовое"}
                  </Table.Td>
                  <Table.Td>
                    {schedule.type === "recurring"
                      ? DAYS_OF_WEEK.find((d) => d.value === String(schedule.dayOfWeek))?.label
                      : schedule.date}
                  </Table.Td>
                  <Table.Td>
                    {schedule.timeRanges.map((tr, i) => (
                      <Text key={i} size="sm">
                        {tr.startTime} - {tr.endTime}
                      </Text>
                    ))}
                  </Table.Td>
                  <Table.Td>
                    <ActionIcon
                      color="red"
                      onClick={() => handleDelete(schedule.id)}
                    >
                      <IconTrash size={16} />
                    </ActionIcon>
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        </Paper>
      )}

      <Modal opened={opened} onClose={close} title="Добавить расписание" size="lg">
        <Stack gap="md">
          <Select
            label="Тип расписания"
            value={type}
            onChange={(value) => setType(value as "recurring" | "one-time")}
            data={[
              { value: "recurring", label: "Повторяющееся" },
              { value: "one-time", label: "Разовое" },
            ]}
          />

          {type === "recurring" ? (
            <Select
              label="День недели"
              value={dayOfWeek}
              onChange={setDayOfWeek}
              data={DAYS_OF_WEEK}
            />
          ) : (
            <TextInput
              label="Дата"
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />
          )}

          <Title order={5}>Интервалы времени</Title>
          {timeRanges.map((range, index) => (
            <Group key={index} grow>
              <TimeInput
                label="Начало"
                value={range.startTime}
                onChange={(e) =>
                  updateTimeRange(index, "startTime", e.target.value)
                }
              />
              <TimeInput
                label="Конец"
                value={range.endTime}
                onChange={(e) =>
                  updateTimeRange(index, "endTime", e.target.value)
                }
              />
              {timeRanges.length > 1 && (
                <ActionIcon
                  color="red"
                  onClick={() => removeTimeRange(index)}
                  mt="xl"
                >
                  <IconTrash size={16} />
                </ActionIcon>
              )}
            </Group>
          ))}

          <Button variant="light" onClick={addTimeRange}>
            Добавить интервал
          </Button>

          <Group justify="flex-end" mt="md">
            <Button variant="default" onClick={close}>
              Отмена
            </Button>
            <Button onClick={handleSubmit} loading={submitting}>
              Сохранить
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  );
}
