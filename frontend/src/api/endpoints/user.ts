import { api } from '@/api/instance'
import type {
  UserResponse,
  PassengerInput,
  PassengerResponse,
  CreatePassengerResponse,
} from '@/types/user'

export async function getUser(id: number): Promise<UserResponse> {
  const { data } = await api.get<UserResponse>(`/users/${id}`)
  return data
}

export async function addPassenger(
  userId: number,
  data: PassengerInput
): Promise<CreatePassengerResponse> {
  const { data: result } = await api.post<CreatePassengerResponse>(
    `/users/${userId}/passengers`,
    data
  )
  return result
}

export async function getPassengers(userId: number): Promise<PassengerResponse[]> {
  const { data } = await api.get<PassengerResponse[]>(`/users/${userId}/passengers`)
  return data
}

export async function deletePassenger(userId: number, passengerId: number): Promise<void> {
  await api.delete(`/users/${userId}/passengers/${passengerId}`)
}
