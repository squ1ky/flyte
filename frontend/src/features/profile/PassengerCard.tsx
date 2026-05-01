import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { deletePassenger } from '@/api/endpoints/user'
import type { PassengerResponse } from '@/types/user'

const DOC_LABELS: Record<string, string> = {
  passport: 'Паспорт РФ',
  international_passport: 'Загранпаспорт',
}

interface PassengerCardProps {
  passenger: PassengerResponse
  userId: number
}

export function PassengerCard({ passenger, userId }: PassengerCardProps) {
  const qc = useQueryClient()

  const { mutate, isPending } = useMutation({
    mutationFn: () => deletePassenger(userId, passenger.id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['user', userId, 'passengers'] })
      toast.success('Пассажир удалён')
    },
    onError: (err: Error) => {
      toast.error(err.message)
    },
  })

  const fullName = [passenger.last_name, passenger.first_name, passenger.middle_name]
    .filter(Boolean)
    .join(' ')

  const birthDate = new Date(passenger.birth_date).toLocaleDateString('ru-RU')

  return (
    <Card>
      <CardContent className="flex items-start justify-between gap-4 pt-4">
        <div className="flex flex-col gap-1">
          <p className="font-medium">{fullName}</p>
          <p className="text-sm text-muted-foreground">{birthDate}</p>
          <p className="text-sm text-muted-foreground">
            {DOC_LABELS[passenger.document_type] ?? passenger.document_type} ·{' '}
            {passenger.document_number}
          </p>
        </div>
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant="ghost" size="icon" className="shrink-0 text-destructive hover:text-destructive">
              <Trash2 className="h-4 w-4" />
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Удалить пассажира?</AlertDialogTitle>
              <AlertDialogDescription>
                {fullName} будет удалён из вашего профиля. Это действие нельзя отменить.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Отмена</AlertDialogCancel>
              <AlertDialogAction
                onClick={() => mutate()}
                disabled={isPending}
                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              >
                {isPending ? 'Удаляем...' : 'Удалить'}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardContent>
    </Card>
  )
}
