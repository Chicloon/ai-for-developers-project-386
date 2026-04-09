"use client";

import { Modal, Button, Stack, Text, Group, Paper } from "@mantine/core";
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
      data-testid="booking-modal"
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

        <Group justify="flex-end" mt="md" visibleFrom="sm">
          <Button
            variant="default"
            onClick={onClose}
            disabled={loading}
            data-testid="booking-cancel-button"
          >
            Отмена
          </Button>
          <Button onClick={onConfirm} loading={loading} data-testid="booking-confirm-button">
            Подтвердить запись
          </Button>
        </Group>

        <Paper
          hiddenFrom="sm"
          withBorder
          p="xs"
          radius="md"
          style={{ position: "sticky", bottom: 0 }}
        >
          <Stack gap="xs">
            <Button fullWidth onClick={onConfirm} loading={loading}>
              Подтвердить запись
            </Button>
            <Button fullWidth variant="default" onClick={onClose} disabled={loading}>
              Отмена
            </Button>
          </Stack>
        </Paper>
      </Stack>
    </Modal>
  );
}
