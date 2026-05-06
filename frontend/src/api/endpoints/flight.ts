import { api } from '@/api/instance'
import type {
  FlightResponse,
  FareResponse,
  AirportResponse,
  AirlineResponse,
  AircraftResponse,
  SeatResponse,
  SeatClass,
} from '@/types/flight'

export interface SearchFlightsParams {
  from: string
  to: string
  date: string
  seat_class?: SeatClass
}

export interface PaginationParams {
  limit?: number
  offset?: number
}

export async function searchFlights(params: SearchFlightsParams): Promise<FlightResponse[]> {
  const { data } = await api.get<FlightResponse[]>('/flights', { params })
  return data
}

export async function getFlightDetails(id: number): Promise<FlightResponse> {
  const { data } = await api.get<FlightResponse>(`/flights/${id}`)
  return data
}

export async function getFlightFares(id: number): Promise<FareResponse[]> {
  const { data } = await api.get<FareResponse[]>(`/flights/${id}/fares`)
  return data
}

export async function searchAirports(query: string): Promise<AirportResponse[]> {
  const { data } = await api.get<AirportResponse[]>('/airports/search', { params: { query } })
  return data
}

export async function getAirport(code: string): Promise<AirportResponse> {
  const { data } = await api.get<AirportResponse>(`/airports/${code}`)
  return data
}

export async function listAirlines(params?: PaginationParams): Promise<AirlineResponse[]> {
  const { data } = await api.get<AirlineResponse[]>('/airlines', { params })
  return data
}

export async function listFlights(params?: PaginationParams): Promise<FlightResponse[]> {
  const { data } = await api.get<FlightResponse[]>('/flights', { params })
  return data
}

export async function listAircrafts(params?: PaginationParams): Promise<AircraftResponse[]> {
  const { data } = await api.get<AircraftResponse[]>('/aircrafts', { params })
  return data
}

export async function getAirline(id: number): Promise<AirlineResponse> {
  const { data } = await api.get<AirlineResponse>(`/airlines/${id}`)
  return data
}

export async function getAircraftSeats(aircraftId: number): Promise<SeatResponse[]> {
  const { data } = await api.get<SeatResponse[]>(`/aircrafts/${aircraftId}/seats`)
  return data
}

export async function getReservedSeatNumbers(flightId: number): Promise<string[]> {
  const { data } = await api.get<string[]>(`/flights/${flightId}/seats/reserved/public`)
  return data
}
