/** Same-origin `/api/*` only — never use localhost or absolute URLs in the browser (breaks on public HTTP origins). */

// Auth types
export interface User {
  id: string;
  email: string;
  name: string;
  createdAt?: string;
  updatedAt?: string;
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

// Schedule types (updated from AvailabilityRule)
export interface Schedule {
  id: string;
  userId: string;
  type: "recurring" | "one-time";
  dayOfWeek?: number;
  date?: string;
  startTime: string;
  endTime: string;
  isBlocked: boolean;
  createdAt?: string;
}

export interface CreateScheduleRequest {
  type: "recurring" | "one-time";
  dayOfWeek?: number;
  date?: string;
  startTime: string;
  endTime: string;
  isBlocked?: boolean;
}

export interface SchedulesListResponse {
  schedules: Schedule[];
}

// Visibility Group types
export interface VisibilityGroup {
  id: string;
  ownerId: string;
  name: string;
  visibilityLevel: "family" | "work" | "friends" | "public";
  createdAt?: string;
}

export interface CreateGroupRequest {
  name: string;
  visibilityLevel: "family" | "work" | "friends" | "public";
}

export interface GroupsListResponse {
  groups: VisibilityGroup[];
}

// Group Member types
export interface GroupMember {
  id: string;
  groupId: string;
  member: User;
  addedBy: string;
  addedAt: string;
}

export interface AddMemberRequest {
  email?: string;
  userId?: string;
}

export interface MembersListResponse {
  members: GroupMember[];
}

// Booking types (updated)
export interface Booking {
  id: string;
  scheduleId: string;
  booker: User;
  owner: User;
  date: string;
  startTime: string;
  endTime: string;
  status: "active" | "cancelled";
  createdAt?: string;
  cancelledAt?: string;
}

export interface CreateBookingRequest {
  ownerId: string;
  scheduleId: string;
  slotStartTime: string; // Time of the specific 30-min slot (e.g., "09:00")
}

export interface BookingsListResponse {
  bookings: Booking[];
}

// Slot type
export interface Slot {
  id: string;
  date: string;
  startTime: string;
  endTime: string;
  isBooked: boolean;
}

export interface SlotsListResponse {
  slots: Slot[];
}

export interface UsersListResponse {
  users: User[];
}

// Token storage
let authToken: string | null = null;

export function setAuthToken(token: string | null) {
  authToken = token;
  if (token) {
    localStorage.setItem("auth_token", token);
  } else {
    localStorage.removeItem("auth_token");
  }
}

export function getAuthToken(): string | null {
  if (!authToken) {
    authToken = localStorage.getItem("auth_token");
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
    headers["Authorization"] = `Bearer ${token}`;
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
  // #region agent log H1,H5
  fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H5',location:'api.ts:login:entry',message:'Login API call starting',data:{email:data.email},timestamp:Date.now()})}).catch(()=>{});
  // #endregion
  const res = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  // #region agent log H1,H5
  fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H1',location:'api.ts:login:response',message:'Login API response received',data:{status:res.status,statusText:res.statusText,ok:res.ok},timestamp:Date.now()})}).catch(()=>{});
  // #endregion
  if (!res.ok) {
    const err = await res.json();
    // #region agent log H1
    fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H1',location:'api.ts:login:error',message:'Login API returned error',data:{error:err.error},timestamp:Date.now()})}).catch(()=>{});
    // #endregion
    throw new Error(err.error || "Login failed");
  }
  const result = await res.json();
  // #region agent log H2
  fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'eb49d8'},body:JSON.stringify({sessionId:'eb49d8',runId:'debug1',hypothesisId:'H2',location:'api.ts:login:success',message:'Login API success, parsing response',data:{hasToken:!!result.token,tokenLength:result.token?.length,hasUser:!!result.user,userId:result.user?.id},timestamp:Date.now()})}).catch(()=>{});
  // #endregion
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
export async function getUsers(): Promise<UsersListResponse> {
  const res = await authFetch("/api/users");
  if (!res.ok) throw new Error("Failed to fetch users");
  return res.json();
}

export async function getUser(id: string): Promise<User> {
  const res = await authFetch(`/api/users/${encodeURIComponent(id)}`);
  if (!res.ok) throw new Error("Failed to fetch user");
  return res.json();
}

export async function getUserSlots(
  userId: string,
  date: string
): Promise<SlotsListResponse> {
  const res = await authFetch(
    `/api/users/${encodeURIComponent(userId)}/slots?date=${encodeURIComponent(date)}`
  );
  if (!res.ok) throw new Error("Failed to fetch user slots");
  return res.json();
}

// Schedules API
export async function getMySchedules(): Promise<SchedulesListResponse> {
  const res = await authFetch("/api/my/schedules");
  if (!res.ok) throw new Error("Failed to fetch schedules");
  return res.json();
}

export async function createSchedule(
  data: CreateScheduleRequest
): Promise<Schedule> {
  const res = await authFetch("/api/my/schedules", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create schedule");
  return res.json();
}

export async function updateSchedule(
  id: string,
  data: CreateScheduleRequest
): Promise<Schedule> {
  const res = await authFetch(`/api/my/schedules/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to update schedule");
  return res.json();
}

export async function deleteSchedule(id: string): Promise<void> {
  const res = await authFetch(`/api/my/schedules/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Failed to delete schedule");
}

// Groups API
export async function getMyGroups(): Promise<GroupsListResponse> {
  const res = await authFetch("/api/my/groups");
  if (!res.ok) throw new Error("Failed to fetch groups");
  return res.json();
}

export async function createGroup(
  data: CreateGroupRequest
): Promise<VisibilityGroup> {
  const res = await authFetch("/api/my/groups", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create group");
  return res.json();
}

export async function updateGroup(
  id: string,
  data: CreateGroupRequest
): Promise<VisibilityGroup> {
  const res = await authFetch(`/api/my/groups/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to update group");
  return res.json();
}

export async function deleteGroup(id: string): Promise<void> {
  const res = await authFetch(`/api/my/groups/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Failed to delete group");
}

export async function getGroupMembers(
  groupId: string
): Promise<MembersListResponse> {
  const res = await authFetch(
    `/api/my/groups/${encodeURIComponent(groupId)}/members`
  );
  if (!res.ok) throw new Error("Failed to fetch group members");
  return res.json();
}

export async function addGroupMember(
  groupId: string,
  data: AddMemberRequest
): Promise<GroupMember> {
  const res = await authFetch(
    `/api/my/groups/${encodeURIComponent(groupId)}/members`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    }
  );
  if (!res.ok) throw new Error("Failed to add group member");
  return res.json();
}

export async function removeGroupMember(
  groupId: string,
  memberId: string
): Promise<void> {
  const res = await authFetch(
    `/api/my/groups/${encodeURIComponent(groupId)}/members/${encodeURIComponent(memberId)}`,
    {
      method: "DELETE",
    }
  );
  if (!res.ok) throw new Error("Failed to remove group member");
}

// Bookings API
export async function getMyBookings(): Promise<BookingsListResponse> {
  const res = await authFetch("/api/my/bookings");
  if (!res.ok) throw new Error("Failed to fetch bookings");
  return res.json();
}

export async function createBooking(
  data: CreateBookingRequest
): Promise<Booking> {
  // #region agent log H4
  fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'221396'},body:JSON.stringify({sessionId:'221396',runId:'debug1',hypothesisId:'H4',location:'api.ts:createBooking:entry',message:'Create booking request',data:{ownerId:data.ownerId,scheduleId:data.scheduleId},timestamp:Date.now()})}).catch(()=>{});
  // #endregion
  const res = await authFetch("/api/my/bookings", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  // #region agent log H4
  fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'221396'},body:JSON.stringify({sessionId:'221396',runId:'debug1',hypothesisId:'H4',location:'api.ts:createBooking:response',message:'Booking response received',data:{status:res.status,statusText:res.statusText},timestamp:Date.now()})}).catch(()=>{});
  // #endregion
  if (!res.ok) {
    const resText = await res.text();
    // #region agent log H4
    fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'221396'},body:JSON.stringify({sessionId:'221396',runId:'debug1',hypothesisId:'H4',location:'api.ts:createBooking:error',message:'Booking response not OK',data:{status:res.status,responsePreview:resText.substring(0,500)},timestamp:Date.now()})}).catch(()=>{});
    // #endregion
    let errMsg = "Failed to create booking";
    try {
      const err = JSON.parse(resText);
      errMsg = err.error || errMsg;
    } catch (e) {
      errMsg = resText || errMsg;
    }
    throw new Error(errMsg);
  }
  const result = await res.json();
  // #region agent log H4
  fetch('http://127.0.0.1:7924/ingest/df065418-75a6-4c94-b505-bfe4e2e4e84a',{method:'POST',headers:{'Content-Type':'application/json','X-Debug-Session-Id':'221396'},body:JSON.stringify({sessionId:'221396',runId:'debug1',hypothesisId:'H4',location:'api.ts:createBooking:success',message:'Booking created successfully',data:{bookingId:result.id},timestamp:Date.now()})}).catch(()=>{});
  // #endregion
  return result;
}

export async function cancelBooking(id: string): Promise<void> {
  const res = await authFetch(`/api/my/bookings/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Failed to cancel booking");
}
