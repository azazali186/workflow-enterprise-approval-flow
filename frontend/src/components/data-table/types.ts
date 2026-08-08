import type { ReactNode } from 'react'
import type { ListResult } from '@/services/api/normalize'

export type SortOrder = 'asc' | 'desc'

export interface ColumnDef<T> {
  key: string
  header: ReactNode
  /** When true the header toggles server-side sorting. */
  sortable?: boolean
  /** Backend field sent as `sort_by` (defaults to `key`). */
  sortKey?: string
  align?: 'left' | 'center' | 'right'
  className?: string
  /** Hide the column below a breakpoint so tables adapt on mobile. */
  hideBelow?: 'sm' | 'md' | 'lg' | 'xl'
  render: (row: T) => ReactNode
}

export interface ServerTableOptions<T> {
  /** Base React Query key, e.g. ['users']. */
  queryKey: readonly unknown[]
  fetcher: (params: Record<string, unknown>) => Promise<ListResult<T>>
  initialPageSize?: number
  initialSortBy?: string
  initialSortOrder?: SortOrder
  enabled?: boolean
}

/** The controller returned by useServerTable and consumed by <DataTable>. */
export interface ServerTable<T> {
  rows: T[]
  meta: ListResult<T>['meta']
  isLoading: boolean
  isFetching: boolean
  isError: boolean
  error: Error | null
  page: number
  pageSize: number
  sortBy?: string
  sortOrder: SortOrder
  search: string
  filters: Record<string, unknown>
  setPage: (page: number) => void
  setPageSize: (size: number) => void
  setSort: (by: string, order: SortOrder) => void
  setSearch: (value: string) => void
  setFilters: (filters: Record<string, unknown>) => void
  refetch: () => void
}
