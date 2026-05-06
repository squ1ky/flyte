import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plane } from 'lucide-react'
import type { FlightResponse } from '@/types/flight'
import { formatDateTime, formatDuration, formatPrice } from '@/lib/format'

interface Props {
  flight: FlightResponse
}

export function FlightCard({ flight }: Props) {
  const navigate = useNavigate()
  const [logoError, setLogoError] = useState(false)

  function handleClick() {
    navigate(`/booking/${flight.id}`)
  }

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={handleClick}
      onKeyDown={e => {
        if (e.key === 'Enter' || e.key === ' ') handleClick()
      }}
      className="flex cursor-pointer items-center gap-4 rounded-xl border bg-card p-4 shadow-sm transition-shadow hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      {/* Airline logo */}
      <div className="flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-lg border bg-muted">
        {flight.airline.logo_url && !logoError ? (
          <img
            src={flight.airline.logo_url}
            alt={flight.airline.name}
            className="h-full w-full object-contain"
            onError={() => setLogoError(true)}
          />
        ) : (
          <span className="text-lg font-bold text-muted-foreground">
            {flight.airline.name[0]}
          </span>
        )}
      </div>

      {/* Departure */}
      <div className="min-w-[90px]">
        <p className="text-lg font-semibold leading-tight">
          {formatDateTime(flight.departure_time, flight.departure_airport.timezone)}
        </p>
        <p className="text-sm text-muted-foreground">{flight.departure_airport.city}</p>
        <p className="text-xs text-muted-foreground">{flight.departure_airport.code}</p>
      </div>

      {/* Duration + line */}
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

      {/* Arrival */}
      <div className="min-w-[90px] text-right">
        <p className="text-lg font-semibold leading-tight">
          {formatDateTime(flight.arrival_time, flight.arrival_airport.timezone)}
        </p>
        <p className="text-sm text-muted-foreground">{flight.arrival_airport.city}</p>
        <p className="text-xs text-muted-foreground">{flight.arrival_airport.code}</p>
      </div>

      {/* Price */}
      <div className="ml-auto shrink-0 text-right">
        <p className="text-xl font-bold text-primary">
          {formatPrice(flight.min_price_cents, flight.currency)}
        </p>
      </div>
    </div>
  )
}
