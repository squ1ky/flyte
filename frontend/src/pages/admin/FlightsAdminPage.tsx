import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm, useFieldArray, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, Trash2 } from 'lucide-react'
import { listFlights, listAirlines, listAircrafts } from '@/api/endpoints/flight'
import { adminCreateFlight, adminUpdateFlightStatus } from '@/api/endpoints/admin'
import { AirportSelect } from '@/features/flight-search/AirportSelect'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useToast } from '@/hooks/use-toast'
import type { FlightStatus } from '@/types/flight'

const STATUSES: FlightStatus[] = ['scheduled', 'departed', 'arrived', 'cancelled']

const statusVariant: Record<FlightStatus, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  scheduled: 'default',
  departed: 'secondary',
  arrived: 'outline',
  cancelled: 'destructive',
}

const fareSchema = z.object({
  seat_class: z.enum(['economy', 'business', 'first']),
  price_rub: z.coerce.number().min(1, 'Минимум 1 ₽'),
})

const schema = z.object({
  flight_number: z.string().min(2, 'Обязательное поле'),
  airline_id: z.coerce.number().min(1, 'Выберите авиакомпанию'),
  aircraft_id: z.coerce.number().min(1, 'Выберите самолёт'),
  departure_airport_code: z.string().min(1, 'Выберите аэропорт отправления'),
  arrival_airport_code: z.string().min(1, 'Выберите аэропорт прибытия'),
  departure_time: z.string().min(1, 'Обязательное поле'),
  arrival_time: z.string().min(1, 'Обязательное поле'),
  fares: z.array(fareSchema).min(1, 'Добавьте хотя бы 1 тариф'),
})

type FormValues = z.infer<typeof schema>

export function FlightsAdminPage() {
  const [open, setOpen] = useState(false)
  const qc = useQueryClient()
  const { toast } = useToast()

  const { data: flights = [], isLoading } = useQuery({
    queryKey: ['flights', 'list'],
    queryFn: () => listFlights({ limit: 100 }),
  })

  const { data: airlines = [] } = useQuery({
    queryKey: ['airlines'],
    queryFn: () => listAirlines({ limit: 100 }),
  })

  const { data: aircrafts = [] } = useQuery({
    queryKey: ['aircrafts'],
    queryFn: () => listAircrafts({ limit: 100 }),
  })

  const {
    register,
    handleSubmit,
    control,
    setValue,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { fares: [{ seat_class: 'economy', price_rub: 1000 }] },
  })

  const { fields, append, remove } = useFieldArray({ control, name: 'fares' })

  const createMutation = useMutation({
    mutationFn: (values: FormValues) =>
      adminCreateFlight({
        flight_number: values.flight_number,
        airline_id: values.airline_id,
        aircraft_id: values.aircraft_id,
        departure_airport_code: values.departure_airport_code,
        arrival_airport_code: values.arrival_airport_code,
        departure_time: new Date(values.departure_time).toISOString(),
        arrival_time: new Date(values.arrival_time).toISOString(),
        fare_rules: values.fares.map((f) => ({
          seat_class: f.seat_class,
          price_cents: Math.round(f.price_rub * 100),
        })),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['flights'] })
      setOpen(false)
      reset()
      toast({ title: 'Рейс создан' })
    },
    onError: (e: Error) => toast({ title: 'Ошибка', description: e.message, variant: 'destructive' }),
  })

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: FlightStatus }) =>
      adminUpdateFlightStatus(id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['flights'] }),
    onError: (e: Error) => toast({ title: 'Ошибка', description: e.message, variant: 'destructive' }),
  })

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Рейсы</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus className="mr-2 h-4 w-4" />
              Создать рейс
            </Button>
          </DialogTrigger>
          <DialogContent className="max-h-[90vh] max-w-lg overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Новый рейс</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleSubmit((v) => createMutation.mutate(v))} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div className="col-span-2">
                  <Label>Номер рейса</Label>
                  <Input {...register('flight_number')} placeholder="SU1234" />
                  {errors.flight_number && (
                    <p className="mt-1 text-xs text-destructive">{errors.flight_number.message}</p>
                  )}
                </div>

                <div>
                  <Label>Авиакомпания</Label>
                  <Select onValueChange={(v) => setValue('airline_id', Number(v))}>
                    <SelectTrigger>
                      <SelectValue placeholder="Выберите" />
                    </SelectTrigger>
                    <SelectContent>
                      {airlines.map((a) => (
                        <SelectItem key={a.id} value={String(a.id)}>
                          {a.iata_code} — {a.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {errors.airline_id && (
                    <p className="mt-1 text-xs text-destructive">{errors.airline_id.message}</p>
                  )}
                </div>

                <div>
                  <Label>Самолёт</Label>
                  <Select onValueChange={(v) => setValue('aircraft_id', Number(v))}>
                    <SelectTrigger>
                      <SelectValue placeholder="Выберите" />
                    </SelectTrigger>
                    <SelectContent>
                      {aircrafts.map((a) => (
                        <SelectItem key={a.id} value={String(a.id)}>
                          {a.model}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {errors.aircraft_id && (
                    <p className="mt-1 text-xs text-destructive">{errors.aircraft_id.message}</p>
                  )}
                </div>

                <div>
                  <Label>Аэропорт отправления</Label>
                  <Controller
                    control={control}
                    name="departure_airport_code"
                    render={({ field }) => (
                      <AirportSelect
                        value={field.value || null}
                        onChange={field.onChange}
                        placeholder="Откуда"
                      />
                    )}
                  />
                  {errors.departure_airport_code && (
                    <p className="mt-1 text-xs text-destructive">
                      {errors.departure_airport_code.message}
                    </p>
                  )}
                </div>

                <div>
                  <Label>Аэропорт прибытия</Label>
                  <Controller
                    control={control}
                    name="arrival_airport_code"
                    render={({ field }) => (
                      <AirportSelect
                        value={field.value || null}
                        onChange={field.onChange}
                        placeholder="Куда"
                      />
                    )}
                  />
                  {errors.arrival_airport_code && (
                    <p className="mt-1 text-xs text-destructive">
                      {errors.arrival_airport_code.message}
                    </p>
                  )}
                </div>

                <div>
                  <Label>Вылет</Label>
                  <Input type="datetime-local" {...register('departure_time')} />
                  {errors.departure_time && (
                    <p className="mt-1 text-xs text-destructive">{errors.departure_time.message}</p>
                  )}
                </div>

                <div>
                  <Label>Прилёт</Label>
                  <Input type="datetime-local" {...register('arrival_time')} />
                  {errors.arrival_time && (
                    <p className="mt-1 text-xs text-destructive">{errors.arrival_time.message}</p>
                  )}
                </div>
              </div>

              <div>
                <div className="mb-2 flex items-center justify-between">
                  <Label>Тарифы</Label>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => append({ seat_class: 'economy', price_rub: 1000 })}
                  >
                    <Plus className="mr-1 h-3 w-3" />
                    Добавить
                  </Button>
                </div>
                {fields.map((field, i) => (
                  <div key={field.id} className="mb-2 flex items-center gap-2">
                    <Select
                      defaultValue={field.seat_class}
                      onValueChange={(v) =>
                        setValue(
                          `fares.${i}.seat_class`,
                          v as 'economy' | 'business' | 'first',
                        )
                      }
                    >
                      <SelectTrigger className="w-36">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="economy">Economy</SelectItem>
                        <SelectItem value="business">Business</SelectItem>
                        <SelectItem value="first">First</SelectItem>
                      </SelectContent>
                    </Select>
                    <Input
                      type="number"
                      min={1}
                      placeholder="Цена, ₽"
                      {...register(`fares.${i}.price_rub`)}
                      className="flex-1"
                    />
                    {fields.length > 1 && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => remove(i)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                ))}
                {errors.fares && (
                  <p className="mt-1 text-xs text-destructive">{errors.fares.message}</p>
                )}
              </div>

              <Button type="submit" className="w-full" disabled={isSubmitting}>
                {isSubmitting ? 'Создание…' : 'Создать рейс'}
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {isLoading ? (
        <p className="text-sm text-muted-foreground">Загрузка…</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Рейс</TableHead>
              <TableHead>Авиакомпания</TableHead>
              <TableHead>Маршрут</TableHead>
              <TableHead>Вылет</TableHead>
              <TableHead>Статус</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {flights.map((f) => (
              <TableRow key={f.id}>
                <TableCell className="font-mono">{f.flight_number}</TableCell>
                <TableCell>{f.airline?.name ?? '—'}</TableCell>
                <TableCell>
                  {f.departure_airport?.code} → {f.arrival_airport?.code}
                </TableCell>
                <TableCell>
                  {new Date(f.departure_time).toLocaleString('ru-RU', {
                    day: '2-digit',
                    month: '2-digit',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </TableCell>
                <TableCell>
                  <Select
                    defaultValue={f.status}
                    onValueChange={(v) =>
                      statusMutation.mutate({ id: f.id, status: v as FlightStatus })
                    }
                  >
                    <SelectTrigger className="h-7 w-32 text-xs">
                      <Badge variant={statusVariant[f.status]} className="text-xs">
                        {f.status}
                      </Badge>
                    </SelectTrigger>
                    <SelectContent>
                      {STATUSES.map((s) => (
                        <SelectItem key={s} value={s}>
                          {s}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  )
}
