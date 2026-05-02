import { useEffect, useRef, useState } from 'react'

interface Props {
  expiresAt: string
  onExpire: () => void
}

export function ExpiryCountdown({ expiresAt, onExpire }: Props) {
  const [remaining, setRemaining] = useState(() => new Date(expiresAt).getTime() - Date.now())
  const onExpireRef = useRef(onExpire)

  useEffect(() => {
    onExpireRef.current = onExpire
  })

  useEffect(() => {
    const interval = setInterval(() => {
      const r = new Date(expiresAt).getTime() - Date.now()
      setRemaining(r)
      if (r <= 0) {
        clearInterval(interval)
        onExpireRef.current()
      }
    }, 1000)
    return () => clearInterval(interval)
  }, [expiresAt])

  if (remaining <= 0) {
    return <p className="text-sm font-medium text-destructive">Время бронирования истекло</p>
  }

  const totalSeconds = Math.floor(remaining / 1000)
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60

  return (
    <p className="text-sm font-medium text-amber-600">
      Осталось {String(minutes).padStart(2, '0')}:{String(seconds).padStart(2, '0')}
    </p>
  )
}
