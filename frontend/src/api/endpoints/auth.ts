import { api } from '@/api/instance'
import type { SignUpRequest, SignUpResponse, SignInRequest, SignInResponse } from '@/types/user'

export async function signUp(data: SignUpRequest): Promise<SignUpResponse> {
  const { data: result } = await api.post<SignUpResponse>('/auth/sign-up', data)
  return result
}

export async function signIn(data: SignInRequest): Promise<SignInResponse> {
  const { data: result } = await api.post<SignInResponse>('/auth/sign-in', data)
  return result
}
