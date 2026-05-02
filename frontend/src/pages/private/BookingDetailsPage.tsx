import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Plane } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { getBooking, cancelBooking } from '@/api/endpoints/booking'
import { getFlightDetails } from '@/api/endpoints/flight'
import { ExpiryCountdown } from '@/features/booking/ExpiryCountdown'
import { formatPrice, formatDateTime, formatDuration } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { BookingStatus } from '@/types/booking'

const STATUS_LABELS: Record<BookingStatus, string> = {
  pending: 'Ожидает оплаты',
  paid: 'Оплачено',
  cancelled: 'Отменено',
}

const SEAT_CLASS_LABELS: Record<string, string> = {
  economy: 'Эконом',
  business: 'Бизнес',
  first: 'Первый',
}

function StatusBadge({ status }: { status: BookingStatus }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        status === 'pending' && 'bg-yellow-100 text-yellow-800',
        status === 'paid' && 'bg-green-100 text-green-800',
        status === 'cancelled' && 'bg-gray-100 text-gray-600',
      )}
    >
      {STATUS_LABELS[status]}
    </span>
  )
}

export function BookingDetailsPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: booking, isLoading, error } = useQuery({
    queryKey: ['booking', id],
    queryFn: () => getBooking(id!),
    enabled: !!id,
  })

  const { data: flight } = useQuery({
    queryKey: ['flight', booking?.flight_id],
    queryFn: () => getFlightDetails(booking!.flight_id),
    enabled: !!booking,
  })

  const { mutate: cancel, isPending: cancelling } = useMutation({
    mutationFn: () => cancelBooking(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['booking', id] })
      queryClient.invalidateQueries({ queryKey: ['bookings'] })
      toast.success('Бронирование отменено')
    },
    onError: (err: Error) => {
      toast.error(err.message)
    },
  })

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8 text-center text-muted-foreground">
        Загрузка...
      </div>
    )
  }

  if (error || !booking) {
    return (
      <div className="container mx-auto px-4 py-8 text-center text-muted-foreground">
        Бронирование не найдено.
      </div>
    )
  }

  return (
    <div className="container mx-auto max-w-2xl px-4 py-8 flex flex-col gap-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm text-muted-foreground mb-1">Код бронирования</p>
          <h1 className="text-3xl font-bold tracking-widest">{booking.ref_code}</h1>
        </div>
        <StatusBadge status={booking.status} />
      </div>

      {booking.status === 'pending' && booking.expires_at && (
        <ExpiryCountdown
          expiresAt={booking.expires_at}
          onExpire={() => queryClient.invalidateQueries({ queryKey: ['booking', id] })}
        />
      )}

      {flight && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Рейс</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap items-center gap-4">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border bg-muted text-xs font-bold text-muted-foreground">
              {flight.airline.iata_code}
            </div>
            <div className="min-w-[90px]">
              <p className="font-semibold leading-tight">
                {formatDateTime(flight.departure_time, flight.departure_airport.timezone)}
              </p>
              <p className="text-sm text-muted-foreground">
                {flight.departure_airport.city} ({flight.departure_airport.code})
              </p>
            </div>
            <div className="flex flex-1 flex-col items-center gap-1">
              <p className="text-xs text-muted-foreground">
                {formatDuration(flight.departure_time, flight.arrival_time)}
              </p>
              <div className="flex w-full items-center gap-1">
                <div className="h-px flex-1 bg-border" />
                <Plane className="h-3 w-3 rotate-90 text-muted-foreground" />
                <div className="h-px flex-1 bg-border" />
              </div>
              <p className="text-xs text-muted-foreground">{flight.flight_number}</p>
            </div>
            <div className="min-w-[90px] text-right">
              <p className="font-semibold leading-tight">
                {formatDateTime(flight.arrival_time, flight.arrival_airport.timezone)}
              </p>
              <p className="text-sm text-muted-foreground">
                {flight.arrival_airport.city} ({flight.arrival_airport.code})
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Билеты</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {booking.tickets.map((ticket) => (
            <div
              key={ticket.id}
              className="flex items-center justify-between gap-2 rounded-lg border p-3"
            >
              <div>
                <p className="font-medium">
                  {ticket.passenger_last_name} {ticket.passenger_first_name}
                </p>
                <p className="text-sm text-muted-foreground">
                  Место {ticket.seat_number} · {SEAT_CLASS_LABELS[ticket.seat_class] ?? ticket.seat_class}
                </p>
                <p className="text-xs text-muted-foreground">{ticket.passenger_doc_num}</p>
              </div>
              <p className="font-semibold shrink-0">
                {formatPrice(ticket.price_cents, booking.currency)}
              </p>
            </div>
          ))}
          <div className="flex items-center justify-between border-t pt-3">
            <span className="font-semibold">Итого</span>
            <span className="text-lg font-bold">
              {formatPrice(booking.total_price_cents, booking.currency)}
            </span>
          </div>
        </CardContent>
      </Card>

      {booking.status === 'pending' && (
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant="destructive" disabled={cancelling}>
              Отменить бронь
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Отменить бронирование?</AlertDialogTitle>
              <AlertDialogDescription>
                Бронирование {booking.ref_code} будет отменено. Это действие необратимо.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Назад</AlertDialogCancel>
              <AlertDialogAction onClick={() => cancel()}>Отменить</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      <Button variant="ghost" className="self-start" onClick={() => navigate('/profile')}>
        ← В профиль
      </Button>
    </div>
  )
}
