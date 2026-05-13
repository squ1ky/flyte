import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm, useFieldArray, FormProvider } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Plane } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { SeatMap } from '@/features/booking/SeatMap'
import { PassengerInputCard } from '@/features/booking/PassengerInputCard'
import { getFlightDetails, getFlightFares, getAircraftSeats, getReservedSeatNumbers } from '@/api/endpoints/flight'
import { getPassengers } from '@/api/endpoints/user'
import { createBooking } from '@/api/endpoints/booking'
import { useAuthStore } from '@/store/authStore'
import { formatDateTime, formatDuration, formatPrice } from '@/lib/format'
import type { FlightResponse } from '@/types/flight'

const passengerSchema = z.object({
  first_name: z.string().min(1, 'Обязательное поле'),
  last_name: z.string().min(1, 'Обязательное поле'),
  document_number: z.string().min(1, 'Обязательное поле'),
})

const schema = z.object({
  contact_email: z.string().email('Некорректный email'),
  passengers: z.array(passengerSchema).min(1, 'Выберите хотя бы одно место'),
})

type FormValues = z.infer<typeof schema>

function FlightSummary({ flight }: { flight: FlightResponse }) {
  return (
    <div className="flex flex-wrap items-center gap-4 rounded-xl border bg-card p-4 shadow-sm">
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border bg-muted text-sm font-bold text-muted-foreground">
        {flight.airline.iata_code}
      </div>

      <div className="min-w-[80px]">
        <p className="font-semibold leading-tight">
          {formatDateTime(flight.departure_time, flight.departure_airport.timezone)}
        </p>
        <p className="text-sm text-muted-foreground">{flight.departure_airport.city}</p>
      </div>

      <div className="flex flex-1 flex-col items-center gap-1">
        <p className="text-sm text-muted-foreground">
          {formatDuration(flight.departure_time, flight.arrival_time)}
        </p>
        <div className="flex w-full items-center gap-1">
          <div className="h-px flex-1 bg-border" />
          <Plane className="h-4 w-4 rotate-90 text-muted-foreground" />
          <div className="h-px flex-1 bg-border" />
        </div>
        <p className="text-xs text-muted-foreground">{flight.flight_number}</p>
      </div>

      <div className="min-w-[80px] text-right">
        <p className="font-semibold leading-tight">
          {formatDateTime(flight.arrival_time, flight.arrival_airport.timezone)}
        </p>
        <p className="text-sm text-muted-foreground">{flight.arrival_airport.city}</p>
      </div>

      <div className="ml-auto shrink-0 text-right">
        <p className="text-lg font-bold text-primary">
          от {formatPrice(flight.min_price_cents, flight.currency)}
        </p>
        <p className="text-xs text-muted-foreground">{flight.airline.name}</p>
      </div>
    </div>
  )
}

export function BookingPage() {
  const { flightId } = useParams<{ flightId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { userId, email } = useAuthStore()

  const [selectedSeats, setSelectedSeats] = useState<string[]>([])

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    mode: 'onChange',
    defaultValues: {
      contact_email: email ?? '',
      passengers: [],
    },
  })

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: 'passengers',
  })

  const flightIdNum = Number(flightId)

  const { data: flight, isLoading: flightLoading } = useQuery({
    queryKey: ['flight', flightIdNum],
    queryFn: () => getFlightDetails(flightIdNum),
    enabled: !!flightIdNum,
  })

  const { data: fares = [] } = useQuery({
    queryKey: ['flight', flightIdNum, 'fares'],
    queryFn: () => getFlightFares(flightIdNum),
    enabled: !!flightIdNum,
  })

  const { data: reservedSeats = [] } = useQuery({
    queryKey: ['flight', flightIdNum, 'reserved-seats'],
    queryFn: () => getReservedSeatNumbers(flightIdNum),
    enabled: !!flightIdNum,
  })

  const { data: seats = [] } = useQuery({
    queryKey: ['aircraft', flight?.aircraft.id ?? 0, 'seats'],
    queryFn: () => getAircraftSeats(flight!.aircraft.id),
    enabled: !!flight,
  })

  const { data: savedPassengers = [] } = useQuery({
    queryKey: ['user', userId, 'passengers'],
    queryFn: () => getPassengers(userId!),
    enabled: !!userId,
  })

  const { mutate, isPending } = useMutation({
    mutationFn: createBooking,
    onSuccess: (data) => {
      navigate(`/bookings/${data.booking_id}`)
    },
    onError: (error) => {
      const err = error as Error & { status?: number }
      if (err.status === 409) {
        toast.error('Место уже занято. Пожалуйста, выберите другое.')
        queryClient.invalidateQueries({ queryKey: ['flight', flightIdNum, 'reserved-seats'] })
      } else {
        toast.error(err.message)
      }
    },
  })

  function handleSeatsChange(newSeats: string[]) {
    if (newSeats.length > selectedSeats.length) {
      append({ first_name: '', last_name: '', document_number: '' })
    } else if (newSeats.length < selectedSeats.length) {
      const removedIdx = selectedSeats.findIndex(s => !newSeats.includes(s))
      if (removedIdx !== -1) remove(removedIdx)
    }
    setSelectedSeats(newSeats)
  }

  function onSubmit(values: FormValues) {
    mutate({
      flight_id: flightIdNum,
      contact_email: values.contact_email,
      passengers: values.passengers.map((p, i) => ({
        ...p,
        seat_number: selectedSeats[i],
        seat_class: seats.find(s => s.seat_number === selectedSeats[i])?.seat_class ?? 'economy',
      })),
    })
  }

  if (flightLoading) {
    return (
      <div className="container mx-auto px-4 py-8 text-center text-muted-foreground">
        Загрузка рейса...
      </div>
    )
  }

  if (!flight) {
    return (
      <div className="container mx-auto px-4 py-8 text-center text-muted-foreground">
        Рейс не найден.
      </div>
    )
  }

  const isSubmitDisabled = isPending || selectedSeats.length === 0 || !form.formState.isValid

  return (
    <div className="container mx-auto max-w-3xl space-y-6 px-4 py-8">
      <h1 className="text-2xl font-bold">Бронирование билета</h1>

      <FlightSummary flight={flight} />

      <section className="space-y-3">
        <h2 className="text-lg font-semibold">
          Выбор мест{' '}
          {selectedSeats.length > 0 && (
            <span className="text-sm font-normal text-muted-foreground">
              (выбрано {selectedSeats.length})
            </span>
          )}
        </h2>
        <SeatMap
          aircraftId={flight.aircraft.id}
          fares={fares}
          reservedSeats={reservedSeats}
          selected={selectedSeats}
          onChange={handleSeatsChange}
        />
      </section>

      <FormProvider {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          {fields.map((field, i) => (
            <PassengerInputCard
              key={field.id}
              index={i}
              seatNumber={selectedSeats[i]}
              seatClass={seats.find(s => s.seat_number === selectedSeats[i])?.seat_class ?? 'economy'}
              savedPassengers={savedPassengers}
            />
          ))}

          {selectedSeats.length > 0 && (
            <div className="space-y-1">
              <Label htmlFor="contact_email">Контактный email</Label>
              <Input
                id="contact_email"
                type="email"
                {...form.register('contact_email')}
              />
              {form.formState.errors.contact_email && (
                <p className="text-xs text-destructive">
                  {form.formState.errors.contact_email.message}
                </p>
              )}
            </div>
          )}

          <Button
            type="submit"
            disabled={isSubmitDisabled}
            className="w-full"
            size="lg"
          >
            {isPending ? 'Бронирование...' : 'Забронировать'}
          </Button>
        </form>
      </FormProvider>
    </div>
  )
}
