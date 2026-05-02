import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle } from '@/components/ui/card'
import { listBookings } from '@/api/endpoints/booking'
import { getFlightDetails } from '@/api/endpoints/flight'
import { formatPrice, formatDateTime } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { BookingResponse, BookingStatus } from '@/types/booking'

const LIMIT = 10

const STATUS_LABELS: Record<BookingStatus, string> = {
  pending: 'Ожидает оплаты',
  paid: 'Оплачено',
  cancelled: 'Отменено',
}

function StatusBadge({ status }: { status: BookingStatus }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
        status === 'pending' && 'bg-yellow-100 text-yellow-800',
        status === 'paid' && 'bg-green-100 text-green-800',
        status === 'cancelled' && 'bg-gray-100 text-gray-600',
      )}
    >
      {STATUS_LABELS[status]}
    </span>
  )
}

function BookingRow({ booking }: { booking: BookingResponse }) {
  const navigate = useNavigate()

  const { data: flight } = useQuery({
    queryKey: ['flight', booking.flight_id],
    queryFn: () => getFlightDetails(booking.flight_id),
  })

  return (
    <div
      role="button"
      tabIndex={0}
      className="flex items-center justify-between gap-4 rounded-lg border bg-card p-4 shadow-sm cursor-pointer hover:bg-accent transition-colors"
      onClick={() => navigate(`/bookings/${booking.id}`)}
      onKeyDown={(e) => e.key === 'Enter' && navigate(`/bookings/${booking.id}`)}
    >
      <div className="flex flex-col gap-1 min-w-0">
        <p className="font-bold tracking-widest text-sm">{booking.ref_code}</p>
        {flight ? (
          <p className="text-sm text-muted-foreground truncate">
            {flight.departure_airport.city} → {flight.arrival_airport.city}
            {' · '}
            {formatDateTime(flight.departure_time, flight.departure_airport.timezone)}
          </p>
        ) : (
          <p className="text-sm text-muted-foreground">Загрузка...</p>
        )}
        <StatusBadge status={booking.status} />
      </div>
      <p className="font-semibold shrink-0">{formatPrice(booking.total_price_cents, booking.currency)}</p>
    </div>
  )
}

export function BookingsList() {
  const [offset, setOffset] = useState(0)

  const { data: bookings = [], isLoading } = useQuery({
    queryKey: ['bookings', { limit: LIMIT, offset }],
    queryFn: () => listBookings({ limit: LIMIT, offset }),
  })

  if (isLoading) {
    return <p className="text-sm text-muted-foreground">Загрузка...</p>
  }

  if (bookings.length === 0 && offset === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-base font-normal text-muted-foreground text-center py-4">
            У вас пока нет бронирований
          </CardTitle>
        </CardHeader>
      </Card>
    )
  }

  const hasPrev = offset > 0
  const hasNext = bookings.length === LIMIT

  return (
    <div className="flex flex-col gap-3">
      {bookings.map((b) => (
        <BookingRow key={b.id} booking={b} />
      ))}

      {(hasPrev || hasNext) && (
        <div className="flex justify-between pt-1">
          <Button
            variant="outline"
            size="sm"
            disabled={!hasPrev}
            onClick={() => setOffset((o) => Math.max(0, o - LIMIT))}
          >
            Назад
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!hasNext}
            onClick={() => setOffset((o) => o + LIMIT)}
          >
            Вперёд
          </Button>
        </div>
      )}
    </div>
  )
}
