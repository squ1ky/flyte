import { api } from '@/api/instance'
import type {
  CreateFlightRequest,
  CreateAirlineRequest,
  CreateAircraftRequest,
  CreateSeatRequest,
  FlightStatus,
} from '@/types/flight'

export interface CreateFlightPayload extends CreateFlightRequest {
  fare_rules: { seat_class: string; price_cents: number }[]
}

export async function adminCreateFlight(payload: CreateFlightPayload): Promise<{ id: number }> {
  const { data } = await api.post<{ id: number }>('/flights', payload)
  return data
}

export async function adminUpdateFlightStatus(
  flightId: number,
  status: FlightStatus,
): Promise<void> {
  await api.patch(`/flights/${flightId}/status`, { status })
}

export async function adminCreateAirline(payload: CreateAirlineRequest): Promise<{ id: number }> {
  const { data } = await api.post<{ id: number }>('/airlines', payload)
  return data
}

export async function adminCreateAircraft(
  payload: CreateAircraftRequest,
): Promise<{ id: number }> {
  const { data } = await api.post<{ id: number }>('/aircrafts', payload)
  return data
}

export async function adminAddAircraftSeats(
  aircraftId: number,
  seats: CreateSeatRequest[],
): Promise<{ seats_added: number }> {
  const { data } = await api.post<{ seats_added: number }>(
    `/aircrafts/${aircraftId}/seats`,
    { seats },
  )
  return data
}
