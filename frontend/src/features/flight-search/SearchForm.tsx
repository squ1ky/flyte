import { useEffect } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useSearchParams } from 'react-router-dom'
import { ArrowLeftRight, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { AirportSelect } from './AirportSelect'
import type { SeatClass } from '@/types/flight'

const today = new Date().toISOString().slice(0, 10)

const schema = z.object({
  from: z.string().length(3, 'Выберите аэропорт вылета'),
  to: z.string().length(3, 'Выберите аэропорт прилёта'),
  date: z
    .string()
    .min(1, 'Выберите дату')
    .refine(v => v >= today, 'Дата не может быть в прошлом'),
  passengers: z.coerce.number().min(1, 'Мин. 1').max(9, 'Макс. 9'),
  seat_class: z.enum(['economy', 'business', 'first']).optional(),
})

type FormValues = z.infer<typeof schema>

function paramsToValues(params: URLSearchParams): FormValues {
  return {
    from: params.get('from') ?? '',
    to: params.get('to') ?? '',
    date: params.get('date') ?? '',
    passengers: Number(params.get('passengers') ?? 1),
    seat_class: (params.get('seat_class') as SeatClass | null) ?? undefined,
  }
}

export function SearchForm() {
  const [searchParams, setSearchParams] = useSearchParams()

  const {
    register,
    handleSubmit,
    control,
    setValue,
    getValues,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: paramsToValues(searchParams),
  })

  useEffect(() => {
    reset(paramsToValues(searchParams))
  }, [searchParams, reset])

  function onSubmit(values: FormValues) {
    const params: Record<string, string> = {
      from: values.from,
      to: values.to,
      date: values.date,
      passengers: String(values.passengers),
    }
    if (values.seat_class) params.seat_class = values.seat_class
    setSearchParams(params)
  }

  function swapAirports() {
    const { from, to } = getValues()
    setValue('from', to)
    setValue('to', from)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="rounded-xl border bg-card p-6 shadow-sm">
      <div className="flex flex-wrap items-end gap-3">
        {/* From */}
        <div className="min-w-[160px] flex-1 space-y-1">
          <Label>Откуда</Label>
          <Controller
            name="from"
            control={control}
            render={({ field }) => (
              <AirportSelect
                value={field.value || null}
                onChange={field.onChange}
                placeholder="Аэропорт вылета"
              />
            )}
          />
          {errors.from && <p className="text-xs text-destructive">{errors.from.message}</p>}
        </div>

        {/* Swap */}
        <div className="mb-0.5">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={swapAirports}
            aria-label="Поменять местами"
          >
            <ArrowLeftRight className="h-4 w-4" />
          </Button>
        </div>

        {/* To */}
        <div className="min-w-[160px] flex-1 space-y-1">
          <Label>Куда</Label>
          <Controller
            name="to"
            control={control}
            render={({ field }) => (
              <AirportSelect
                value={field.value || null}
                onChange={field.onChange}
                placeholder="Аэропорт прилёта"
              />
            )}
          />
          {errors.to && <p className="text-xs text-destructive">{errors.to.message}</p>}
        </div>

        {/* Date */}
        <div className="w-40 space-y-1">
          <Label>Дата</Label>
          <Input type="date" min={today} {...register('date')} />
          {errors.date && <p className="text-xs text-destructive">{errors.date.message}</p>}
        </div>

        {/* Passengers */}
        <div className="w-28 space-y-1">
          <Label>Пассажиры</Label>
          <Input type="number" min={1} max={9} {...register('passengers')} />
          {errors.passengers && (
            <p className="text-xs text-destructive">{errors.passengers.message}</p>
          )}
        </div>

        {/* Seat class */}
        <div className="w-36 space-y-1">
          <Label>Класс</Label>
          <Controller
            name="seat_class"
            control={control}
            render={({ field }) => (
              <Select
                value={field.value ?? 'any'}
                onValueChange={(val: string) =>
                  field.onChange(val === 'any' ? undefined : (val as SeatClass))
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="any">Любой класс</SelectItem>
                  <SelectItem value="economy">Эконом</SelectItem>
                  <SelectItem value="business">Бизнес</SelectItem>
                  <SelectItem value="first">Первый</SelectItem>
                </SelectContent>
              </Select>
            )}
          />
        </div>

        {/* Submit */}
        <div className="mb-0.5">
          <Button type="submit">
            <Search className="mr-2 h-4 w-4" />
            Найти
          </Button>
        </div>
      </div>
    </form>
  )
}
