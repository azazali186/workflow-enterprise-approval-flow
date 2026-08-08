import { useEffect, useState, type ReactNode } from 'react'
import { ArrowDown, ArrowUp, ArrowUpDown, SearchX } from 'lucide-react'
import { cn } from '@/utils/cn'
import { Skeleton } from '@/components/ui/skeleton'
import { EmptyState } from '@/components/ui/empty-state'
import { ErrorState } from '@/components/ui/error-state'
import { PaginationBar } from '@/components/ui/pagination-bar'
import { SearchInput } from '@/components/ui/search-input'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/table'
import { useDebounce } from '@/hooks/use-debounce'
import type { ColumnDef, ServerTable } from './types'

const hideClasses: Record<NonNullable<ColumnDef<unknown>['hideBelow']>, string> = {
  sm: 'hidden sm:table-cell',
  md: 'hidden md:table-cell',
  lg: 'hidden lg:table-cell',
  xl: 'hidden xl:table-cell',
}

export interface DataTableProps<T> {
  table: ServerTable<T>
  columns: ColumnDef<T>[]
  rowKey: (row: T) => string
  /** Singular/plural noun for summaries, e.g. "users". */
  noun: string
  searchable?: boolean
  searchPlaceholder?: string
  title?: ReactNode
  description?: ReactNode
  toolbar?: ReactNode
  emptyTitle?: string
  emptyDescription?: string
  onRowClick?: (row: T) => void
}

export function DataTable<T>({
  table,
  columns,
  rowKey,
  noun,
  searchable = true,
  searchPlaceholder = 'Search…',
  title,
  description,
  toolbar,
  emptyTitle,
  emptyDescription,
  onRowClick,
}: DataTableProps<T>) {
  const [query, setQuery] = useState(table.search)
  const debouncedQuery = useDebounce(query, 350)

  useEffect(() => {
    if (debouncedQuery !== table.search) table.setSearch(debouncedQuery)
  }, [debouncedQuery, table.search, table])

  const handleSort = (column: ColumnDef<T>) => {
    if (!column.sortable) return
    const sortKey = column.sortKey ?? column.key
    const nextOrder = table.sortBy === sortKey && table.sortOrder === 'asc' ? 'desc' : 'asc'
    table.setSort(sortKey, nextOrder)
  }

  const renderSortIcon = (column: ColumnDef<T>) => {
    if (!column.sortable) return null
    const sortKey = column.sortKey ?? column.key
    if (table.sortBy === sortKey) {
      return table.sortOrder === 'asc' ? (
        <ArrowUp className="h-3 w-3 text-primary-600" />
      ) : (
        <ArrowDown className="h-3 w-3 text-primary-600" />
      )
    }
    return <ArrowUpDown className="h-3 w-3 text-slate-300 group-hover:text-slate-400" />
  }

  return (
    <div className="overflow-hidden rounded-xl border border-slate-200/80 bg-white shadow-card">
      {(title || searchable || toolbar) && (
        <div className="flex flex-col gap-3 border-b border-slate-100 px-4 py-3.5 sm:flex-row sm:items-center sm:justify-between">
          <div className="min-w-0">
            {title && <h2 className="text-sm font-semibold text-slate-900">{title}</h2>}
            {description && <p className="mt-0.5 text-xs text-slate-500">{description}</p>}
          </div>
          <div className="flex items-center gap-2">
            {searchable && (
              <SearchInput
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                onClear={() => setQuery('')}
                placeholder={searchPlaceholder}
                className="w-56 sm:w-64"
              />
            )}
            {toolbar}
          </div>
        </div>
      )}

      <div className="h-0.5 overflow-hidden bg-slate-100">
        <div
          className={cn(
            'h-full bg-primary-500 transition-all duration-300',
            table.isFetching ? 'w-2/3 animate-pulse' : 'w-0',
          )}
        />
      </div>

      <div className="relative">
        <Table>
          <THead>
            <tr>
              {columns.map((column) => (
                <TH
                  key={column.key}
                  align={column.align}
                  className={cn(
                    column.hideBelow && hideClasses[column.hideBelow],
                    column.sortable &&
                      'group cursor-pointer select-none hover:text-slate-700',
                    column.className,
                  )}
                  onClick={column.sortable ? () => handleSort(column) : undefined}
                  aria-sort={
                    column.sortable && table.sortBy === (column.sortKey ?? column.key)
                      ? table.sortOrder === 'asc'
                        ? 'ascending'
                        : 'descending'
                      : undefined
                  }
                >
                  <span className="inline-flex items-center gap-1.5">
                    {column.header}
                    {renderSortIcon(column)}
                  </span>
                </TH>
              ))}
            </tr>
          </THead>
          <TBody>
            {table.isLoading ? (
              Array.from({ length: Math.min(table.pageSize || 5, 8) }).map((_, index) => (
                <TR key={`skeleton-${index}`}>
                  {columns.map((column) => (
                    <TD
                      key={column.key}
                      className={cn(column.hideBelow && hideClasses[column.hideBelow])}
                    >
                      <Skeleton className="h-4 w-3/4 max-w-[180px]" />
                    </TD>
                  ))}
                </TR>
              ))
            ) : table.isError ? (
              <tr>
                <td colSpan={columns.length}>
                  <ErrorState
                    title={`Couldn't load ${noun}`}
                    error={table.error}
                    onRetry={table.refetch}
                  />
                </td>
              </tr>
            ) : table.rows.length === 0 ? (
              <tr>
                <td colSpan={columns.length}>
                  <EmptyState
                    icon={SearchX}
                    title={emptyTitle ?? `No ${noun} found`}
                    description={
                      emptyDescription ??
                      (table.search
                        ? `Nothing matches “${table.search}”. Try a different search or clear the filters.`
                        : `There are no ${noun} to show yet.`)
                    }
                  />
                </td>
              </tr>
            ) : (
              table.rows.map((row) => (
                <TR
                  key={rowKey(row)}
                  className={onRowClick ? 'cursor-pointer' : undefined}
                  onClick={onRowClick ? () => onRowClick(row) : undefined}
                >
                  {columns.map((column) => (
                    <TD
                      key={column.key}
                      align={column.align}
                      className={cn(
                        column.hideBelow && hideClasses[column.hideBelow],
                        column.className,
                      )}
                    >
                      {column.render(row)}
                    </TD>
                  ))}
                </TR>
              ))
            )}
          </TBody>
        </Table>
      </div>

      <div className="border-t border-slate-100">
        <PaginationBar
          page={table.page}
          pageSize={table.pageSize}
          total={table.meta.total}
          hasMore={table.meta.hasMore}
          hasPrevious={table.meta.hasPrevious}
          noun={noun}
          onPageChange={table.setPage}
          onPageSizeChange={table.setPageSize}
        />
      </div>
    </div>
  )
}
