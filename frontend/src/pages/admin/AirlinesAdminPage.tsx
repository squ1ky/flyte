import { adminCreateAirline } from '@/api/endpoints/admin'
import { listAirlines } from '@/api/endpoints/flight'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useToast } from '@/hooks/use-toast'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

const schema = z.object({
  iata_code: z.string().length(2, 'Ровно 2 символа').toUpperCase(),
  name: z.string().min(1, 'Обязательное поле'),
  country: z.string().min(1, 'Обязательное поле'),
  logo_url: z.string().url('Введите корректный URL').or(z.literal('')),
})

type FormValues = z.infer<typeof schema>

const PAGE_SIZE = 20

export function AirlinesAdminPage() {
  const [open, setOpen] = useState(false)
  const [offset, setOffset] = useState(0)
  const qc = useQueryClient()
  const { toast } = useToast()

  const { data: airlines = [], isLoading } = useQuery({
    queryKey: ['airlines', offset],
    queryFn: () => listAirlines({ limit: PAGE_SIZE, offset }),
  })

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { iata_code: '', name: '', country: '', logo_url: '' },
  })

  const createMutation = useMutation({
    mutationFn: adminCreateAirline,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['airlines'] })
      setOpen(false)
      reset()
      toast({ title: 'Авиакомпания создана' })
    },
    onError: (e: Error) =>
      toast({ title: 'Ошибка', description: e.message, variant: 'destructive' }),
  })

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Авиакомпании</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus className="mr-2 h-4 w-4" />
              Создать
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>Новая авиакомпания</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleSubmit((v) => createMutation.mutate(v))} className="space-y-3">
              <div>
                <Label>IATA-код (2 символа)</Label>
                <Input {...register('iata_code')} placeholder="SU" maxLength={2} />
                {errors.iata_code && (
                  <p className="mt-1 text-xs text-destructive">{errors.iata_code.message}</p>
                )}
              </div>
              <div>
                <Label>Название</Label>
                <Input {...register('name')} placeholder="Aeroflot" />
                {errors.name && (
                  <p className="mt-1 text-xs text-destructive">{errors.name.message}</p>
                )}
              </div>
              <div>
                <Label>Страна</Label>
                <Input {...register('country')} placeholder="Россия" />
                {errors.country && (
                  <p className="mt-1 text-xs text-destructive">{errors.country.message}</p>
                )}
              </div>
              <div>
                <Label>URL логотипа (опционально)</Label>
                <Input {...register('logo_url')} placeholder="https://..." />
                {errors.logo_url && (
                  <p className="mt-1 text-xs text-destructive">{errors.logo_url.message}</p>
                )}
              </div>
              <Button type="submit" className="w-full" disabled={isSubmitting}>
                {isSubmitting ? 'Создание…' : 'Создать'}
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {isLoading ? (
        <p className="text-sm text-muted-foreground">Загрузка…</p>
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>IATA</TableHead>
                <TableHead>Название</TableHead>
                <TableHead>Страна</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {airlines.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="text-muted-foreground">{a.id}</TableCell>
                  <TableCell className="font-mono font-semibold">{a.iata_code}</TableCell>
                  <TableCell>{a.name}</TableCell>
                  <TableCell>{a.country}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          <div className="mt-4 flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={offset === 0}
              onClick={() => setOffset((o) => Math.max(0, o - PAGE_SIZE))}
            >
              Назад
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={airlines.length < PAGE_SIZE}
              onClick={() => setOffset((o) => o + PAGE_SIZE)}
            >
              Вперёд
            </Button>
          </div>
        </>
      )}
    </div>
  )
}
