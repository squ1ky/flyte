import { Link, useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { Plane } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/store/authStore'
import { Badge } from '@/components/ui/badge'

export function Header() {
  const { token, email, role } = useAuthStore()
  const logout = useAuthStore((s) => s.logout)
  const qc = useQueryClient()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    qc.clear()
    navigate('/')
  }

  return (
    <header className="sticky top-0 z-50 border-b border-white/40 bg-white/80 shadow-sm backdrop-blur-xl">
      <div className="container mx-auto flex h-14 items-center justify-between px-4">
        <Link to="/" className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 shadow-sm">
            <Plane className="h-4 w-4 rotate-45 text-white" />
          </div>
          <span className="bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-lg font-extrabold tracking-tight text-transparent">
            Flyte
          </span>
        </Link>

        <nav className="flex items-center gap-1">
          <Link
            to="/"
            className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-blue-50 hover:text-foreground"
          >
            Поиск
          </Link>
          {token && (
            <Link
              to="/profile"
              className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-blue-50 hover:text-foreground"
            >
              Профиль
            </Link>
          )}
          {role === 'admin' && (
            <Link
              to="/admin"
              className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-blue-50 hover:text-foreground"
            >
              Админка
              <Badge variant="secondary" className="h-4 px-1 text-[10px]">
                admin
              </Badge>
            </Link>
          )}
        </nav>

        <div className="flex items-center gap-2">
          {token ? (
            <>
              <span className="hidden text-sm text-muted-foreground sm:inline">{email}</span>
              <Button variant="outline" size="sm" onClick={handleLogout}>
                Выйти
              </Button>
            </>
          ) : (
            <>
              <Button variant="ghost" size="sm" asChild>
                <Link to="/login">Войти</Link>
              </Button>
              <Button
                size="sm"
                className="bg-gradient-to-r from-blue-600 to-indigo-600 shadow-sm hover:from-blue-700 hover:to-indigo-700"
                asChild
              >
                <Link to="/register">Регистрация</Link>
              </Button>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
