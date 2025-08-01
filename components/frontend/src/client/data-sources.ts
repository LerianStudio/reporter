import { DataSourceDto } from '@/core/application/dto/data-source-dto'
import { getFetcher } from '@/lib/fetcher'
import { useQuery } from '@tanstack/react-query'
import { getRuntimeEnv } from '@lerianstudio/console-layout'

const basePath =
  getRuntimeEnv('NEXT_PUBLIC_PLUGIN_UI_BASE_PATH') ??
  process.env.NEXT_PUBLIC_PLUGIN_UI_BASE_PATH

type UseGetDataSourceByIdProps = {
  dataSourceId: string
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
 * Hook for fetching detailed information about a specific data source
 */
export const useGetDataSourceById = ({
  dataSourceId
}: UseGetDataSourceByIdProps) => {
  return useQuery<DataSourceDto>({
    queryKey: ['data-sources', dataSourceId],
    queryFn: getFetcher(`${basePath}/api/data-sources/${dataSourceId}`)
  })
}
