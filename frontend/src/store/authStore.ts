import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Role } from '@/types/user'

interface AuthState {
  token: string | null
  userId: number | null
  role: Role | null
  email: string | null
  setAuth: (token: string, userId: number, role: Role, email?: string | null) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userId: null,
      role: null,
      email: null,
      setAuth: (token, userId, role, email = null) => set({ token, userId, role, email }),
      logout: () => set({ token: null, userId: null, role: null, email: null }),
    }),
    { name: 'auth' }
  )
)
