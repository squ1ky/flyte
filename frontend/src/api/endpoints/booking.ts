import { api } from '@/api/instance'
import type {
  CreateBookingRequest,
  CreateBookingResponse,
  BookingResponse,
} from '@/types/booking'

export interface ListBookingsParams {
  limit?: number
  offset?: number
}

export async function createBooking(data: CreateBookingRequest): Promise<CreateBookingResponse> {
  const { data: result } = await api.post<CreateBookingResponse>('/bookings', data)
  return result
}

export async function listBookings(params?: ListBookingsParams): Promise<BookingResponse[]> {
  const { data } = await api.get<BookingResponse[]>('/bookings', { params })
  return data
}

export async function getBooking(id: string): Promise<BookingResponse> {
  const { data } = await api.get<BookingResponse>(`/bookings/${id}`)
  return data
}

export async function getBookingByRef(ref: string): Promise<BookingResponse> {
  const { data } = await api.get<BookingResponse>(`/bookings/ref/${ref}`)
  return data
}

export async function cancelBooking(id: string): Promise<void> {
  await api.post(`/bookings/${id}/cancel`)
}
