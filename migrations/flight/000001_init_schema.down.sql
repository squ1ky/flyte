DROP INDEX IF EXISTS idx_outbox_status_created;
DROP TABLE IF EXISTS flight_outbox;

DROP INDEX IF EXISTS idx_seats_reservation_cleanup;
DROP TABLE IF EXISTS seats;

DROP INDEX IF EXISTS idx_flights_route;
DROP TABLE IF EXISTS flights;

DROP TABLE IF EXISTS aircraft_seats;
DROP TABLE IF EXISTS aircrafts;
DROP TABLE IF EXISTS airports;
