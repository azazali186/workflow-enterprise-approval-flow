import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts'
import { titleCase } from '@/utils/format'

const STATUS_COLORS: Record<string, string> = {
  approved: '#10b981',
  rejected: '#f43f5e',
  pending: '#f59e0b',
  in_review: '#0ea5e9',
  submitted: '#6366f1',
  escalated: '#f97316',
  completed: '#14b8a6',
  draft: '#94a3b8',
  cancelled: '#64748b',
}

const FALLBACK_COLORS = ['#6366f1', '#8b5cf6', '#0ea5e9', '#10b981', '#f59e0b', '#f43f5e', '#94a3b8']

export interface StatusDatum {
  name: string
  value: number
  color: string
}

export interface StatusDonutProps {
  data: StatusDatum[]
  totalLabel?: string
}

export function StatusDonut({ data, totalLabel = 'Total' }: StatusDonutProps) {
  const total = data.reduce((sum, item) => sum + item.value, 0)

  if (data.length === 0) {
    return (
      <div className="flex h-56 items-center justify-center text-sm text-slate-400">
        No data available
      </div>
    )
  }

  return (
    <div className="relative h-56">
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={data}
            dataKey="value"
            nameKey="name"
            innerRadius={62}
            outerRadius={88}
            paddingAngle={2}
            strokeWidth={2}
          >
            {data.map((item) => (
              <Cell key={item.name} fill={item.color} />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              borderRadius: 12,
              border: '1px solid rgb(226 232 240)',
              boxShadow: '0 12px 32px -8px rgb(15 23 42 / 0.18)',
              fontSize: 12,
            }}
            formatter={(value) => [`${value}`, 'Count']}
          />
        </PieChart>
      </ResponsiveContainer>
      <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
        <span className="text-2xl font-bold tabular-nums text-slate-900">{total}</span>
        <span className="text-xs text-slate-400">{totalLabel}</span>
      </div>
      <div className="mt-2 flex flex-wrap justify-center gap-x-3 gap-y-1">
        {data.map((item) => (
          <span key={item.name} className="flex items-center gap-1.5 text-xs text-slate-500">
            <span className="h-2 w-2 rounded-full" style={{ backgroundColor: item.color }} />
            {titleCase(item.name)}
          </span>
        ))}
      </div>
    </div>
  )
}

/** Convert a backend `{ "status": count }` map into chart data. */
export function statusMapToData(source: Record<string, unknown> | null | undefined): StatusDatum[] {
  if (!source) return []
  return Object.entries(source)
    .filter(([, value]) => typeof value === 'number' && value > 0)
    .map(([key, value], index) => ({
      name: key,
      value: value as number,
      color: STATUS_COLORS[key.toLowerCase()] ?? FALLBACK_COLORS[index % FALLBACK_COLORS.length],
    }))
}
