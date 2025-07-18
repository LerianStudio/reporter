/**
 * DTO for paginated responses from Smart Templates API
 */
export type SmartPaginationDto<T> = {
  items: T[]
  limit: number
  page: number
}
