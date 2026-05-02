import { Outlet } from 'react-router-dom'
import { Header } from '@/components/layout/Header'
import { Toaster } from '@/components/ui/sonner'
import { ErrorBoundary } from '@/components/ErrorBoundary'

export function RootLayout() {
  return (
    <ErrorBoundary>
      <Header />
      <main>
        <Outlet />
      </main>
      <Toaster />
    </ErrorBoundary>
  )
}
