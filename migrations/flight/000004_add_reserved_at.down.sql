DROP INDEX IF EXISTS idx_seats_reservation_cleanup;

ALTER TABLE seats
DROP COLUMN IF EXISTS reserved_at;