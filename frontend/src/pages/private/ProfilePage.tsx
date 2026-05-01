import { useQuery } from '@tanstack/react-query'
import { UserCircle, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getUser, getPassengers } from '@/api/endpoints/user'
import { useAuthStore } from '@/store/authStore'
import { PassengerForm } from '@/features/profile/PassengerForm'
import { PassengerCard } from '@/features/profile/PassengerCard'

export function ProfilePage() {
  const userId = useAuthStore((s) => s.userId)

  const { data: user, isLoading: userLoading } = useQuery({
    queryKey: ['user', userId ?? 0],
    queryFn: () => getUser(userId!),
    enabled: !!userId,
  })

  const { data: passengers = [], isLoading: passengersLoading } = useQuery({
    queryKey: ['user', userId ?? 0, 'passengers'],
    queryFn: () => getPassengers(userId!),
    enabled: !!userId,
  })

  if (!userId) return null

  return (
    <div className="container mx-auto max-w-2xl px-4 py-8 flex flex-col gap-8">
      <section className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <UserCircle className="h-5 w-5 text-muted-foreground" />
          <h2 className="text-xl font-semibold">Аккаунт</h2>
        </div>
        <Card>
          <CardContent className="pt-4 flex flex-col gap-2">
            {userLoading ? (
              <p className="text-sm text-muted-foreground">Загрузка...</p>
            ) : user ? (
              <>
                <div className="flex gap-2 text-sm">
                  <span className="text-muted-foreground w-24 shrink-0">Email</span>
                  <span>{user.email}</span>
                </div>
                <div className="flex gap-2 text-sm">
                  <span className="text-muted-foreground w-24 shrink-0">Телефон</span>
                  <span>{user.phone_number}</span>
                </div>
              </>
            ) : null}
          </CardContent>
        </Card>
      </section>

      <section className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users className="h-5 w-5 text-muted-foreground" />
            <h2 className="text-xl font-semibold">Пассажиры</h2>
          </div>
          <PassengerForm userId={userId}>
            <Button size="sm">Добавить пассажира</Button>
          </PassengerForm>
        </div>

        {passengersLoading ? (
          <p className="text-sm text-muted-foreground">Загрузка...</p>
        ) : passengers.length === 0 ? (
          <Card>
            <CardHeader>
              <CardTitle className="text-base font-normal text-muted-foreground text-center py-4">
                Пассажиры не добавлены
              </CardTitle>
            </CardHeader>
          </Card>
        ) : (
          <div className="grid gap-3">
            {passengers.map((p) => (
              <PassengerCard key={p.id} passenger={p} userId={userId} />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
