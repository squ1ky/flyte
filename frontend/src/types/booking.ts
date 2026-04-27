import type { SeatClass } from '@/types/flight'

export type BookingStatus = 'pending' | 'paid' | 'cancelled'

export type TicketStatus = 'reserved' | 'issued'

/** Simplified passenger for booking — not the full profile from PassengerInput */
export interface BookingPassengerInput {
  first_name: string
  last_name: string
  document_number: string
  seat_number: string
  seat_class: SeatClass
}

export interface CreateBookingRequest {
  flight_id: number
  contact_email: string
  passengers: BookingPassengerInput[]
}

export interface CreateBookingResponse {
  booking_id: string
  ref_code: string
  /** ISO 8601 UTC — booking expires 15 minutes after creation if unpaid */
  expires_at: string
}

export interface TicketResponse {
  id: string
  passenger_first_name: string
  passenger_last_name: string
  passenger_doc_num: string
  seat_number: string
  seat_class: SeatClass
  /** Price in kopecks/cents — divide by 100 for display */
  price_cents: number
  ticket_number: string
  status: TicketStatus
}

export interface BookingResponse {
  id: string
  user_id: number
  flight_id: number
  ref_code: string
  status: BookingStatus
  /** Total price in kopecks/cents — divide by 100 for display */
  total_price_cents: number
  currency: string
  contact_email: string
  tickets: TicketResponse[]
  /** ISO 8601 UTC — present only when status is "pending" */
  expires_at?: string
  /** ISO 8601 UTC */
  created_at: string
}
