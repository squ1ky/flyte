import { Link, useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/store/authStore'

export function Header() {
  const { token, userId } = useAuthStore()
  const logout = useAuthStore((s) => s.logout)
  const qc = useQueryClient()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    qc.clear()
    navigate('/')
  }

  return (
    <header className="sticky top-0 z-50 border-b bg-background">
      <div className="container mx-auto flex h-14 items-center justify-between px-4">
        <Link to="/" className="text-lg font-bold">
          Flyte
        </Link>

        <nav className="flex items-center gap-6">
          <Link to="/" className="text-sm hover:text-foreground/80">
            Поиск
          </Link>
          {token && (
            <Link to="/profile" className="text-sm hover:text-foreground/80">
              Профиль
            </Link>
          )}
        </nav>

        <div className="flex items-center gap-2">
          {token ? (
            <>
              <span className="text-sm text-muted-foreground">#{userId}</span>
              <Button variant="outline" size="sm" onClick={handleLogout}>
                Выйти
              </Button>
            </>
          ) : (
            <>
              <Button variant="ghost" size="sm" asChild>
                <Link to="/login">Войти</Link>
              </Button>
              <Button size="sm" asChild>
                <Link to="/register">Регистрация</Link>
              </Button>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
