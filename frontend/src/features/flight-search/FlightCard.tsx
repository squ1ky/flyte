import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowRight, Plane } from 'lucide-react'
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
      className="group flex cursor-pointer items-center gap-4 rounded-xl border bg-card p-4 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/30 hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      {/* Logo + IATA */}
      <div className="flex shrink-0 flex-col items-center gap-1.5">
        <div className="flex h-12 w-12 items-center justify-center overflow-hidden rounded-xl border bg-gradient-to-br from-muted to-secondary shadow-sm">
          {flight.airline.logo_url && !logoError ? (
            <img
              src={flight.airline.logo_url}
              alt={flight.airline.name}
              className="h-full w-full object-contain"
              onError={() => setLogoError(true)}
            />
          ) : (
            <span className="text-base font-bold text-muted-foreground">
              {flight.airline.name[0]}
            </span>
          )}
        </div>
        <p className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
          {flight.airline.iata_code}
        </p>
      </div>

      {/* Departure */}
      <div className="min-w-[90px]">
        <p className="text-xl font-bold tabular-nums leading-tight">
          {formatDateTime(flight.departure_time, flight.departure_airport.timezone)}
        </p>
        <p className="text-sm font-medium">{flight.departure_airport.city}</p>
        <p className="text-xs text-muted-foreground">{flight.departure_airport.code}</p>
      </div>

      {/* Flight path */}
      <div className="flex flex-1 flex-col items-center gap-1.5">
        <p className="text-xs font-medium text-muted-foreground">
          {formatDuration(flight.departure_time, flight.arrival_time)}
        </p>
        <div className="flex w-full items-center gap-1.5">
          <div className="h-px flex-1 bg-gradient-to-r from-transparent via-primary/30 to-primary/60" />
          <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 ring-1 ring-primary/20">
            <Plane className="h-3.5 w-3.5 rotate-90 text-primary" />
          </div>
          <div className="h-px flex-1 bg-gradient-to-r from-primary/60 via-primary/30 to-transparent" />
        </div>
        <p className="text-[10px] font-medium text-muted-foreground">{flight.flight_number}</p>
      </div>

      {/* Arrival */}
      <div className="min-w-[90px] text-right">
        <p className="text-xl font-bold tabular-nums leading-tight">
          {formatDateTime(flight.arrival_time, flight.arrival_airport.timezone)}
        </p>
        <p className="text-sm font-medium">{flight.arrival_airport.city}</p>
        <p className="text-xs text-muted-foreground">{flight.arrival_airport.code}</p>
      </div>

      {/* Price + hint */}
      <div className="ml-auto flex shrink-0 flex-col items-end gap-1">
        <p className="bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-2xl font-extrabold text-transparent">
          {formatPrice(flight.min_price_cents, flight.currency)}
        </p>
        <span className="flex items-center gap-0.5 text-xs font-medium text-primary opacity-0 transition-opacity duration-200 group-hover:opacity-100">
          Выбрать <ArrowRight className="h-3 w-3" />
        </span>
      </div>
    </div>
  )
}
