import { DataSourceDto } from '@/core/application/dto/data-source-dto'
import { getFetcher } from '@/lib/fetcher'
import { useQuery } from '@tanstack/react-query'
import { getRuntimeEnv } from '@lerianstudio/console-layout'

const basePath =
  getRuntimeEnv('NEXT_PUBLIC_PLUGIN_UI_BASE_PATH') ??
  process.env.NEXT_PUBLIC_PLUGIN_UI_BASE_PATH

type UseGetDataSourceProps = {
  organizationId: string
  dataSourceId: string
  enabled?: boolean
}

/**
 * Hook for fetching all available data sources for an organization
 */
export const useListDataSources = ({ ...options }) => {
  return useQuery<DataSourceDto[]>({
    queryKey: ['data-sources'],
    queryFn: getFetcher(`${basePath}/api/data-sources`),
    ...options
  })
}

/**
 * Hook for fetching detailed information about a specific data source in an organization
 */
export const useGetDataSource = ({
  organizationId,
  dataSourceId,
  enabled = true
}: UseGetDataSourceProps) => {
  return useQuery<DataSourceDto>({
    queryKey: ['data-sources', organizationId, dataSourceId],
    queryFn: getFetcher(
      `${basePath}/api/organizations/${organizationId}/data-sources/${dataSourceId}`
    ),
    enabled: enabled && !!organizationId && !!dataSourceId
  })
}
