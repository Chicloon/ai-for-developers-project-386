"use client";

import { useState, useEffect } from 'react';
import { Paper, Text, Stack, Title, Loader, Center } from '@mantine/core';
import { DatePickerInput } from '@mantine/dates';
import { getSlots, createBooking, getBookings, cancelBooking, Slot, Booking } from '@/lib/api';
import SlotPicker from '@/components/SlotPicker';
import BookingForm from '@/components/BookingForm';
import BookingList from '@/components/BookingList';
import { useRouter } from 'next/navigation';

export default function Home() {
  const router = useRouter();
  const [mode, setMode] = useState<'client' | 'admin'>('client');
  const [selectedDate, setSelectedDate] = useState<string | null>(
    new Date().toISOString().split('T')[0]
  );
  const [slots, setSlots] = useState<Slot[]>([]);
  const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null);
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const saved = localStorage.getItem('booking-mode') as 'client' | 'admin';
    if (saved === 'admin') {
      router.push('/admin');
      return;
    }
    setMode('client');
    setLoading(false);
  }, [router]);

  useEffect(() => {
    if (mode !== 'client' || !selectedDate) return;
    loadSlots();
    loadBookings();
  }, [mode, selectedDate]);

  const loadSlots = async () => {
    try {
      setLoading(true);
      const data = await getSlots(dateStr);
      setSlots(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const loadBookings = async () => {
    try {
      const data = await getBookings();
      setBookings(data);
    } catch (e) {
      console.error(e);
    }
  };

  const handleBooking = async (name: string, email: string) => {
    if (!selectedSlot) return;
    await createBooking({
      slotDate: selectedSlot.date,
      slotStartTime: selectedSlot.startTime,
      name,
      email,
    });
    setSelectedSlot(null);
    await loadSlots();
    await loadBookings();
  };

  const handleCancel = async (id: string) => {
    await cancelBooking(id);
    await loadBookings();
    await loadSlots();
  };

  if (mode !== 'client') {
    return null;
  }

  const dateStr = selectedDate || '';

  return (
    <Stack gap="xl">
      <Paper p="md" withBorder>
        <Title order={4} mb="md">Выберите дату</Title>
        <DatePickerInput
          label="Дата"
          placeholder="Выберите дату"
          value={selectedDate}
          onChange={setSelectedDate}
          locale="ru"
          weekendDays={[]}
          minDate={new Date()}
        />
      </Paper>

      <Paper p="md" withBorder>
        <Title order={4} mb="md">
          Доступное время на {dateStr}
        </Title>
        {loading ? (
          <Center>
            <Loader />
          </Center>
        ) : selectedSlot ? (
          <BookingForm
            slot={selectedSlot}
            onSubmit={handleBooking}
            onCancel={() => setSelectedSlot(null)}
          />
        ) : (
          <SlotPicker
            slots={slots}
            selectedSlot={selectedSlot}
            onSelect={setSelectedSlot}
          />
        )}
      </Paper>

      <Paper p="md" withBorder>
        <Title order={4} mb="md">Предстоящие встречи</Title>
        <BookingList bookings={bookings} onCancel={handleCancel} />
      </Paper>
    </Stack>
  );
}
