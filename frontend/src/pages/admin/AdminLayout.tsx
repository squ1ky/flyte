import { NavLink, Outlet } from 'react-router-dom'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/admin', label: 'Рейсы', end: true },
  { to: '/admin/airlines', label: 'Авиакомпании' },
  { to: '/admin/aircrafts', label: 'Самолёты' },
]

export function AdminLayout() {
  return (
    <div className="flex min-h-[calc(100vh-3.5rem)]">
      <aside className="w-48 shrink-0 border-r bg-muted/40 p-4">
        <p className="mb-4 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Панель admin
        </p>
        <nav className="flex flex-col gap-1">
          {navItems.map(({ to, label, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                cn(
                  'rounded-md px-3 py-2 text-sm transition-colors hover:bg-accent',
                  isActive && 'bg-accent font-medium',
                )
              }
            >
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <main className="flex-1 p-6">
        <Outlet />
      </main>
    </div>
  )
}
