import type { ApiEnvelope, ApiFieldError } from '@/types/api'

export interface ApiErrorOptions {
  status?: number
  code?: number
  isNetwork?: boolean
  fieldErrors?: ApiFieldError[]
}

/**
 * Normalized error for every failed API call. Distinguishes network failures
 * from HTTP errors so UI states can react appropriately.
 */
export class ApiError extends Error {
  readonly status: number
  readonly code: number
  readonly isNetwork: boolean
  readonly fieldErrors?: ApiFieldError[]

  constructor(message: string, options: ApiErrorOptions = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = options.status ?? 0
    this.code = options.code ?? options.status ?? 0
    this.isNetwork = options.isNetwork ?? false
    this.fieldErrors = options.fieldErrors
  }

  static fromResponse(envelope: ApiEnvelope | null, httpStatus: number): ApiError {
    const message = envelope?.message || `Request failed (${httpStatus})`
    return new ApiError(message, {
      status: httpStatus,
      code: envelope?.code ?? httpStatus,
      fieldErrors: envelope?.errors,
    })
  }

  /** Human-readable field errors, e.g. "email: must be a valid email". */
  get fieldMessage(): string | undefined {
    if (!this.fieldErrors?.length) return undefined
    return this.fieldErrors.map((e) => `${e.field}: ${e.message}`).join('\n')
  }

  get isUnauthorized(): boolean {
    return this.status === 401 || this.code === 401
  }

  get isForbidden(): boolean {
    return this.status === 403 || this.code === 403
  }
}

export function toErrorMessage(error: unknown, fallback = 'Something went wrong'): string {
  if (error instanceof ApiError) {
    if (error.fieldMessage) return error.fieldMessage
    return error.message || fallback
  }
  if (error instanceof Error) return error.message || fallback
  return fallback
}
