"use client";

import { Modal, Button, Stack, Text, Group } from "@mantine/core";
import { Slot } from "@/lib/api";

interface BookingConfirmationModalProps {
  opened: boolean;
  onClose: () => void;
  slot: Slot | null;
  userName: string;
  onConfirm: () => void;
  loading: boolean;
}

export default function BookingConfirmationModal({
  opened,
  onClose,
  slot,
  userName,
  onConfirm,
  loading,
}: BookingConfirmationModalProps) {
  if (!slot) return null;

  // Format date for display
  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("ru-RU", {
      day: "numeric",
      month: "long",
      year: "numeric",
    });
  };

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title="Подтверждение записи"
      centered
      size="sm"
    >
      <Stack gap="md">
        <Text>
          Вы собираетесь записаться на консультацию к <strong>{userName}</strong>
        </Text>

        <Stack gap="xs">
          <Group gap="xs">
            <Text c="dimmed" size="sm">
              Дата:
            </Text>
            <Text size="sm" fw={500}>
              {formatDate(slot.date)}
            </Text>
          </Group>
          <Group gap="xs">
            <Text c="dimmed" size="sm">
              Время:
            </Text>
            <Text size="sm" fw={500}>
              {slot.startTime} - {slot.endTime}
            </Text>
          </Group>
        </Stack>

        <Text size="sm" c="dimmed">
          После подтверждения запись будет создана и отобразится в разделе "Мои
          бронирования".
        </Text>

        <Group justify="flex-end" mt="md">
          <Button variant="default" onClick={onClose} disabled={loading}>
            Отмена
          </Button>
          <Button onClick={onConfirm} loading={loading}>
            Подтвердить запись
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
