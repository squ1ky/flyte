import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Role } from '@/types/user'

interface AuthState {
  token: string | null
  userId: number | null
  role: Role | null
  setAuth: (token: string, userId: number, role: Role) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userId: null,
      role: null,
      setAuth: (token, userId, role) => set({ token, userId, role }),
      logout: () => set({ token: null, userId: null, role: null }),
    }),
    { name: 'auth' }
  )
)
