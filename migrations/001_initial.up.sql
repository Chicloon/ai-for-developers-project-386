CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY NOT NULL,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  name TEXT NOT NULL,
  is_public INTEGER NOT NULL DEFAULT 0 CHECK (is_public IN (0, 1)),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_users_is_public ON users(is_public) WHERE is_public = 1;

CREATE TABLE IF NOT EXISTS visibility_groups (
  id TEXT PRIMARY KEY NOT NULL,
  owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  visibility_level TEXT NOT NULL CHECK (visibility_level IN ('family', 'work', 'friends')),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS group_members (
  id TEXT PRIMARY KEY NOT NULL,
  group_id TEXT NOT NULL REFERENCES visibility_groups(id) ON DELETE CASCADE,
  member_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  added_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  added_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  UNIQUE(group_id, member_id)
);

CREATE TABLE IF NOT EXISTS schedules (
  id TEXT PRIMARY KEY NOT NULL,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('recurring', 'one-time')),
  day_of_week INTEGER CHECK (day_of_week BETWEEN 0 AND 6),
  date TEXT,
  start_time TEXT NOT NULL,
  end_time TEXT NOT NULL,
  is_blocked INTEGER NOT NULL DEFAULT 0 CHECK (is_blocked IN (0, 1)),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  CHECK (end_time > start_time),
  CHECK ((type = 'recurring' AND day_of_week IS NOT NULL AND date IS NULL) OR (type = 'one-time' AND date IS NOT NULL AND day_of_week IS NULL))
);

CREATE INDEX IF NOT EXISTS idx_schedules_user_id ON schedules(user_id);
CREATE INDEX IF NOT EXISTS idx_schedules_user_date ON schedules(user_id, date);

CREATE TABLE IF NOT EXISTS schedule_visibility_groups (
  id TEXT PRIMARY KEY NOT NULL,
  schedule_id TEXT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  group_id TEXT NOT NULL REFERENCES visibility_groups(id) ON DELETE CASCADE,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  UNIQUE(schedule_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_schedule_visibility_schedule_id ON schedule_visibility_groups(schedule_id);
CREATE INDEX IF NOT EXISTS idx_schedule_visibility_group_id ON schedule_visibility_groups(group_id);

CREATE TABLE IF NOT EXISTS bookings (
  id TEXT PRIMARY KEY NOT NULL,
  schedule_id TEXT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  booker_id TEXT NOT NULL REFERENCES users(id),
  owner_id TEXT NOT NULL REFERENCES users(id),
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  cancelled_at TEXT,
  cancelled_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  slot_date TEXT,
  slot_start_time TEXT
);

CREATE INDEX IF NOT EXISTS idx_bookings_booker_id ON bookings(booker_id);
CREATE INDEX IF NOT EXISTS idx_bookings_owner_id ON bookings(owner_id);
CREATE INDEX IF NOT EXISTS idx_bookings_schedule_id ON bookings(schedule_id);
CREATE INDEX IF NOT EXISTS idx_group_members_member_id ON group_members(member_id);
CREATE INDEX IF NOT EXISTS idx_visibility_groups_owner_id ON visibility_groups(owner_id);
