/**
 * Type definition for pagination request parameters
 */
export type PaginationRequest = {
  /** Number of items per page */
  limit?: number
  /** Page number (1-based) */
  page?: number
}
