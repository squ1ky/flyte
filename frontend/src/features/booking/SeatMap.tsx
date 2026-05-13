import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { getAircraftSeats } from '@/api/endpoints/flight'
import { formatPrice } from '@/lib/format'
import type { FareResponse, SeatClass, SeatResponse } from '@/types/flight'

interface Props {
  aircraftId: number
  fares: FareResponse[]
  reservedSeats?: string[]
  selected: string[]
  onChange: (seats: string[]) => void
}

const CLASS_COLORS: Record<SeatClass, string> = {
  economy: 'bg-blue-50 hover:bg-blue-100 border-blue-300 text-blue-800',
  business: 'bg-amber-50 hover:bg-amber-100 border-amber-300 text-amber-800',
  first: 'bg-violet-50 hover:bg-violet-100 border-violet-300 text-violet-800',
}

const CLASS_SELECTED: Record<SeatClass, string> = {
  economy: 'bg-blue-500 border-blue-600 text-white hover:bg-blue-600',
  business: 'bg-amber-500 border-amber-600 text-white hover:bg-amber-600',
  first: 'bg-violet-500 border-violet-600 text-white hover:bg-violet-600',
}

const CLASS_LABELS: Record<SeatClass, string> = {
  economy: 'Эконом',
  business: 'Бизнес',
  first: 'Первый',
}

export function SeatMap({ aircraftId, fares, reservedSeats = [], selected, onChange }: Props) {
  const { data: seats = [], isLoading } = useQuery({
    queryKey: ['aircraft', aircraftId, 'seats'],
    queryFn: () => getAircraftSeats(aircraftId),
  })

  if (isLoading) {
    return <div className="py-8 text-center text-sm text-muted-foreground">Загрузка схемы самолёта...</div>
  }

  const allCols = [...new Set(seats.map(s => s.col))].sort((a, b) => a - b)

  const rows = seats.reduce<Record<number, SeatResponse[]>>((acc, seat) => {
    if (!acc[seat.row]) acc[seat.row] = []
    acc[seat.row].push(seat)
    return acc
  }, {})

  const sortedRowNums = Object.keys(rows).map(Number).sort((a, b) => a - b)

  function handleClick(seat: SeatResponse) {
    if (reservedSeats.includes(seat.seat_number)) return
    if (selected.includes(seat.seat_number)) {
      onChange(selected.filter(s => s !== seat.seat_number))
    } else {
      onChange([...selected, seat.seat_number])
    }
  }

  return (
    <div className="space-y-4">
      <div className="overflow-x-auto rounded-lg border p-4">
        {/* Column headers */}
        <div className="mb-1 flex items-center gap-1">
          <div className="w-7 shrink-0" />
          {allCols.map(col => (
            <div key={col} className="flex h-7 w-10 items-center justify-center text-xs font-medium text-muted-foreground">
              {String.fromCharCode(64 + col)}
            </div>
          ))}
        </div>

        {/* Seat rows */}
        <div className="space-y-1">
          {sortedRowNums.map(rowNum => {
            const rowSeats = rows[rowNum]
            const seatByCol = Object.fromEntries(rowSeats.map(s => [s.col, s]))
            return (
              <div key={rowNum} className="flex items-center gap-1">
                <span className="w-7 shrink-0 text-right text-xs text-muted-foreground">{rowNum}</span>
                {allCols.map(col => {
                  const seat = seatByCol[col]
                  if (!seat) {
                    return <div key={col} className="h-8 w-10" />
                  }
                  const isReserved = reservedSeats.includes(seat.seat_number)
                  const isSelected = selected.includes(seat.seat_number)
                  return (
                    <button
                      key={seat.seat_number}
                      type="button"
                      disabled={isReserved}
                      onClick={() => handleClick(seat)}
                      title={`${seat.seat_number} (${CLASS_LABELS[seat.seat_class]})`}
                      className={cn(
                        'flex h-8 w-10 items-center justify-center rounded border text-[10px] font-medium transition-colors',
                        isReserved && 'cursor-not-allowed border-muted-foreground/20 bg-muted text-muted-foreground',
                        !isReserved && isSelected && CLASS_SELECTED[seat.seat_class],
                        !isReserved && !isSelected && CLASS_COLORS[seat.seat_class],
                      )}
                    >
                      {seat.seat_number}
                    </button>
                  )
                })}
              </div>
            )
          })}
        </div>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-x-6 gap-y-2">
        {(['economy', 'business', 'first'] as SeatClass[])
          .filter(cls => fares.some(f => f.seat_class === cls))
          .map(cls => {
            const fare = fares.find(f => f.seat_class === cls)!
            return (
              <div key={cls} className="flex items-center gap-2 text-sm">
                <div className={cn('h-5 w-8 rounded border', CLASS_COLORS[cls])} />
                <span>{CLASS_LABELS[cls]}</span>
                <span className="text-muted-foreground">— {formatPrice(fare.price_cents)}</span>
              </div>
            )
          })}
        <div className="flex items-center gap-2 text-sm">
          <div className="h-5 w-8 rounded border border-muted-foreground/20 bg-muted" />
          <span className="text-muted-foreground">Занято</span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <div className="h-5 w-8 rounded border border-blue-600 bg-blue-500" />
          <span className="text-muted-foreground">Выбрано</span>
        </div>
      </div>
    </div>
  )
}
