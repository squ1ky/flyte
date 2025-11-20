INSERT INTO airports (code, name, city, country)
VALUES
    ('DME', 'Domodedovo Moscow Airport', 'Moscow', 'Russia'),
    ('IST', 'Istanbul Airport', 'Istanbul', 'Turkey');

INSERT INTO aircrafts (code, name, total_seats)
VALUES
    ( 'A320-180', 'Airbus A320, 180 seats', 180 ),
    ( 'B738-189', 'Boeing 737-800, 189 seats', 189 );

INSERT INTO aircraft_seats (
    aircraft_id, row_number, seat_column, seat_number,
    cabin_class, is_window, is_aisle, is_exit_row
)
VALUES
    -- A320, row 1
    (1, 1, 'A', '1A', 'BUSINESS', TRUE,  FALSE, FALSE),
    (1, 1, 'B', '1B', 'BUSINESS', FALSE, FALSE, FALSE),
    (1, 1, 'C', '1C', 'BUSINESS', FALSE, TRUE,  FALSE),
    (1, 1, 'D', '1D', 'BUSINESS', FALSE, TRUE,  FALSE),
    (1, 1, 'E', '1E', 'BUSINESS', FALSE, FALSE, FALSE),
    (1, 1, 'F', '1F', 'BUSINESS', TRUE,  FALSE, FALSE),

    -- A320, row 2 (эконом)
    (1, 2, 'A', '2A', 'ECONOMY', TRUE,  FALSE, FALSE),
    (1, 2, 'B', '2B', 'ECONOMY', FALSE, FALSE, FALSE),
    (1, 2, 'C', '2C', 'ECONOMY', FALSE, TRUE,  FALSE),
    (1, 2, 'D', '2D', 'ECONOMY', FALSE, TRUE,  FALSE),
    (1, 2, 'E', '2E', 'ECONOMY', FALSE, FALSE, FALSE),
    (1, 2, 'F', '2F', 'ECONOMY', TRUE,  FALSE, FALSE);

INSERT INTO flights (
    flight_number,
    origin_airport_id,
    destination_airport_id,
    departure_time,
    arrival_time,
    base_price,
    currency,
    status,
    aircraft_id
)
VALUES
    ('TK401', 1, 2, '2025-12-10 06:30:00', '2025-12-10 09:10:00', 12000.00, 'RUB', 'SCHEDULED', 1),
    ('TK402', 2, 1, '2025-12-10 20:45:00', '2025-12-11 00:15:00', 13500.00, 'RUB', 'SCHEDULED', 1);
