using Http;
using Rest;

namespace CallBooking;

model TimeRange {
  startTime: string;
  endTime: string;
}

model AvailabilityRule {
  id: string;
  type: "recurring" | "one-time";
  dayOfWeek?: int32;
  date?: string;
  timeRanges: TimeRange[];
}

model CreateAvailabilityRule {
  type: "recurring" | "one-time";
  dayOfWeek?: int32;
  date?: string;
  timeRanges: TimeRange[];
}

model BlockedDay {
  id: string;
  date: string;
}

model CreateBlockedDay {
  date: string;
}

model Slot {
  id: string;
  date: string;
  startTime: string;
  endTime: string;
  isBooked: boolean;
}

model Booking {
  id: string;
  slotDate: string;
  slotStartTime: string;
  name: string;
  email: string;
  status: "active" | "cancelled";
  recurrence?: "none" | "daily" | "weekly" | "yearly";
  dayOfWeek?: int32;
  endDate?: string;
}

model CreateBooking {
  slotDate: string;
  slotStartTime: string;
  name: string;
  email: string;
  recurrence?: "none" | "daily" | "weekly" | "yearly";
  dayOfWeek?: int32;
  endDate?: string;
}

model ErrorResponse {
  error: string;
}
