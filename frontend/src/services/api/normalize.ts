import type { ApiEnvelope } from '@/types/api'

export interface TableMeta {
  total: number
  page: number
  pageSize: number
  totalPages: number
  hasMore: boolean
  hasPrevious: boolean
  nextCursor?: string
  previousCursor?: string
}

export interface ListResult<T> {
  rows: T[]
  meta: TableMeta
}

export function emptyListMeta(page = 1, pageSize = 10): TableMeta {
  return {
    total: 0,
    page,
    pageSize,
    totalPages: 1,
    hasMore: false,
    hasPrevious: false,
  }
}

function asRecord(value: unknown): Record<string, unknown> | null {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    return value as Record<string, unknown>
  }
  return null
}

function toNumber(value: unknown, fallback = 0): number {
  const n = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(n) ? n : fallback
}

/**
 * Normalize any backend list response into { rows, meta }.
 *
 * Handles every shape the backend produces:
 *  - data is a plain array (e.g. /admin/users, /admin/roles)
 *  - data + pagination / summary at the envelope level
 *  - data is a nested ListResponse { data, pagination, summary }
 */
export function normalizeList<T>(envelope: ApiEnvelope): ListResult<T> {
  let rows: T[] = []
  const data = envelope.data

  if (Array.isArray(data)) {
    rows = data as T[]
  } else {
    const nested = asRecord(data)
    if (nested && Array.isArray(nested.data)) {
      rows = nested.data as T[]
    }
  }

  const pagination = asRecord(envelope.pagination) ?? asRecord(envelope.summary) ?? {}
  const summary = asRecord(envelope.summary) ?? {}
  const meta = { ...pagination, ...summary }

  const total = toNumber(meta.total_count ?? meta.total_records ?? meta.filtered_records, rows.length)
  const page = toNumber(meta.page, 1)
  const pageSize = toNumber(meta.page_size, rows.length || 10)
  const totalPages = toNumber(
    meta.total_pages,
    total > 0 ? Math.max(1, Math.ceil(total / pageSize)) : 1,
  )
  const nextCursor =
    typeof meta.next_cursor === 'string' && meta.next_cursor ? meta.next_cursor : undefined
  const previousCursor =
    typeof meta.previous_cursor === 'string' && meta.previous_cursor
      ? meta.previous_cursor
      : undefined
  const hasMore = typeof meta.has_more === 'boolean' ? meta.has_more : Boolean(nextCursor)
  const hasPrevious =
    typeof meta.has_previous === 'boolean' ? meta.has_previous : Boolean(previousCursor) || page > 1

  return {
    rows,
    meta: {
      total,
      page,
      pageSize,
      totalPages,
      hasMore,
      hasPrevious,
      nextCursor,
      previousCursor,
    },
  }
}
