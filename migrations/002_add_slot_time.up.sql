-- Migration from legacy booking schema to new schema
-- Legacy: id, slot_date, slot_start_time, name, email, status, recurrence, day_of_week, end_date, created_at
-- New: id, schedule_id, booker_id, owner_id, status, created_at, cancelled_at, cancelled_by, slot_start_time

-- Add new columns (if they don't exist)
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS schedule_id UUID;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS booker_id UUID;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS owner_id UUID;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMP;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS cancelled_by UUID;

-- Note: slot_start_time already exists in legacy schema!

-- Create indexes for new schema
CREATE INDEX IF NOT EXISTS idx_bookings_schedule_id ON bookings(schedule_id);
CREATE INDEX IF NOT EXISTS idx_bookings_booker_id ON bookings(booker_id);
CREATE INDEX IF NOT EXISTS idx_bookings_owner_id ON bookings(owner_id);
CREATE INDEX IF NOT EXISTS idx_bookings_schedule_slot ON bookings(schedule_id, slot_start_time, status);

-- Make old columns nullable (for gradual migration)
ALTER TABLE bookings ALTER COLUMN slot_date DROP NOT NULL;
ALTER TABLE bookings ALTER COLUMN name DROP NOT NULL;
ALTER TABLE bookings ALTER COLUMN email DROP NOT NULL;
ALTER TABLE bookings ALTER COLUMN recurrence DROP NOT NULL;
