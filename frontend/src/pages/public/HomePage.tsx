import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { SearchForm } from '@/features/flight-search/SearchForm'
import { FlightCard } from '@/features/flight-search/FlightCard'
import { searchFlights, type SearchFlightsParams } from '@/api/endpoints/flight'
import type { SeatClass } from '@/types/flight'

function FlightSkeleton() {
  return <div className="h-24 animate-pulse rounded-xl border bg-card" />
}

export function HomePage() {
  const [searchParams] = useSearchParams()

  const from = searchParams.get('from') ?? ''
  const to = searchParams.get('to') ?? ''
  const date = searchParams.get('date') ?? ''
  const passengers = Number(searchParams.get('passengers') ?? 1)
  const seat_class = (searchParams.get('seat_class') as SeatClass | null) ?? undefined

  const params: SearchFlightsParams = { from, to, date, passengers, seat_class }

  const isReady = !!from && !!to && !!date

  const { data: flights, isLoading, error } = useQuery({
    queryKey: ['flights', params],
    queryFn: () => searchFlights(params),
    enabled: isReady,
  })

  return (
    <div className="container mx-auto max-w-4xl px-4 py-8">
      <h1 className="mb-6 text-3xl font-bold">Поиск авиабилетов</h1>
      <SearchForm />

      <div className="mt-8">
        {!isReady && (
          <p className="text-center text-muted-foreground">Введите параметры поиска</p>
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
            {flights.map(flight => (
              <FlightCard key={flight.id} flight={flight} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
