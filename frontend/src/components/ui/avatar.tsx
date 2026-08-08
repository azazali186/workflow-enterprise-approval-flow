import { useMemo } from 'react'
import { cn } from '@/utils/cn'
import { initials } from '@/utils/format'

const palettes = [
  'from-primary-500 to-violet-600',
  'from-sky-500 to-primary-600',
  'from-emerald-500 to-teal-600',
  'from-amber-500 to-orange-600',
  'from-rose-500 to-pink-600',
  'from-violet-500 to-purple-600',
]

function paletteFor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i += 1) {
    hash = (hash * 31 + name.charCodeAt(i)) >>> 0
  }
  return palettes[hash % palettes.length]
}

export interface AvatarProps extends React.HTMLAttributes<HTMLDivElement> {
  name: string
  size?: 'sm' | 'md' | 'lg'
}

const sizes = {
  sm: 'h-7 w-7 text-[10px]',
  md: 'h-9 w-9 text-xs',
  lg: 'h-11 w-11 text-sm',
}

export function Avatar({ name, size = 'md', className, ...props }: AvatarProps) {
  const gradient = useMemo(() => paletteFor(name || '?'), [name])
  return (
    <div
      className={cn(
        'flex shrink-0 items-center justify-center rounded-full bg-gradient-to-br font-semibold text-white shadow-sm ring-2 ring-white',
        gradient,
        sizes[size],
        className,
      )}
      title={name}
      aria-label={name}
      {...props}
    >
      {initials(name) || '?'}
    </div>
  )
}
