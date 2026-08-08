import { useCallback, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { emptyListMeta } from '@/services/api/normalize'
import type { ServerTable, ServerTableOptions, SortOrder } from './types'

/**
 * Server-driven table state: keeps the cursor stack for cursor pagination,
 * resets to page 1 whenever search/sort/filters/page-size change, and keeps
 * the previous page visible while fetching the next one.
 */
export function useServerTable<T>(options: ServerTableOptions<T>): ServerTable<T> {
  const [page, setPageState] = useState(1)
  const [pageSize, setPageSizeState] = useState(options.initialPageSize ?? 10)
  const [sortBy, setSortBy] = useState<string | undefined>(options.initialSortBy)
  const [sortOrder, setSortOrder] = useState<SortOrder>(options.initialSortOrder ?? 'desc')
  const [search, setSearchState] = useState('')
  const [filters, setFiltersState] = useState<Record<string, unknown>>({})
  const cursors = useRef<string[]>([])

  const setPageSize = useCallback((size: number) => {
    cursors.current = []
    setPageSizeState(size)
    setPageState(1)
  }, [])

  const setSort = useCallback((by: string, order: SortOrder) => {
    cursors.current = []
    setSortBy(by)
    setSortOrder(order)
    setPageState(1)
  }, [])

  const setSearch = useCallback((value: string) => {
    cursors.current = []
    setSearchState(value)
    setPageState(1)
  }, [])

  const setFilters = useCallback((next: Record<string, unknown>) => {
    cursors.current = []
    setFiltersState(next)
    setPageState(1)
  }, [])

  // Cursor that produced the current page (undefined for page 1).
  const cursor = page > 1 ? cursors.current[page - 1] : undefined

  const query = useQuery({
    queryKey: [...options.queryKey, page, pageSize, sortBy, sortOrder, search, filters],
    queryFn: () =>
      options.fetcher({
        limit: pageSize,
        cursor,
        search: search || undefined,
        sort_by: sortBy,
        sort_order: sortOrder,
        page,
        ...filters,
      }),
    enabled: options.enabled ?? true,
    placeholderData: (previous) => previous,
  })

  const data = query.data

  const nextPage = useCallback(() => {
    if (!data?.meta.nextCursor) return
    cursors.current = [...cursors.current, data.meta.nextCursor]
    setPageState((current) => current + 1)
  }, [data])

  const prevPage = useCallback(() => {
    setPageState((current) => Math.max(1, current - 1))
  }, [])

  const setPage = useCallback(
    (target: number) => {
      if (target === page + 1) nextPage()
      else if (target === page - 1) prevPage()
      else if (target === 1) {
        cursors.current = []
        setPageState(1)
      }
    },
    [page, nextPage, prevPage],
  )

  return {
    rows: data?.rows ?? [],
    meta: data?.meta ?? emptyListMeta(page, pageSize),
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    isError: query.isError,
    error: query.error,
    page,
    pageSize,
    sortBy,
    sortOrder,
    search,
    filters,
    setPage,
    setPageSize,
    setSort,
    setSearch,
    setFilters,
    refetch: query.refetch,
  }
}
