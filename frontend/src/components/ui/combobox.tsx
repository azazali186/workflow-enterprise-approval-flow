import { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import { Check, ChevronDown, Search, X } from 'lucide-react'
import { cn } from '@/utils/cn'

export interface ComboboxOption {
  id: string
  name: string
}

export interface ComboboxProps {
  options: ComboboxOption[]
  value?: string
  onChange?: (value: string) => void
  placeholder?: string
  searchPlaceholder?: string
  emptyText?: string
  disabled?: boolean
  invalid?: boolean
  error?: string
  className?: string
  id?: string
  name?: string
  required?: boolean
}

export function Combobox({
  options,
  value,
  onChange,
  placeholder = 'Select...',
  searchPlaceholder = 'Search...',
  emptyText = 'No results found',
  disabled = false,
  invalid = false,
  error,
  className,
  id,
  name,
  required,
}: ComboboxProps) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  // -1 = nothing highlighted until the user navigates (standard combobox behavior)
  const [highlightedIndex, setHighlightedIndex] = useState(-1)
  const containerRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const listboxId = useMemo(() => `listbox-${id ?? Math.random().toString(36).slice(2)}`, [id])

  const selectedOption = options.find((opt) => opt.id === value)

  const filteredOptions = useMemo(
    () =>
      options.filter((opt) => opt.name.toLowerCase().includes(search.toLowerCase())),
    [options, search],
  )

  // Close on outside click
  useEffect(() => {
    if (!open) return
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setOpen(false)
        setSearch('')
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [open])

  // Reset highlighted index when search changes
  useEffect(() => {
    setHighlightedIndex(-1)
  }, [search])

  // Scroll highlighted item into view
  useEffect(() => {
    if (!open) return
    const listbox = containerRef.current?.querySelector('[role="listbox"]')
    const highlighted = listbox?.querySelector('[data-highlighted="true"]')
    if (highlighted) {
      highlighted.scrollIntoView({ block: 'nearest' })
    }
  }, [highlightedIndex, open])

  // Focus the search input when the dropdown opens
  useEffect(() => {
    if (open) {
      requestAnimationFrame(() => inputRef.current?.focus())
    }
  }, [open])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (disabled) return

      switch (e.key) {
        case 'Enter':
          e.preventDefault()
          if (open && filteredOptions[highlightedIndex]) {
            onChange?.(filteredOptions[highlightedIndex].id)
            setOpen(false)
            setSearch('')
          } else if (!open) {
            setOpen(true)
          }
          break
        case 'Escape':
          if (open) {
            e.preventDefault()
            e.stopPropagation()
            setOpen(false)
            setSearch('')
          }
          break
        case 'ArrowDown':
          e.preventDefault()
          if (!open) {
            setOpen(true)
          } else {
            setHighlightedIndex((prev) =>
              prev < filteredOptions.length - 1 ? prev + 1 : prev,
            )
          }
          break
        case 'ArrowUp':
          e.preventDefault()
          if (open) {
            setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : 0))
          }
          break
        case 'Backspace':
          if (!search && value) {
            onChange?.('')
          }
          break
        case 'Tab':
          setOpen(false)
          setSearch('')
          break
      }
    },
    [disabled, open, filteredOptions, highlightedIndex, onChange, value, search],
  )

  const handleSelect = useCallback(
    (id: string) => {
      onChange?.(id)
      setOpen(false)
      setSearch('')
    },
    [onChange],
  )

  const handleClear = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation()
      onChange?.('')
      setSearch('')
      inputRef.current?.focus()
    },
    [onChange],
  )

  return (
    <div ref={containerRef} className={cn('relative', className)} id={id} onKeyDown={handleKeyDown}>
      {/* Trigger button */}
      <button
        type="button"
        role="combobox"
        aria-expanded={open}
        aria-haspopup="listbox"
        aria-controls={listboxId}
        aria-invalid={invalid || undefined}
        aria-describedby={error ? `${id}-error` : undefined}
        disabled={disabled}
        onClick={() => {
          if (!disabled) setOpen((prev) => !prev)
        }}
        className={cn(
          'flex h-9 w-full items-center justify-between rounded-lg border bg-white pl-3 pr-2 text-sm shadow-sm transition-colors',
          'hover:border-slate-300 focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/25',
          'disabled:cursor-not-allowed disabled:opacity-50',
          invalid
            ? 'border-rose-300 focus:border-rose-500 focus:ring-rose-500/25'
            : 'border-slate-200',
          !selectedOption && 'text-slate-400',
        )}
      >
        <span className="truncate">
          {selectedOption ? selectedOption.name : placeholder}
        </span>
        <div className="flex items-center gap-1">
          {selectedOption && !disabled && (
            <X
              className="h-4 w-4 shrink-0 text-slate-400 hover:text-slate-600"
              onClick={handleClear}
            />
          )}
          <ChevronDown
            className={cn(
              'h-4 w-4 shrink-0 text-slate-400 transition-transform',
              open && 'rotate-180',
            )}
          />
        </div>
      </button>

      {/* Error message */}
      {error && (
        <p id={`${id}-error`} className="mt-1 text-xs text-rose-500">
          {error}
        </p>
      )}

      {/* Dropdown */}
      {open && (
        <div className="absolute z-50 mt-1 w-full overflow-hidden rounded-lg border border-slate-200 bg-white shadow-lg">
          {/* Search input */}
          <div className="flex items-center border-b border-slate-200 px-3">
            <Search className="h-4 w-4 shrink-0 text-slate-400" />
            {/* Note: keydown handling lives on the container div — the input's
                events bubble up to it, so attaching onKeyDown here as well would
                fire the handler twice per keystroke. */}
            <input
              ref={inputRef}
              type="text"
              placeholder={searchPlaceholder}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-9 flex-1 bg-transparent pl-2 text-sm outline-none placeholder:text-slate-400"
              aria-label={searchPlaceholder}
              autoComplete="off"
              role="searchbox"
            />
          </div>

          {/* Options list */}
          <ul
            id={listboxId}
            role="listbox"
            className="max-h-60 overflow-auto py-1"
            aria-label={placeholder}
          >
            {filteredOptions.length === 0 ? (
              <li className="px-3 py-2 text-sm text-slate-500" role="option" aria-disabled>
                {emptyText}
              </li>
            ) : (
              filteredOptions.map((option, index) => (
                <li
                  key={option.id}
                  role="option"
                  aria-selected={option.id === value}
                  data-highlighted={index === highlightedIndex}
                  onClick={() => handleSelect(option.id)}
                  onMouseEnter={() => setHighlightedIndex(index)}
                  className={cn(
                    'flex cursor-pointer items-center gap-2 px-3 py-2 text-sm transition-colors',
                    'hover:bg-slate-100',
                    option.id === value && 'bg-primary-50 text-primary-700',
                    index === highlightedIndex && option.id !== value && 'bg-slate-100',
                  )}
                >
                  <span className="flex-1 truncate">{option.name}</span>
                  {option.id === value && (
                    <Check className="h-4 w-4 shrink-0 text-primary-500" />
                  )}
                </li>
              ))
            )}
          </ul>
        </div>
      )}

      {/* Hidden input for form submission */}
      {name && <input type="hidden" value={value || ''} name={name} required={required} />}
    </div>
  )
}
