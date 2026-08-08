import { formatNumber } from './format'

export interface RangeInfo {
  from: number
  to: number
  total: number
}

export function rangeInfo(page: number, pageSize: number, total: number): RangeInfo {
  const safePage = Math.max(1, page)
  if (total <= 0) return { from: 0, to: 0, total: 0 }
  return {
    from: (safePage - 1) * pageSize + 1,
    to: Math.min(safePage * pageSize, total),
    total,
  }
}

/** "Showing 1–10 of 120 users" */
export function showingSummary(
  page: number,
  pageSize: number,
  total: number,
  noun: string,
): string {
  if (total <= 0) return `No ${noun} found`
  const { from, to } = rangeInfo(page, pageSize, total)
  return `Showing ${from}–${to} of ${formatNumber(total)} ${noun}`
}
