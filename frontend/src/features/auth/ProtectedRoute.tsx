import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'

interface Props {
  children: ReactNode
  requireRole?: 'admin'
}

export function ProtectedRoute({ children, requireRole }: Props) {
  const { token, role } = useAuthStore()

  if (token === null) {
    return <Navigate to="/login" replace />
  }

  if (requireRole === 'admin' && role !== 'admin') {
    return <Navigate to="/" replace />
  }

  return <>{children}</>
}
