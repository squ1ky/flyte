import { useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { addPassenger } from '@/api/endpoints/user'
import type { DocumentType, Gender } from '@/types/user'

const today = new Date()
today.setHours(0, 0, 0, 0)
const minDate = new Date('1900-01-01')

const schema = z.object({
  first_name: z.string().min(1, 'Обязательное поле'),
  last_name: z.string().min(1, 'Обязательное поле'),
  middle_name: z.string().optional(),
  birth_date: z
    .string()
    .min(1, 'Обязательное поле')
    .refine((v) => {
      const d = new Date(v)
      return d <= today && d >= minDate
    }, 'Некорректная дата рождения'),
  gender: z.enum(['male', 'female']),
  document_type: z.enum(['passport', 'international_passport']),
  document_number: z.string().min(1, 'Обязательное поле'),
  citizenship: z
    .string()
    .length(3, 'Код гражданства — 3 символа ISO (например RUS)')
    .toUpperCase(),
})

type FormValues = z.infer<typeof schema>

interface PassengerFormProps {
  userId: number
  children: React.ReactNode
}

export function PassengerForm({ userId, children }: PassengerFormProps) {
  const [open, setOpen] = useState(false)
  const qc = useQueryClient()

  const {
    register,
    handleSubmit,
    setError,
    setValue,
    control,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { gender: 'male', document_type: 'passport' },
  })

  const { mutate, isPending } = useMutation({
    mutationFn: (data: FormValues) =>
      addPassenger(userId, {
        ...data,
        gender: data.gender as Gender,
        document_type: data.document_type as DocumentType,
        middle_name: data.middle_name || undefined,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['user', userId, 'passengers'] })
      toast.success('Пассажир добавлен')
      reset()
      setOpen(false)
    },
    onError: (err: Error & { status?: number }) => {
      if (err.status === 409) {
        setError('document_number', { message: 'Документ уже добавлен' })
      } else {
        setError('root', { message: err.message })
      }
    },
  })

  const genderValue = useWatch({ control, name: 'gender' })
  const docTypeValue = useWatch({ control, name: 'document_type' })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Добавить пассажира</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit((data) => mutate(data))} className="flex flex-col gap-4 pt-2">
          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="last_name">Фамилия</Label>
              <Input id="last_name" {...register('last_name')} />
              {errors.last_name && (
                <p className="text-sm text-destructive">{errors.last_name.message}</p>
              )}
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="first_name">Имя</Label>
              <Input id="first_name" {...register('first_name')} />
              {errors.first_name && (
                <p className="text-sm text-destructive">{errors.first_name.message}</p>
              )}
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="middle_name">Отчество (необязательно)</Label>
            <Input id="middle_name" {...register('middle_name')} />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="birth_date">Дата рождения</Label>
            <Input
              id="birth_date"
              type="date"
              max={today.toISOString().slice(0, 10)}
              min="1900-01-01"
              {...register('birth_date')}
            />
            {errors.birth_date && (
              <p className="text-sm text-destructive">{errors.birth_date.message}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <Label>Пол</Label>
            <RadioGroup
              value={genderValue}
              onValueChange={(v) => setValue('gender', v as 'male' | 'female')}
              className="flex gap-4"
            >
              <div className="flex items-center gap-2">
                <RadioGroupItem value="male" id="gender-male" />
                <Label htmlFor="gender-male" className="font-normal">
                  Мужской
                </Label>
              </div>
              <div className="flex items-center gap-2">
                <RadioGroupItem value="female" id="gender-female" />
                <Label htmlFor="gender-female" className="font-normal">
                  Женский
                </Label>
              </div>
            </RadioGroup>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label>Тип документа</Label>
              <Select
                value={docTypeValue}
                onValueChange={(v) => setValue('document_type', v as DocumentType)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="passport">Паспорт РФ</SelectItem>
                  <SelectItem value="international_passport">Загранпаспорт</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="document_number">Номер документа</Label>
              <Input id="document_number" {...register('document_number')} />
              {errors.document_number && (
                <p className="text-sm text-destructive">{errors.document_number.message}</p>
              )}
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="citizenship">Гражданство (ISO α-3)</Label>
            <Input
              id="citizenship"
              placeholder="RUS"
              maxLength={3}
              {...register('citizenship')}
            />
            {errors.citizenship && (
              <p className="text-sm text-destructive">{errors.citizenship.message}</p>
            )}
          </div>

          {errors.root && (
            <p className="text-sm text-destructive">{errors.root.message}</p>
          )}

          <Button type="submit" disabled={isPending} className="mt-1">
            {isPending ? 'Сохраняем...' : 'Добавить'}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  )
}
