/**
 * Wire types for the Go backend API (all endpoints are POST /api/v1/...).
 *
 * The backend always responds with a uniform envelope:
 *   { code, message, data }
 *   { code, message, data, pagination }                    (paginated lists)
 *   { code, message, data, pagination, summary }           (paginated + summary)
 *   { code, message, errors: [{ field, message }] }        (validation failures)
 *
 * Pagination is cursor-based: requests carry { cursor, limit, sort_by, sort_order }.
 */

export interface ApiFieldError {
  field: string
  message: string
}

export interface ApiEnvelope<T = unknown> {
  code: number
  message: string
  data?: T
  pagination?: Record<string, unknown> | null
  summary?: Record<string, unknown> | null
  errors?: ApiFieldError[]
}

/** Common cursor-based list request accepted by every list endpoint. */
export interface ListQuery {
  cursor?: string
  limit?: number
  search?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
  /** Some endpoints (e.g. login logs) paginate by page instead of cursor. */
  page?: number
}

export interface SelectOption {
  value: string
  label: string
}
