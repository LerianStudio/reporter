/**
 * DTO for paginated responses from Reporter API
 */
export type ReporterPaginationDto<T> = {
  items: T[]
  limit: number
  page: number
}
