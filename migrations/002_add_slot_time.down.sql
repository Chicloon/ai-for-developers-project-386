-- Revert slot_start_time addition
DROP INDEX IF EXISTS idx_bookings_schedule_slot;
ALTER TABLE bookings DROP COLUMN IF EXISTS slot_start_time;
