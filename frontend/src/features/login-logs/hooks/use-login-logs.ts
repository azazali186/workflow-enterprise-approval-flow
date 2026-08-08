import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { LoginLog } from '@/types/models'
import { loginLogsService, type LoginLogQuery } from '@/services/login-logs.service'

export function useLoginLogsTable(): ServerTable<LoginLog> {
  return useServerTable<LoginLog>({
    queryKey: ['login-logs'],
    fetcher: (params) => loginLogsService.list(params as LoginLogQuery),
    initialSortBy: 'attempted_at',
  })
}
