export interface Seat {
    id: number;
    seat_number: string;
    is_booked: boolean;
    price_multiplier: number;
}

export interface Booking {
    id: number;
    flight_id: number;
    seat_number: string;
    passenger_name: string;
    price_cents: number;
    status: string;
    created_at: string;
}

export interface CreateBookingRequest {
    flight_id: number;
    seat_number: string;
    passenger_name: string;
    passenger_passport: string;
    price_cents: number;
    currency: string;
}
