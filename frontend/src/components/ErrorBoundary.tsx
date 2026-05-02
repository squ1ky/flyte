import { Component, type ReactNode, type ErrorInfo } from 'react'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false }

  static getDerivedStateFromError(): State {
    return { hasError: true }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-screen flex-col items-center justify-center gap-4 px-4 text-center">
          <h1 className="text-2xl font-semibold">Что-то пошло не так</h1>
          <p className="max-w-sm text-muted-foreground">
            Произошла непредвиденная ошибка. Попробуйте перезагрузить страницу.
          </p>
          <Button onClick={() => window.location.reload()}>Перезагрузить страницу</Button>
        </div>
      )
    }
    return this.props.children
  }
}
