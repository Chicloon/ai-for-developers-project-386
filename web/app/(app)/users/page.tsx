"use client";

import { useEffect, useState, useCallback, useMemo, useRef } from "react";
import {
  Anchor,
  Paper,
  Title,
  Stack,
  Select,
  Alert,
  Button,
  Group,
  Center,
  Skeleton,
  Stepper,
  Text,
  Tooltip,
  Card,
  ThemeIcon,
} from "@mantine/core";
import { IconInfoCircle } from "@tabler/icons-react";
import { notifications } from "@mantine/notifications";
import { Schedule, ScheduleHeader, ScheduleEventData, ScheduleViewLevel, DateStringValue } from '@mantine/schedule';
import { SegmentedControl } from '@mantine/core';
import Link from "next/link";
import {
  User,
  Slot,
  getUsers,
  getUser,
  getUserSlotsRange,
  createBooking,
} from "@/lib/api";
import { useAuth } from "@/components/auth/AuthProvider";
import BookingConfirmationModal from "@/components/BookingConfirmationModal";
import dayjs from "dayjs";
import customParseFormat from "dayjs/plugin/customParseFormat";
import "dayjs/locale/ru";

dayjs.locale("ru");
dayjs.extend(customParseFormat);

const DEFAULT_SCHEDULE_VIEW: ScheduleViewLevel = "month";
const getSlotEventId = (slot: Slot) => `${slot.id}_${slot.date}`;
const SLOT_DATE_TIME_FORMAT = "YYYY-MM-DD HH:mm";

const isSlotInFuture = (slot: Slot) =>
  dayjs(`${slot.date} ${slot.startTime}`, SLOT_DATE_TIME_FORMAT, true).isAfter(dayjs());

const isSlotBookable = (slot: Slot) => !slot.isBooked && isSlotInFuture(slot);

// Русские названия месяцев
const russianMonths = [
  'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
];

const getPreviousDate = (date: string, view: ScheduleViewLevel) => {
  const d = dayjs(date);
  switch (view) {
    case 'month': return d.subtract(1, 'month').format('YYYY-MM-DD');
    case 'week': return d.subtract(1, 'week').format('YYYY-MM-DD');
    case 'year': return d.subtract(1, 'year').format('YYYY-MM-DD');
    default: return date;
  }
};

const getNextDate = (date: string, view: ScheduleViewLevel) => {
  const d = dayjs(date);
  switch (view) {
    case 'month': return d.add(1, 'month').format('YYYY-MM-DD');
    case 'week': return d.add(1, 'week').format('YYYY-MM-DD');
    case 'year': return d.add(1, 'year').format('YYYY-MM-DD');
    default: return date;
  }
};

const getHeaderLabel = (date: string, view: ScheduleViewLevel) => {
  const d = dayjs(date);
  const monthIndex = d.month();
  
  switch (view) {
    case 'month': 
      return `${russianMonths[monthIndex]} ${d.year()}`;
    case 'week': {
      const start = d.startOf('week');
      const end = start.add(6, 'day');
      const startMonth = start.month();
      const endMonth = end.month();
      
      if (startMonth === endMonth) {
        return `${start.date()} – ${end.date()} ${russianMonths[startMonth]} ${d.year()}`;
      }
      return `${start.date()} ${russianMonths[startMonth]} – ${end.date()} ${russianMonths[endMonth]} ${d.year()}`;
    }
    case 'year': 
      return `${d.year()} год`;
    default: 
      return d.format('LL');
  }
};

const getSlotStartDate = (slot: Slot) =>
  dayjs(`${slot.date} ${slot.startTime}`, SLOT_DATE_TIME_FORMAT, true);

export default function UsersCatalogPage() {
  const { user: currentUser } = useAuth();
  const scheduleContainerRef = useRef<HTMLDivElement>(null);

  // Список всех пользователей для Select
  const [users, setUsers] = useState<User[]>([]);
  const [usersLoading, setUsersLoading] = useState(true);

  // Выбранный пользователь
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [userLoading, setUserLoading] = useState(false);

  // Schedule state
  const [scheduleDate, setScheduleDate] = useState<DateStringValue>(dayjs().format('YYYY-MM-DD'));
  const [scheduleView, setScheduleView] = useState<ScheduleViewLevel>(DEFAULT_SCHEDULE_VIEW);
  const [scheduleEvents, setScheduleEvents] = useState<ScheduleEventData[]>([]);
  const [daySlots, setDaySlots] = useState<Slot[]>([]);

  // Загрузка данных
  const [datesLoading, setDatesLoading] = useState(false);
  const [slotsLoading, setSlotsLoading] = useState(false);

  // Модалка подтверждения
  const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [bookingLoading, setBookingLoading] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [bookingCompleted, setBookingCompleted] = useState(false);

  const bookingStep = bookingCompleted ? 2 : selectedUserId ? 1 : 0;

  const canGoToPreviousPeriod = useMemo(() => {
    const previousDate = getPreviousDate(scheduleDate, scheduleView);
    const previousPeriodEnd =
      scheduleView === "year"
        ? dayjs(previousDate).endOf("year")
        : scheduleView === "month"
          ? dayjs(previousDate).endOf("month")
          : dayjs(previousDate).endOf("week");
    return !previousPeriodEnd.isBefore(dayjs(), "day");
  }, [scheduleDate, scheduleView]);

  const weekSlotsGroupedByDate = useMemo(() => {
    if (scheduleView !== "week") return [];

    const map = new Map<string, Slot[]>();
    for (const event of scheduleEvents) {
      const slot = event.payload?.slot as Slot | undefined;
      if (!slot) continue;
      if (!map.has(slot.date)) map.set(slot.date, []);
      map.get(slot.date)!.push(slot);
    }

    return Array.from(map.entries())
      .map(([date, slots]) => ({
        date,
        slots: slots.sort((a, b) => getSlotStartDate(a).unix() - getSlotStartDate(b).unix()),
      }))
      .sort((a, b) => dayjs(a.date).unix() - dayjs(b.date).unix());
  }, [scheduleEvents, scheduleView]);

  // Загрузка списка пользователей при монтировании
  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      setUsersLoading(true);
      const data = await getUsers();
      setUsers(data.users);
    } catch (e) {
      console.error(e);
      setError("Не удалось загрузить список пользователей");
    } finally {
      setUsersLoading(false);
    }
  };

  // Загрузка выбранного пользователя
  useEffect(() => {
    if (selectedUserId) {
      loadSelectedUser();
      // Reset schedule state when user changes
      setScheduleDate(dayjs().format('YYYY-MM-DD'));
      setScheduleView(DEFAULT_SCHEDULE_VIEW);
      loadScheduleEvents(DEFAULT_SCHEDULE_VIEW, dayjs().format('YYYY-MM-DD'), selectedUserId);
    } else {
      setSelectedUser(null);
      setScheduleEvents([]);
      setDaySlots([]);
    }
    setBookingCompleted(false);
  }, [selectedUserId]);

  const loadSelectedUser = async () => {
    if (!selectedUserId) return;
    try {
      setUserLoading(true);
      const data = await getUser(selectedUserId);
      setSelectedUser(data);
    } catch (e) {
      console.error(e);
      setError("Не удалось загрузить профиль пользователя");
    } finally {
      setUserLoading(false);
    }
  };

  // Загрузка событий для Schedule
  const loadScheduleEvents = useCallback(async (
    view: ScheduleViewLevel,
    date: string,
    userId?: string
  ) => {
    const targetUserId = userId || selectedUserId;
    if (!targetUserId) return;

    try {
      setDatesLoading(true);

      if (view === 'month') {
        const start = dayjs(date).startOf("month").format("YYYY-MM-DD");
        const end = dayjs(date).endOf("month").format("YYYY-MM-DD");
        const data = await getUserSlotsRange(targetUserId, start, end);
        const availableSlots = (data.slots ?? []).filter(isSlotBookable);
        const events: ScheduleEventData[] = availableSlots.map((slot) => ({
          id: getSlotEventId(slot),
          title: `${slot.startTime} - ${slot.endTime}`,
          start: `${slot.date} ${slot.startTime}:00`,
          end: `${slot.date} ${slot.endTime}:00`,
          color: 'green',
          payload: { type: 'slot', slot }
        }));
        setScheduleEvents(events);
      } else if (view === 'week') {
        const start = dayjs(date).startOf("week").format("YYYY-MM-DD");
        const end = dayjs(date).endOf("week").format("YYYY-MM-DD");
        const data = await getUserSlotsRange(targetUserId, start, end);
        const weekSlots = (data.slots ?? []).filter(isSlotBookable);
        setDaySlots(weekSlots); // Сохраняем для модалки

        // Маппим в events для Schedule (как в Day view)
        const weekEvents: ScheduleEventData[] = weekSlots.map((slot) => ({
          id: getSlotEventId(slot),
          title: `${slot.startTime} - ${slot.endTime}`,
          start: `${slot.date} ${slot.startTime}:00`,
          end: `${slot.date} ${slot.endTime}:00`,
          color: 'green',
          payload: { type: 'slot', slot },
        }));

        setScheduleEvents(weekEvents);
      } else if (view === 'day') {
        const data = await getUserSlotsRange(targetUserId, date, date);
        const availableSlots = (data.slots ?? []).filter(isSlotBookable);
        setDaySlots(availableSlots);
        const events: ScheduleEventData[] = availableSlots.map((slot) => ({
          id: getSlotEventId(slot),
          title: `${slot.startTime} - ${slot.endTime}`,
          start: `${slot.date} ${slot.startTime}:00`,
          end: `${slot.date} ${slot.endTime}:00`,
          color: 'green',
          payload: { type: 'slot', slot }
        }));
        setScheduleEvents(events);
      }
    } catch (e) {
      console.error(e);
      setError("Не удалось загрузить доступные даты");
    } finally {
      setDatesLoading(false);
    }
  }, [selectedUserId]);

  // Обработчик смены view
  const handleViewChange = (view: ScheduleViewLevel) => {
    setScheduleView(view);
    if (selectedUserId) {
      loadScheduleEvents(view, scheduleDate);
    }
  };

  // Обработчик смены даты
  const handleDateChange = (date: string) => {
    setScheduleDate(date);
    if (selectedUserId) {
      loadScheduleEvents(scheduleView, date);
    }
  };

  // В YearView клики по дням переводим в MonthView без onDayClick,
  // чтобы избежать протекания unsupported props в DOM.
  useEffect(() => {
    if (scheduleView !== "year") return;
    const container = scheduleContainerRef.current;
    if (!container) return;

    const handleYearDayClick = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      const button = target.closest("button");
      if (!button) return;

      const ariaLabel = button.getAttribute("aria-label");
      if (!ariaLabel) return;

      const parsedDate = dayjs(
        ariaLabel,
        ["MMMM D, YYYY", "D MMMM YYYY", "YYYY-MM-DD"],
        "ru",
        true
      );
      if (!parsedDate.isValid()) return;

      const targetDate = parsedDate.format("YYYY-MM-DD");
      setScheduleDate(targetDate);
      setScheduleView("month");
      if (selectedUserId) {
        loadScheduleEvents("month", targetDate);
      }
    };

    container.addEventListener("click", handleYearDayClick, true);
    return () => container.removeEventListener("click", handleYearDayClick, true);
  }, [scheduleView, selectedUserId, loadScheduleEvents]);

  const handleEventClick = useCallback((event: ScheduleEventData) => {
    if (event.payload?.type === 'slot') {
      if (!isSlotInFuture(event.payload.slot)) {
        setError("Нельзя записаться на время, которое уже прошло");
        return;
      }
      setSelectedSlot(event.payload.slot);
      setModalOpen(true);
    } else if (event.payload?.type === 'badge') {
      setScheduleDate(event.payload.date);
      setScheduleView('day');
      loadScheduleEvents('day', event.payload.date);
    }
  }, [loadScheduleEvents]);

  // Тело события (клик обрабатывает ScheduleEvent; в «+N ещё» см. patch @mantine/schedule — onEventClick в MoreEvents).
  const renderScheduleEventBody = useCallback((event: ScheduleEventData) => (
    <Text size="sm" truncate>
      {event.title}
    </Text>
  ), []);

  // Обработка подтверждения бронирования
  const handleBookingConfirm = async () => {
    if (!selectedSlot || !selectedUserId || !currentUser) return;
    if (!isSlotInFuture(selectedSlot)) {
      setError("Нельзя записаться на время, которое уже прошло");
      setModalOpen(false);
      return;
    }

    try {
      setBookingLoading(true);
      setError(null);

      const [scheduleId, slotStartTime] = selectedSlot.id.split("_");

      await createBooking({
        ownerId: selectedUserId,
        scheduleId: scheduleId,
        slotStartTime: slotStartTime,
        slotDate: selectedSlot.date,
      });

      setBookingCompleted(true);
      setModalOpen(false);
      notifications.show({
        message: (
          <Stack gap={4}>
            <Text size="sm">Запись успешно создана.</Text>
            <Anchor component={Link} href="/my/bookings" size="sm" fw={500}>
              Мои бронирования
            </Anchor>
          </Stack>
        ),
        color: "teal",
        position: "top-right",
      });

      // Обновляем доступные даты
      await loadScheduleEvents(scheduleView, scheduleDate);
    } catch (e: any) {
      console.error(e);
      setError(e.message || "Не удалось создать запись");
    } finally {
      setBookingLoading(false);
    }
  };

  // Формирование данных для Select
  const selectData = users.map((user) => ({
    value: user.id,
    label: `${user.name} (${user.email})`,
  }));

  return (
    <Stack gap="md" data-testid="users-page">
      <Title order={2} data-testid="users-title">
        Запись на встречу
      </Title>

      <Stepper active={bookingStep} allowNextStepsSelect={false}>
        <Stepper.Step label="Шаг 1" description="Выберите пользователя" />
        <Stepper.Step label="Шаг 2" description="Выберите слот" />
        <Stepper.Step label="Шаг 3" description="Подтвердите запись" />
      </Stepper>

      {error && (
        <Alert color="red" onClose={() => setError(null)} withCloseButton>
          {error}
        </Alert>
      )}

      <Select
        label="Выберите пользователя"
        placeholder="Начните вводить имя или email"
        description={
          !selectedUserId ? (
            <Paper className="users-select-hint" shadow="sm" p="md" radius="md" mt={6}>
              <Group gap="sm" align="flex-start" wrap="nowrap">
                <ThemeIcon variant="filled" color="orange" size="lg" radius="md" aria-hidden>
                  <IconInfoCircle size={22} stroke={1.5} />
                </ThemeIcon>
                <Stack gap={6} style={{ flex: 1, minWidth: 0 }}>
                  <Text size="sm" fw={700} className="users-select-hint-title">
                    Кого можно найти в списке
                  </Text>
                  <Text size="sm" lh={1.6} fw={500} className="users-select-hint-body">
                    Поиск осуществляется по пользователям с публичным профилем и по пользователям, у которых
                    вы состоите в группе.
                  </Text>
                </Stack>
              </Group>
            </Paper>
          ) : undefined
        }
        data={selectData}
        value={selectedUserId}
        onChange={setSelectedUserId}
        searchable
        clearable
        disabled={usersLoading}
        data-testid="users-select"
      />

      {usersLoading && (
        <Center h="100px">
          <Skeleton h={20} w="60%" />
        </Center>
      )}

      {userLoading && (
        <Stack gap="sm">
          <Skeleton h={70} />
          <Skeleton h={320} />
        </Stack>
      )}

      {selectedUser && !userLoading && (
        <>
          <Card withBorder data-testid={`user-card-${selectedUser.id}`}>
            <Text fw={500} size="lg" data-testid={`user-name-${selectedUser.id}`}>
              {selectedUser.name}
            </Text>
            <Text c="dimmed" data-testid={`user-email-${selectedUser.id}`}>
              {selectedUser.email}
            </Text>
          </Card>

          <Paper p="md" withBorder>
            <Group justify="space-between" mb="sm">
              <Text size="sm" c="dimmed">
                Прошедшие даты и слоты недоступны для записи.
              </Text>
            </Group>
            {datesLoading || slotsLoading ? (
              <Stack gap="sm">
                <Skeleton h={36} />
                <Skeleton h={300} />
              </Stack>
            ) : (
              <>
                <ScheduleHeader>
                  <Tooltip
                    label="Нельзя перейти к полностью прошедшему периоду"
                    disabled={canGoToPreviousPeriod}
                  >
                    <div>
                      <ScheduleHeader.Previous
                        disabled={!canGoToPreviousPeriod}
                        onClick={() => {
                          if (!canGoToPreviousPeriod) return;
                          const newDate = getPreviousDate(scheduleDate, scheduleView);
                          setScheduleDate(newDate);
                          handleDateChange(newDate);
                        }}
                      />
                    </div>
                  </Tooltip>
                  <ScheduleHeader.Control interactive={false} miw={200}>
                    {getHeaderLabel(scheduleDate, scheduleView)}
                  </ScheduleHeader.Control>
                  <ScheduleHeader.Next 
                    onClick={() => {
                      const newDate = getNextDate(scheduleDate, scheduleView);
                      setScheduleDate(newDate);
                      handleDateChange(newDate);
                    }} 
                  />
                  <ScheduleHeader.Today 
                    onClick={() => {
                      const today = dayjs().format('YYYY-MM-DD');
                      setScheduleDate(today);
                      handleDateChange(today);
                    }} 
                  >
                    Сегодня
                  </ScheduleHeader.Today>
                  <div style={{ marginInlineStart: 'auto' }}>
                    <SegmentedControl
                      value={scheduleView}
                      onChange={(v) => handleViewChange(v as ScheduleViewLevel)}
                      data={[
                        { label: 'Неделя', value: 'week' },
                        { label: 'Месяц', value: 'month' },
                        { label: 'Год', value: 'year' },
                      ]}
                    />
                  </div>
                </ScheduleHeader>

                <div ref={scheduleContainerRef}>
                  <Schedule
                    view={scheduleView}
                    onViewChange={handleViewChange}
                    date={scheduleDate}
                    onDateChange={handleDateChange}
                    labels={{
                      day: "День",
                      week: "Неделя",
                      month: "Месяц",
                      year: "Год",
                      today: "Сегодня",
                      next: "Вперед",
                      previous: "Назад",
                      allDay: "Весь день",
                      weekday: "День недели",
                      timeSlot: "Временной слот",
                      selectMonth: "Выбрать месяц",
                      selectYear: "Выбрать год",
                      switchToDayView: "Переключить на день",
                      switchToWeekView: "Переключить на неделю",
                      switchToMonthView: "Переключить на месяц",
                      switchToYearView: "Переключить на год",
                      viewSelectLabel: "Вид календаря",
                      noEvents: "Нет событий",
                      more: "Еще",
                      moreLabel: (count) => `+${count} еще`,
                    }}
                    events={scheduleEvents}
                    onEventClick={handleEventClick}
                    renderEventBody={renderScheduleEventBody}
                    dayViewProps={{
                      withHeader: false,
                    }}
                    monthViewProps={{
                      withHeader: false,
                      maxEventsPerDay: 2,
                    }}
                    weekViewProps={{
                      withHeader: false,
                    }}
                    yearViewProps={{ withHeader: false }}
                  />
                </div>
                <Text size="xs" c="dimmed" mt="xs">
                  День «сегодня» выделен в календаре, выбранная дата отображается в заголовке периода.
                </Text>

                {scheduleView === "week" && weekSlotsGroupedByDate.length > 0 && (
                  <Stack gap="xs" mt="md">
                    <Text size="sm" fw={600}>Слоты по дням недели</Text>
                    {weekSlotsGroupedByDate.map((group) => (
                      <Card key={group.date} withBorder p="sm">
                        <Text size="sm" fw={500} mb={6}>
                          {dayjs(group.date).format("D MMMM")}
                        </Text>
                        <Group gap="xs">
                          {group.slots.map((slot) => (
                            <Button
                              key={getSlotEventId(slot)}
                              size="xs"
                              variant="light"
                              onClick={() => {
                                setSelectedSlot(slot);
                                setModalOpen(true);
                              }}
                            >
                              {slot.startTime} - {slot.endTime}
                            </Button>
                          ))}
                        </Group>
                      </Card>
                    ))}
                  </Stack>
                )}
              </>
            )}

            {scheduleView === 'day' && daySlots.length === 0 && !slotsLoading && (
              <Stack gap="sm" align="center" py="xl">
                <Text c="dimmed">Нет доступных слотов на этот день</Text>
                <Group>
                  <Button
                    size="xs"
                    variant="default"
                    onClick={() => {
                      const today = dayjs().format("YYYY-MM-DD");
                      setScheduleDate(today);
                      handleDateChange(today);
                    }}
                  >
                    Выбрать другую дату
                  </Button>
                  <Button
                    size="xs"
                    variant="light"
                    onClick={() => setSelectedUserId(null)}
                  >
                    Сменить пользователя
                  </Button>
                </Group>
              </Stack>
            )}
          </Paper>
        </>
      )}

      <BookingConfirmationModal
        opened={modalOpen}
        onClose={() => setModalOpen(false)}
        slot={selectedSlot}
        userName={selectedUser?.name || ""}
        onConfirm={handleBookingConfirm}
        loading={bookingLoading}
      />
    </Stack>
  );
}
