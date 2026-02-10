export interface Airport {
    code: string;
    name: string;
    city: string;
    country: string;
}

export interface Aircraft {
    id: number;
    model: string;
    total_seats: number;
}

export interface Flight {
    id: number;
    flight_number: string;
    // aircraft_id: number;
    departure_airport: string;
    arrival_airport: string;
    departure_time: string;
    arrival_time: string;
    base_price_cents: number;
    status: string;
    total_seats: number;
    available_seats: number;
}

export interface SearchFlightParams {
    from: string;
    to: string;
    date: string;
    passengers: number;
}

export interface CreateFlightRequest {
    flight_number: string;
    aircraft_id: number;
    departure_airport: string;
    arrival_airport: string;
    departure_time: string;
    arrival_time: string;
    base_price_cents: number;
}

export interface Seat {
    id: number
    seat_number: string;
    is_booked: boolean;
    price_multiplier: number;
    // seat_class: string;
}
