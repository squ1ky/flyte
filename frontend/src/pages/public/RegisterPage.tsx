import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation } from '@tanstack/react-query'
import { Link, useNavigate } from 'react-router-dom'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { signUp, signIn } from '@/api/endpoints/auth'
import { getUser } from '@/api/endpoints/user'
import { useAuthStore } from '@/store/authStore'

const schema = z.object({
  email: z.string().min(1, 'Обязательное поле').email('Некорректный email'),
  password: z.string().min(8, 'Минимум 8 символов'),
  phone_number: z
    .string()
    .regex(/^\+7\d{10}$/, 'Формат: +7XXXXXXXXXX'),
})

type FormValues = z.infer<typeof schema>

export function RegisterPage() {
  const setAuth = useAuthStore((s) => s.setAuth)
  const navigate = useNavigate()

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) })

  const { mutate, isPending } = useMutation({
    mutationFn: async (data: FormValues) => {
      await signUp(data)
      return signIn({ email: data.email, password: data.password })
    },
    onSuccess: async ({ token, user_id }) => {
      setAuth(token, user_id, 'user')
      try {
        const user = await getUser(user_id)
        setAuth(token, user_id, user.role, user.email)
      } catch {
        // token already set; role/email update best-effort
      }
      navigate('/', { replace: true })
    },
    onError: (err: Error) => {
      if (err.message === 'email already taken') {
        setError('email', { message: 'Email уже занят' })
      } else {
        setError('root', { message: err.message })
      }
    },
  })

  return (
    <div className="flex min-h-[calc(100vh-3.5rem)] items-center justify-center px-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Регистрация</CardTitle>
          <CardDescription>Создайте аккаунт для бронирования билетов</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit((data) => mutate(data))} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" placeholder="you@example.com" {...register('email')} />
              {errors.email && (
                <p className="text-sm text-destructive">{errors.email.message}</p>
              )}
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="password">Пароль</Label>
              <Input id="password" type="password" {...register('password')} />
              {errors.password && (
                <p className="text-sm text-destructive">{errors.password.message}</p>
              )}
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="phone_number">Телефон</Label>
              <Input
                id="phone_number"
                type="tel"
                placeholder="+79991234567"
                {...register('phone_number')}
              />
              {errors.phone_number && (
                <p className="text-sm text-destructive">{errors.phone_number.message}</p>
              )}
            </div>

            {errors.root && (
              <p className="text-sm text-destructive">{errors.root.message}</p>
            )}

            <Button type="submit" disabled={isPending}>
              {isPending ? 'Создаём аккаунт...' : 'Зарегистрироваться'}
            </Button>
          </form>

          <p className="mt-4 text-center text-sm text-muted-foreground">
            Уже есть аккаунт?{' '}
            <Link to="/login" className="underline hover:text-foreground">
              Войти
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
