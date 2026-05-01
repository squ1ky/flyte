import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Check, ChevronsUpDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { searchAirports } from '@/api/endpoints/flight'
import { useDebounce } from './hooks/useDebounce'

interface Props {
  value: string | null
  onChange: (code: string) => void
  placeholder: string
  disabled?: boolean
}

export function AirportSelect({ value, onChange, placeholder, disabled }: Props) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const debouncedQuery = useDebounce(query, 300)

  const { data: airports = [] } = useQuery({
    queryKey: ['airports', debouncedQuery],
    queryFn: () => searchAirports(debouncedQuery),
    enabled: debouncedQuery.length >= 1,
  })

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className="w-full justify-between font-normal"
        >
          <span className="truncate">{value ?? placeholder}</span>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[320px] p-0" align="start">
        <Command shouldFilter={false}>
          <CommandInput
            placeholder="Поиск по названию или коду..."
            value={query}
            onValueChange={setQuery}
          />
          <CommandList>
            {debouncedQuery.length < 1 && (
              <CommandEmpty>Начните вводить название</CommandEmpty>
            )}
            {debouncedQuery.length >= 1 && airports.length === 0 && (
              <CommandEmpty>Аэропорт не найден</CommandEmpty>
            )}
            <CommandGroup>
              {airports.map(airport => (
                <CommandItem
                  key={airport.code}
                  value={airport.code}
                  onSelect={() => {
                    onChange(airport.code)
                    setOpen(false)
                    setQuery('')
                  }}
                >
                  <Check
                    className={cn(
                      'mr-2 h-4 w-4',
                      value === airport.code ? 'opacity-100' : 'opacity-0',
                    )}
                  />
                  {airport.code} — {airport.name} ({airport.city})
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
