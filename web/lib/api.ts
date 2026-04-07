/** Same-origin `/api/*` only — never use localhost or absolute URLs in the browser (breaks on public HTTP origins). */

// Auth types
export interface User {
  id: string;
  email: string;
  name: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// Token storage
let authToken: string | null = null;

export function setAuthToken(token: string | null) {
  authToken = token;
  if (token) {
    localStorage.setItem('auth_token', token);
  } else {
    localStorage.removeItem('auth_token');
  }
}

export function getAuthToken(): string | null {
  if (!authToken) {
    authToken = localStorage.getItem('auth_token');
  }
  return authToken;
}

// Auth API
export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const res = await fetch("/api/auth/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || "Registration failed");
  }
  const result = await res.json();
  setAuthToken(result.token);
  return result;
}

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const res = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || "Login failed");
  }
  const result = await res.json();
  setAuthToken(result.token);
  return result;
}

export async function getMe(): Promise<User> {
  const token = getAuthToken();
  const res = await fetch("/api/auth/me", {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error("Failed to get user");
  return res.json();
}

export function logout() {
  setAuthToken(null);
}

export interface TimeRange {
  startTime: string;
  endTime: string;
}

export interface AvailabilityRule {
  id: string;
  type: "recurring" | "one-time";
  dayOfWeek?: number;
  date?: string;
  timeRanges: TimeRange[];
}

export interface CreateAvailabilityRule {
  type: "recurring" | "one-time";
  dayOfWeek?: number;
  date?: string;
  timeRanges: TimeRange[];
}

export interface BlockedDay {
  id: string;
  date: string;
}

export interface Slot {
  id: string;
  date: string;
  startTime: string;
  endTime: string;
  isBooked: boolean;
}

export interface Booking {
  id: string;
  slotDate: string;
  slotStartTime: string;
  name: string;
  email: string;
  status: string;
  recurrence?: string;
  dayOfWeek?: number;
  endDate?: string;
}

export async function getSlots(date: string): Promise<Slot[]> {
  const res = await fetch(`/api/slots?date=${encodeURIComponent(date)}`);
  if (!res.ok) throw new Error("Failed to fetch slots");
  return res.json();
}

export async function createBooking(data: {
  slotDate: string;
  slotStartTime: string;
  name: string;
  email: string;
}): Promise<Booking> {
  const res = await fetch("/api/bookings", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create booking");
  return res.json();
}

export async function getBookings(): Promise<Booking[]> {
  const res = await fetch("/api/bookings");
  if (!res.ok) throw new Error("Failed to fetch bookings");
  return res.json();
}

export async function cancelBooking(id: string): Promise<void> {
  const res = await fetch(`/api/bookings/${encodeURIComponent(id)}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Failed to cancel booking");
}

export async function getAvailabilityRules(): Promise<AvailabilityRule[]> {
  const res = await fetch("/api/availability-rules");
  if (!res.ok) throw new Error("Failed to fetch availability rules");
  return res.json();
}

export async function createAvailabilityRule(data: CreateAvailabilityRule): Promise<AvailabilityRule> {
  const res = await fetch("/api/availability-rules", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create availability rule");
  return res.json();
}

export async function deleteAvailabilityRule(id: string): Promise<void> {
  const res = await fetch(`/api/availability-rules/${encodeURIComponent(id)}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Failed to delete availability rule");
}

export async function getBlockedDays(): Promise<BlockedDay[]> {
  const res = await fetch("/api/blocked-days");
  if (!res.ok) throw new Error("Failed to fetch blocked days");
  return res.json();
}

export async function createBlockedDay(date: string): Promise<BlockedDay> {
  const res = await fetch("/api/blocked-days", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ date }),
  });
  if (!res.ok) throw new Error("Failed to create blocked day");
  return res.json();
}

export async function deleteBlockedDay(id: string): Promise<void> {
  const res = await fetch(`/api/blocked-days/${encodeURIComponent(id)}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Failed to delete blocked day");
}
