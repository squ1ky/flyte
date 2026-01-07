INSERT INTO airports (code, name, city, country, timezone)
VALUES ('KZN', 'Kazan International Airport', 'Kazan', 'Russia', 'Europe/Moscow'),
       ('AER', 'Sochi International Airport', 'Sochi', 'Russia', 'Europe/Moscow'),
       ('VKO', 'Vnukovo International Airport', 'Moscow', 'Russia', 'Europe/Moscow'),
       ('MSQ', 'Minsk National Airport', 'Minsk', 'Belarus', 'Europe/Minsk'),
       ('SVX', 'Koltsovo International Airport', 'Yekaterinburg', 'Russia', 'Asia/Yekaterinburg'),
       ('HKT', 'Phuket International Airport', 'Phuket', 'Thailand', 'Asia/Bangkok');

INSERT INTO flights (flight_number, departure_airport, arrival_airport, departure_time, arrival_time, price,
                     total_seats, status)
VALUES
-- Рейс 1: Казань -> Сочи (Утро 25 мая)
('N4-123', 'KZN', 'AER', '2026-05-25 08:30:00', '2026-05-25 12:00:00', 12500.00, 6, 'scheduled'),

-- Рейс 2: Москва -> Минск (Вечер 25 мая)
('B2-999', 'VKO', 'MSQ', '2026-05-25 19:45:00', '2026-05-25 21:10:00', 8500.00, 4, 'scheduled'),

-- Рейс 3: Екб -> Пхукет (Ночь 26 мая)
('ZF-777', 'SVX', 'HKT', '2026-05-26 01:00:00', '2026-05-26 10:30:00', 65000.00, 5, 'scheduled');

-- Казань-Сочи
INSERT INTO seats (flight_id, seat_number, is_booked, price_multiplier)
VALUES (1, '1A', false, 1.5),
       (1, '1B', false, 1.5), -- Comfort
       (1, '5A', false, 1.0),
       (1, '5B', true, 1.0),  -- 5B занято
       (1, '6A', false, 1.0),
       (1, '6B', false, 1.0);

-- Москва-Минск
INSERT INTO seats (flight_id, seat_number, is_booked, price_multiplier)
VALUES (2, '10A', false, 1.0),
       (2, '10B', false, 1.0),
       (2, '11A', false, 1.0),
       (2, '11B', false, 1.0);

-- Екб-Пхукет - Бизнес и эконом
INSERT INTO seats (flight_id, seat_number, is_booked, price_multiplier)
VALUES (3, '1A', false, 2.5), -- Business
       (3, '10A', false, 1.0),
       (3, '10B', false, 1.0),
       (3, '11A', true, 1.0), -- 11A занято
       (3, '11B', false, 1.0);
