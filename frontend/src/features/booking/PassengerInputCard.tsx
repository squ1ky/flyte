import { useFormContext } from 'react-hook-form'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { PassengerResponse } from '@/types/user'
import type { SeatClass } from '@/types/flight'

const CLASS_LABELS: Record<SeatClass, string> = {
  economy: 'эконом',
  business: 'бизнес',
  first: 'первый',
}

interface Props {
  index: number
  seatNumber: string
  seatClass: SeatClass
  savedPassengers: PassengerResponse[]
}

export function PassengerInputCard({ index, seatNumber, seatClass, savedPassengers }: Props) {
  const {
    register,
    setValue,
    formState: { errors },
  } = useFormContext()

  const passengerErrors = (errors.passengers as Record<number, Record<string, { message?: string }>> | undefined)?.[index]

  function fillFromSaved(passengerId: string) {
    const p = savedPassengers.find(p => p.id === Number(passengerId))
    if (!p) return
    setValue(`passengers.${index}.first_name`, p.first_name, { shouldValidate: true })
    setValue(`passengers.${index}.last_name`, p.last_name, { shouldValidate: true })
    setValue(`passengers.${index}.document_number`, p.document_number, { shouldValidate: true })
  }

  return (
    <div className="rounded-lg border p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">
          Место {seatNumber}{' '}
          <span className="text-sm font-normal text-muted-foreground">({CLASS_LABELS[seatClass]})</span>
        </h3>

        {savedPassengers.length > 0 && (
          <Select onValueChange={fillFromSaved}>
            <SelectTrigger className="w-[220px]">
              <SelectValue placeholder="Выбрать из сохранённых" />
            </SelectTrigger>
            <SelectContent>
              {savedPassengers.map(p => (
                <SelectItem key={p.id} value={String(p.id)}>
                  {p.first_name} {p.last_name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div className="space-y-1">
          <Label htmlFor={`passengers.${index}.first_name`}>Имя</Label>
          <Input
            id={`passengers.${index}.first_name`}
            placeholder="Иван"
            {...register(`passengers.${index}.first_name`)}
          />
          {passengerErrors?.first_name && (
            <p className="text-xs text-destructive">{passengerErrors.first_name.message}</p>
          )}
        </div>

        <div className="space-y-1">
          <Label htmlFor={`passengers.${index}.last_name`}>Фамилия</Label>
          <Input
            id={`passengers.${index}.last_name`}
            placeholder="Иванов"
            {...register(`passengers.${index}.last_name`)}
          />
          {passengerErrors?.last_name && (
            <p className="text-xs text-destructive">{passengerErrors.last_name.message}</p>
          )}
        </div>

        <div className="space-y-1">
          <Label htmlFor={`passengers.${index}.document_number`}>Номер документа</Label>
          <Input
            id={`passengers.${index}.document_number`}
            placeholder="7712345678"
            {...register(`passengers.${index}.document_number`)}
          />
          {passengerErrors?.document_number && (
            <p className="text-xs text-destructive">{passengerErrors.document_number.message}</p>
          )}
        </div>
      </div>
    </div>
  )
}
