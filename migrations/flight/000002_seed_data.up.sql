INSERT INTO airports (code, name, city, country, timezone)
VALUES ('LED', 'Pulkovo Airport', 'St. Petersburg', 'Russia', 'Europe/Moscow'),
       ('DME', 'Domodedovo International Airport', 'Moscow', 'Russia', 'Europe/Moscow'),
       ('IST', 'Istanbul Airport', 'Istanbul', 'Turkey', 'Europe/Istanbul');

INSERT INTO aircrafts (id, model, total_seats)
VALUES (1, 'Sukhoi Superjet 100', 10);

INSERT INTO aircraft_seats (aircraft_id, seat_number, seat_class, price_multiplier)
VALUES (1, '1A', 'business', 2.5),
       (1, '1C', 'business', 2.5),
       (1, '2A', 'economy', 1.0),
       (1, '2C', 'economy', 1.0),
       (1, '3A', 'economy', 1.0),
       (1, '3C', 'economy', 1.0),
       (1, '4A', 'economy', 1.0),
       (1, '4C', 'economy', 1.0),
       (1, '5A', 'economy', 1.0),
       (1, '5C', 'economy', 1.0);

INSERT INTO flights (id, flight_number, aircraft_id, departure_airport, arrival_airport, departure_time, arrival_time,
                     base_price_cents, status)
VALUES (1, 'SU-100', 1, 'LED', 'DME', NOW() + interval '1 day', NOW() + interval '1 day 1 hour 30 minutes', 500000,
        'scheduled'),
       (2, 'SU-101', 1, 'DME', 'LED', NOW() + interval '2 days', NOW() + interval '2 days 1 hour 30 minutes', 450000,
        'scheduled'),
       (3, 'SU-200', 1, 'DME', 'IST', NOW() + interval '3 days', NOW() + interval '3 days 4 hours', 1500000,
        'scheduled'),
       (4, 'SU-201', 1, 'IST', 'DME', NOW() + interval '4 days', NOW() + interval '4 days 4 hours', 1400000,
        'scheduled'),
       (5, 'SU-300', 1, 'LED', 'IST', NOW() + interval '5 days', NOW() + interval '5 days 5 hours', 1800000,
        'scheduled');

INSERT INTO seats (flight_id, seat_number, seat_class, price_multiplier, is_booked)
SELECT f.id,
       s.seat_number,
       s.seat_class,
       s.price_multiplier,
       FALSE
FROM flights f
         CROSS JOIN aircraft_seats s
WHERE f.aircraft_id = s.aircraft_id;


INSERT INTO flight_outbox (event_type, payload, status)
VALUES ('FLIGHT_CREATED',
        '{"flight_id": 1, "flight_number": "SU-100", "departure_airport": "LED", "arrival_airport": "DME"}', 'PENDING'),
       ('FLIGHT_CREATED',
        '{"flight_id": 2, "flight_number": "SU-101", "departure_airport": "DME", "arrival_airport": "LED"}', 'PENDING'),
       ('FLIGHT_CREATED',
        '{"flight_id": 3, "flight_number": "SU-200", "departure_airport": "DME", "arrival_airport": "IST"}', 'PENDING'),
       ('FLIGHT_CREATED',
        '{"flight_id": 4, "flight_number": "SU-201", "departure_airport": "IST", "arrival_airport": "DME"}', 'PENDING'),
       ('FLIGHT_CREATED',
        '{"flight_id": 5, "flight_number": "SU-300", "departure_airport": "LED", "arrival_airport": "IST"}', 'PENDING');