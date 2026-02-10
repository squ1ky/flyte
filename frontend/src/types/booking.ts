export const BookingStatus = {
    PENDING: 'PENDING',
    CONFIRMED: 'CONFIRMED',
    CANCELLED: 'CANCELLED',
    FAILED: 'FAILED'
} as const;

export type BookingStatusType = typeof BookingStatus[keyof typeof BookingStatus]

export interface Booking {
    id: string;
    flight_id: number;
    seat_number: string;
    passenger_name: string;
    passenger_passport: string;
    status: BookingStatusType | string;
    price_cents: number;
    currency: string;
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

export interface CreateBookingResponse {
    booking_id: string;
}