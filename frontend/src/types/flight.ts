export type SeatClass = 'economy' | 'business' | 'first'

export type FlightStatus = 'scheduled' | 'departed' | 'arrived' | 'cancelled'

export interface AirportResponse {
  code: string
  name: string
  city: string
  country: string
  /** IANA timezone, e.g. "Europe/Moscow" — use for local time display */
  timezone: string
}

export interface AirlineResponse {
  id: number
  iata_code: string
  name: string
  country: string
  logo_url: string
}

export interface AircraftResponse {
  id: number
  model: string
  total_seats: number
}

/** Template seat layout for an aircraft — not tied to a specific flight */
export interface SeatResponse {
  seat_number: string
  seat_class: SeatClass
  row: number
  col: number
}

export interface FareResponse {
  seat_class: SeatClass
  /** Price in kopecks/cents — divide by 100 for display */
  price_cents: number
}

export interface FlightResponse {
  id: number
  flight_number: string
  airline: AirlineResponse
  aircraft: AircraftResponse
  departure_airport: AirportResponse
  arrival_airport: AirportResponse
  /** ISO 8601 UTC */
  departure_time: string
  /** ISO 8601 UTC */
  arrival_time: string
  status: FlightStatus
  /** Minimum price across all seat classes, in kopecks/cents */
  min_price_cents: number
  available_seats: number
  currency: string
}

// --- Admin request types ---

export interface CreateFlightRequest {
  flight_number: string
  airline_id: number
  aircraft_id: number
  departure_airport_code: string
  arrival_airport_code: string
  /** ISO 8601 UTC */
  departure_time: string
  /** ISO 8601 UTC */
  arrival_time: string
}

export interface UpdateFlightStatusRequest {
  status: FlightStatus
}

export interface CreateAirlineRequest {
  iata_code: string
  name: string
  country: string
  logo_url: string
}

export interface CreateAircraftRequest {
  model: string
  total_seats: number
}

export interface CreateSeatRequest {
  seat_number: string
  seat_class: SeatClass
  row: number
  col: number
}
