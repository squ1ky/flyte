ALTER TABLE seats
ADD COLUMN reserved_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;

CREATE INDEX idx_seats_reservation_cleanup
ON seats (reserved_at)
WHERE is_booked = TRUE;