import { Outlet } from 'react-router-dom'
import { Header } from '@/components/layout/Header'
import { Toaster } from '@/components/ui/sonner'

export function RootLayout() {
  return (
    <>
      <Header />
      <main>
        <Outlet />
      </main>
      <Toaster />
    </>
  )
}
