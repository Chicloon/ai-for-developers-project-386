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

// Schedule types
export interface TimeRange {
  startTime: string;
  endTime: string;
}

export interface Schedule {
  id: string;
  type: "recurring" | "one-time";
  dayOfWeek?: number;
  date?: string;
  timeRanges: TimeRange[];
}

export interface CreateScheduleRequest {
  type: "recurring" | "one-time";
  dayOfWeek?: number;
  date?: string;
  timeRanges: TimeRange[];
}

// Group types
export interface Group {
  id: string;
  name: string;
  ownerId: string;
}

export interface CreateGroupRequest {
  name: string;
}

// Booking types
export interface Booking {
  id: string;
  slotDate: string;
  slotStartTime: string;
  slotEndTime?: string;
  name?: string;
  email?: string;
  status: string;
  recurrence?: string;
  dayOfWeek?: number;
  endDate?: string;
  bookerId?: string;
  ownerId?: string;
}

export interface CreateBookingRequest {
  slotDate: string;
  slotStartTime: string;
  ownerId?: string;
  name?: string;
  email?: string;
}

// Slot type
export interface Slot {
  id: string;
  date: string;
  startTime: string;
  endTime: string;
  isBooked: boolean;
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

// Helper function for authenticated requests
async function authFetch(url: string, options?: RequestInit): Promise<Response> {
  const token = getAuthToken();
  const headers: Record<string, string> = {
    ...(options?.headers as Record<string, string>),
  };
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  return fetch(url, {
    ...options,
    headers,
  });
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
  const res = await authFetch("/api/auth/me");
  if (!res.ok) throw new Error("Failed to get user");
  return res.json();
}

export function logout() {
  setAuthToken(null);
}

// Users API
export async function getUsers(): Promise<User[]> {
  const res = await authFetch("/api/users");
  if (!res.ok) throw new Error("Failed to fetch users");
  return res.json();
}

export async function getUser(id: string): Promise<User> {
  const res = await authFetch(`/api/users/${encodeURIComponent(id)}`);
  if (!res.ok) throw new Error("Failed to fetch user");
  return res.json();
}

export async function getUserSlots(userId: string, date: string): Promise<Slot[]> {
  const res = await authFetch(`/api/users/${encodeURIComponent(userId)}/slots?date=${encodeURIComponent(date)}`);
  if (!res.ok) throw new Error("Failed to fetch user slots");
  return res.json();
}

// Schedules API
export async function getMySchedules(): Promise<Schedule[]> {
  const res = await authFetch("/api/my/schedules");
  if (!res.ok) throw new Error("Failed to fetch schedules");
  return res.json();
}

export async function createSchedule(data: CreateScheduleRequest): Promise<Schedule> {
  const res = await authFetch("/api/my/schedules", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create schedule");
  return res.json();
}

export async function deleteSchedule(id: string): Promise<void> {
  const res = await authFetch(`/api/my/schedules/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Failed to delete schedule");
}

// Groups API
export async function getMyGroups(): Promise<Group[]> {
  const res = await authFetch("/api/my/groups");
  if (!res.ok) throw new Error("Failed to fetch groups");
  return res.json();
}

export async function createGroup(data: CreateGroupRequest): Promise<Group> {
  const res = await authFetch("/api/my/groups", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create group");
  return res.json();
}

export async function deleteGroup(id: string): Promise<void> {
  const res = await authFetch(`/api/my/groups/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Failed to delete group");
}

export async function getGroupMembers(groupId: string): Promise<User[]> {
  const res = await authFetch(`/api/groups/${encodeURIComponent(groupId)}/members`);
  if (!res.ok) throw new Error("Failed to fetch group members");
  return res.json();
}

export async function addGroupMember(groupId: string, email: string): Promise<void> {
  const res = await authFetch(`/api/groups/${encodeURIComponent(groupId)}/members`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  if (!res.ok) throw new Error("Failed to add group member");
}

export async function removeGroupMember(groupId: string, userId: string): Promise<void> {
  const res = await authFetch(`/api/groups/${encodeURIComponent(groupId)}/members/${encodeURIComponent(userId)}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Failed to remove group member");
}

// Bookings API
export async function getMyBookings(): Promise<Booking[]> {
  const res = await authFetch("/api/my/bookings");
  if (!res.ok) throw new Error("Failed to fetch bookings");
  return res.json();
}

export async function createBooking(data: CreateBookingRequest): Promise<Booking> {
  const res = await authFetch("/api/bookings", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || "Failed to create booking");
  }
  return res.json();
}

export async function cancelBooking(id: string): Promise<void> {
  const res = await authFetch(`/api/bookings/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Failed to cancel booking");
}

// Legacy API functions (for compatibility)
export async function getSlots(date: string): Promise<Slot[]> {
  const res = await fetch(`/api/slots?date=${encodeURIComponent(date)}`);
  if (!res.ok) throw new Error("Failed to fetch slots");
  return res.json();
}

export async function getBookings(): Promise<Booking[]> {
  const res = await fetch("/api/bookings");
  if (!res.ok) throw new Error("Failed to fetch bookings");
  return res.json();
}

// Legacy availability rule functions
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
