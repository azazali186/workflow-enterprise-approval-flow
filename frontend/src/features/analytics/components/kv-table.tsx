import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/table'
import { titleCase } from '@/utils/format'

function renderValue(value: unknown): string {
  if (value === null || value === undefined || value === '') return '—'
  if (typeof value === 'number') {
    return Number.isInteger(value) ? value.toLocaleString('en-US') : value.toFixed(2)
  }
  if (typeof value === 'boolean') return value ? 'Yes' : 'No'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

export interface KvTableProps {
  data: Record<string, unknown> | null | undefined
}

/**
 * Renders any analytics payload as a readable table. Works with any shape the
 * backend returns, so unknown or evolving schemas never crash the UI.
 */
export function KvTable({ data }: KvTableProps) {
  const entries = (data ? Object.entries(data) : []).filter(
    ([, value]) => value !== null && value !== undefined && value !== '',
  )

  if (entries.length === 0) {
    return (
      <div className="flex flex-col items-center py-10 text-center">
        <p className="text-sm font-medium text-slate-600">No data available</p>
        <p className="mt-0.5 text-xs text-slate-400">This report has nothing to show yet.</p>
      </div>
    )
  }

  return (
    <div className="overflow-hidden rounded-lg border border-slate-200">
      <Table>
        <THead>
          <tr>
            <TH>Metric</TH>
            <TH>Value</TH>
          </tr>
        </THead>
        <TBody>
          {entries.map(([key, value]) => (
            <TR key={key}>
              <TD className="w-1/2 text-xs font-medium uppercase tracking-wide text-slate-400">
                {titleCase(key)}
              </TD>
              <TD className="font-mono text-xs text-slate-700">{renderValue(value)}</TD>
            </TR>
          ))}
        </TBody>
      </Table>
    </div>
  )
}
