import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Plane } from 'lucide-react'
import { SearchForm } from '@/features/flight-search/SearchForm'
import { FlightCard } from '@/features/flight-search/FlightCard'
import { searchFlights, type SearchFlightsParams } from '@/api/endpoints/flight'
import type { SeatClass } from '@/types/flight'

function FlightSkeleton() {
  return <div className="h-24 animate-pulse rounded-xl border bg-card shadow-sm" />
}

export function HomePage() {
  const [searchParams] = useSearchParams()

  const from = searchParams.get('from') ?? ''
  const to = searchParams.get('to') ?? ''
  const date = searchParams.get('date') ?? ''
  const seat_class = (searchParams.get('seat_class') as SeatClass | null) ?? undefined

  const params: SearchFlightsParams = { from, to, date, seat_class }

  const isReady = !!from && !!to && !!date

  const { data: flights, isLoading, error } = useQuery({
    queryKey: ['flights', params],
    queryFn: () => searchFlights(params),
    enabled: isReady,
  })

  return (
    <div>
      {/* Hero */}
      <div className="relative overflow-hidden bg-gradient-to-br from-blue-600 via-blue-700 to-indigo-900 px-4 pb-16 pt-12">
        <div className="bg-grid absolute inset-0" />
        <div className="absolute -right-24 -top-24 h-96 w-96 rounded-full bg-white/5 blur-3xl" />
        <div className="absolute -bottom-16 left-8 h-64 w-64 rounded-full bg-indigo-400/10 blur-2xl" />

        <div className="container relative mx-auto max-w-4xl">
          <div className="animate-fade-in-up mb-3 flex items-center gap-2">
            <div className="animate-float">
              <Plane className="h-5 w-5 rotate-45 text-blue-300" />
            </div>
            <p className="text-sm font-medium uppercase tracking-widest text-blue-300">
              Онлайн‑сервис авиабилетов
            </p>
          </div>

          <h1
            className="animate-fade-in-up mb-2 text-4xl font-extrabold tracking-tight text-white sm:text-5xl"
            style={{ animationDelay: '80ms' }}
          >
            Найдите свой рейс
          </h1>
          <p
            className="animate-fade-in-up mb-10 text-lg text-blue-200"
            style={{ animationDelay: '160ms' }}
          >
            Удобный поиск, мгновенное бронирование, лучшие цены
          </p>

          <div className="animate-fade-in-up" style={{ animationDelay: '240ms' }}>
            <SearchForm />
          </div>
        </div>
      </div>

      {/* Results */}
      <div className="container mx-auto max-w-4xl px-4 py-8">
        {!isReady && (
          <p className="text-center text-muted-foreground">Введите параметры поиска выше</p>
        )}

        {isReady && isLoading && (
          <div className="space-y-3">
            {Array.from({ length: 4 }).map((_, i) => (
              <FlightSkeleton key={i} />
            ))}
          </div>
        )}

        {isReady && error && (
          <p className="text-center text-destructive">
            {error instanceof Error ? error.message : 'Ошибка при поиске рейсов'}
          </p>
        )}

        {isReady && !isLoading && !error && flights?.length === 0 && (
          <p className="text-center text-muted-foreground">По вашему запросу рейсы не найдены</p>
        )}

        {flights && flights.length > 0 && (
          <div className="space-y-3">
            <p className="text-sm font-medium text-muted-foreground">
              Найдено рейсов: {flights.length}
            </p>
            {flights.map((flight, i) => (
              <div
                key={flight.id}
                className="animate-fade-in-up"
                style={{ animationDelay: `${i * 60}ms` }}
              >
                <FlightCard flight={flight} />
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
