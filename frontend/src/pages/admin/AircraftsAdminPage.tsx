import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm, useFieldArray } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, Trash2, ChevronLeft } from 'lucide-react'
import { listAircrafts } from '@/api/endpoints/flight'
import { adminCreateAircraft, adminAddAircraftSeats } from '@/api/endpoints/admin'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useToast } from '@/hooks/use-toast'
import { cn } from '@/lib/utils'
import type { AircraftResponse } from '@/types/flight'

// --- Helpers ---

const COL_LETTERS = 'ABCDEFGHIJ'

function generateSeats(
  rows: number,
  cols: number,
  seatClass: 'economy' | 'business' | 'first',
  startRow: number,
) {
  const seats = []
  for (let row = startRow; row < startRow + rows; row++) {
    for (let col = 1; col <= cols; col++) {
      seats.push({
        seat_number: `${row}${COL_LETTERS[col - 1]}`,
        seat_class: seatClass,
        row,
        col,
      })
    }
  }
  return seats
}

function buildAllSeats(sections: SectionItem[]) {
  let currentRow = 1
  return sections.flatMap((s) => {
    const seats = generateSeats(s.rows, s.cols, s.seat_class, currentRow)
    currentRow += s.rows
    return seats
  })
}

// --- Schemas ---

const step1Schema = z.object({
  model: z.string().min(2, 'Обязательное поле'),
  total_seats: z.coerce.number().int().min(1, 'Минимум 1 место'),
})

const sectionSchema = z.object({
  rows: z.coerce.number().int().min(1, 'Мин. 1').max(100),
  cols: z.coerce.number().int().min(1, 'Мин. 1').max(10),
  seat_class: z.enum(['economy', 'business', 'first']),
})

const step2Schema = z.object({
  sections: z.array(sectionSchema).min(1),
})

type Step1Values = z.infer<typeof step1Schema>
type SectionItem = z.infer<typeof sectionSchema>
type Step2Values = z.infer<typeof step2Schema>

// --- Create Aircraft Wizard ---

function CreateAircraftWizard({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
}) {
  const [step, setStep] = useState<1 | 2>(1)
  const [step1Data, setStep1Data] = useState<Step1Values | null>(null)
  const qc = useQueryClient()
  const { toast } = useToast()

  const step1Form = useForm<Step1Values>({
    resolver: zodResolver(step1Schema),
    defaultValues: { model: '', total_seats: 180 },
  })

  const step2Form = useForm<Step2Values>({
    resolver: zodResolver(step2Schema),
    defaultValues: { sections: [{ rows: 30, cols: 6, seat_class: 'economy' }] },
  })

  const { fields, append, remove } = useFieldArray({
    control: step2Form.control,
    name: 'sections',
  })

  const sections = step2Form.watch('sections')
  const totalGenerated = sections.reduce(
    (sum, s) => sum + (Number(s.rows) || 0) * (Number(s.cols) || 0),
    0,
  )
  const totalSeats = step1Data?.total_seats ?? 0

  const mutation = useMutation({
    mutationFn: async (sectionList: SectionItem[]) => {
      const { id } = await adminCreateAircraft(step1Data!)
      const seats = buildAllSeats(sectionList)
      return adminAddAircraftSeats(id, seats)
    },
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ['aircrafts'] })
      onOpenChange(false)
      toast({ title: `Самолёт создан, добавлено ${res.seats_added} мест` })
    },
    onError: (e: Error) =>
      toast({ title: 'Ошибка', description: e.message, variant: 'destructive' }),
  })

  function handleOpenChange(v: boolean) {
    if (!v) {
      setStep(1)
      setStep1Data(null)
      step1Form.reset()
      step2Form.reset({ sections: [{ rows: 30, cols: 6, seat_class: 'economy' }] })
    }
    onOpenChange(v)
  }

  function onStep1Submit(values: Step1Values) {
    setStep1Data(values)
    setStep(2)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {step === 1 ? 'Новый самолёт' : `Места — ${step1Data?.model}`}
          </DialogTitle>
        </DialogHeader>

        {step === 1 && (
          <form onSubmit={step1Form.handleSubmit(onStep1Submit)} className="space-y-3">
            <div>
              <Label>Модель</Label>
              <Input {...step1Form.register('model')} placeholder="Boeing 737-800" />
              {step1Form.formState.errors.model && (
                <p className="mt-1 text-xs text-destructive">
                  {step1Form.formState.errors.model.message}
                </p>
              )}
            </div>
            <div>
              <Label>Всего мест</Label>
              <Input type="number" min={1} {...step1Form.register('total_seats')} />
              {step1Form.formState.errors.total_seats && (
                <p className="mt-1 text-xs text-destructive">
                  {step1Form.formState.errors.total_seats.message}
                </p>
              )}
            </div>
            <Button type="submit" className="w-full">
              Далее →
            </Button>
          </form>
        )}

        {step === 2 && (
          <form
            onSubmit={step2Form.handleSubmit((v) => mutation.mutate(v.sections))}
            className="space-y-4"
          >
            <div
              className={cn(
                'rounded-md px-3 py-2 text-sm font-medium',
                totalGenerated === totalSeats
                  ? 'bg-green-50 text-green-700'
                  : totalGenerated > totalSeats
                    ? 'bg-red-50 text-red-700'
                    : 'bg-amber-50 text-amber-700',
              )}
            >
              {totalGenerated} / {totalSeats} мест
              {totalGenerated > totalSeats && ' — превышает указанное количество'}
              {totalGenerated < totalSeats && totalGenerated > 0 && ' — меньше указанного количества'}
            </div>

            <div className="space-y-2">
              {fields.map((field, idx) => (
                <div key={field.id} className="flex items-end gap-2 rounded-md border p-3">
                  <div className="flex-1 space-y-1">
                    <Label className="text-xs">Класс</Label>
                    <select
                      {...step2Form.register(`sections.${idx}.seat_class`)}
                      className="w-full rounded-md border bg-background px-2 py-1.5 text-sm"
                    >
                      <option value="economy">Economy</option>
                      <option value="business">Business</option>
                      <option value="first">First</option>
                    </select>
                  </div>
                  <div className="w-16 space-y-1">
                    <Label className="text-xs">Рядов</Label>
                    <Input
                      type="number"
                      min={1}
                      max={100}
                      className="h-8"
                      {...step2Form.register(`sections.${idx}.rows`)}
                    />
                  </div>
                  <div className="w-16 space-y-1">
                    <Label className="text-xs">В ряду</Label>
                    <Input
                      type="number"
                      min={1}
                      max={10}
                      className="h-8"
                      {...step2Form.register(`sections.${idx}.cols`)}
                    />
                  </div>
                  <div className="w-14 space-y-1">
                    <Label className="text-xs">Итого</Label>
                    <p className="flex h-8 items-center text-sm text-muted-foreground">
                      {(Number(sections[idx]?.rows) || 0) * (Number(sections[idx]?.cols) || 0)}
                    </p>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 shrink-0 text-muted-foreground"
                    disabled={fields.length === 1}
                    onClick={() => remove(idx)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>

            <Button
              type="button"
              variant="outline"
              size="sm"
              className="w-full"
              onClick={() => append({ rows: 20, cols: 6, seat_class: 'economy' })}
            >
              <Plus className="mr-1 h-3 w-3" />
              Добавить секцию
            </Button>

            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                className="flex-1"
                onClick={() => setStep(1)}
              >
                <ChevronLeft className="mr-1 h-4 w-4" />
                Назад
              </Button>
              <Button type="submit" className="flex-1" disabled={mutation.isPending}>
                {mutation.isPending ? 'Создание…' : 'Создать самолёт'}
              </Button>
            </div>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}

// --- Add Seats to existing aircraft ---

const addSeatsSchema = z.object({
  rows: z.coerce.number().int().min(1, 'Минимум 1 ряд').max(100),
  cols: z.coerce.number().int().min(1).max(10),
  seat_class: z.enum(['economy', 'business', 'first']),
})

type AddSeatsValues = z.infer<typeof addSeatsSchema>

function AddSeatsDialog({ aircraft }: { aircraft: AircraftResponse }) {
  const [open, setOpen] = useState(false)
  const qc = useQueryClient()
  const { toast } = useToast()

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<AddSeatsValues>({
    resolver: zodResolver(addSeatsSchema),
    defaultValues: { rows: 30, cols: 6, seat_class: 'economy' },
  })

  const mutation = useMutation({
    mutationFn: (values: AddSeatsValues) => {
      const seats = generateSeats(values.rows, values.cols, values.seat_class, 1)
      return adminAddAircraftSeats(aircraft.id, seats)
    },
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ['aircrafts'] })
      setOpen(false)
      toast({ title: `Добавлено ${res.seats_added} мест` })
    },
    onError: (e: Error) =>
      toast({ title: 'Ошибка', description: e.message, variant: 'destructive' }),
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">
          Добавить места
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Места для {aircraft.model}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit((v) => mutation.mutate(v))} className="space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label>Рядов</Label>
              <Input type="number" min={1} max={100} {...register('rows')} />
              {errors.rows && (
                <p className="mt-1 text-xs text-destructive">{errors.rows.message}</p>
              )}
            </div>
            <div>
              <Label>Мест в ряду</Label>
              <Input type="number" min={1} max={10} {...register('cols')} />
              {errors.cols && (
                <p className="mt-1 text-xs text-destructive">{errors.cols.message}</p>
              )}
            </div>
          </div>
          <div>
            <Label>Класс</Label>
            <select
              {...register('seat_class')}
              className="mt-1 w-full rounded-md border bg-background px-3 py-2 text-sm"
            >
              <option value="economy">Economy</option>
              <option value="business">Business</option>
              <option value="first">First</option>
            </select>
          </div>
          <Button type="submit" className="w-full" disabled={isSubmitting}>
            {isSubmitting ? 'Добавление…' : 'Сгенерировать и добавить'}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// --- Main page ---

export function AircraftsAdminPage() {
  const [wizardOpen, setWizardOpen] = useState(false)

  const { data: aircrafts = [], isLoading } = useQuery({
    queryKey: ['aircrafts'],
    queryFn: () => listAircrafts({ limit: 100 }),
  })

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Самолёты</h1>
        <Button size="sm" onClick={() => setWizardOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Создать самолёт
        </Button>
      </div>

      <CreateAircraftWizard open={wizardOpen} onOpenChange={setWizardOpen} />

      {isLoading ? (
        <p className="text-sm text-muted-foreground">Загрузка…</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Модель</TableHead>
              <TableHead>Мест</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            {aircrafts.map((a) => (
              <TableRow key={a.id}>
                <TableCell className="text-muted-foreground">{a.id}</TableCell>
                <TableCell>{a.model}</TableCell>
                <TableCell>{a.total_seats}</TableCell>
                <TableCell>
                  <AddSeatsDialog aircraft={a} />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  )
}
