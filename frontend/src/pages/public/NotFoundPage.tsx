import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'

export function NotFoundPage() {
  return (
    <div className="flex min-h-[calc(100vh-3.5rem)] flex-col items-center justify-center gap-6 px-4 text-center">
      <p className="text-8xl font-bold text-muted-foreground/20">404</p>
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">Страница не найдена</h1>
        <p className="text-muted-foreground">
          Страница, которую вы ищете, не существует или была перемещена.
        </p>
      </div>
      <Button asChild>
        <Link to="/">На главную</Link>
      </Button>
    </div>
  )
}
