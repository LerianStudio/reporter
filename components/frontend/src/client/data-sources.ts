import { DataSourceDto } from '@/core/application/dto/data-source-dto'
import { getFetcher } from '@/lib/fetcher'
import { useQuery } from '@tanstack/react-query'
import { getRuntimeEnv } from '@lerianstudio/console-layout'

const basePath =
  getRuntimeEnv('NEXT_PUBLIC_REPORTER_UI_BASE_PATH') ??
  process.env.NEXT_PUBLIC_REPORTER_UI_BASE_PATH

type UseListDataSourcesProps = {
  organizationId: string
}

/**
 * Hook for fetching all available data sources for an organization
 */
export const useListDataSources = ({
  organizationId,
  ...options
}: UseListDataSourcesProps) => {
  return useQuery<DataSourceDto[]>({
    queryKey: ['data-sources', organizationId],
    queryFn: getFetcher(
      `${basePath}/api/organizations/${organizationId}/data-sources`
    ),
    ...options
  })
}

type UseGetDataSourceByIdProps = {
  organizationId: string
  dataSourceId: string
}

/**
 * Hook for fetching detailed information about a specific data source
 */
export const useGetDataSourceById = ({
  organizationId,
  dataSourceId
}: UseGetDataSourceByIdProps) => {
  return useQuery<DataSourceDto>({
    queryKey: ['data-sources', organizationId, dataSourceId],
    queryFn: getFetcher(
      `${basePath}/api/organizations/${organizationId}/data-sources/${dataSourceId}`
    )
  })
}
