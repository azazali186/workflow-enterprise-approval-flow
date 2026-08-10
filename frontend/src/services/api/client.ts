import { store } from '@/store'
import { clearCredentials } from '@/store/slices/auth.slice'
import { ApiError } from './errors'
import { normalizeList, type ListResult } from './normalize'
import type { ApiEnvelope, AuthResult } from '@/types'
import { storage } from '@/utils/storage'

const API_BASE = (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/+$/, '')

export interface RequestOptions {
  body?: unknown
  signal?: AbortSignal
  skipAuth?: boolean
}

let refreshing: Promise<string | null> | null = null

/**
 * True when the access token is already expired (with a small safety margin),
 * so we refresh before the request instead of after a wasted 401 round-trip.
 */
function isAccessTokenExpired(): boolean {
  const expiresAt = storage.getExpiresAt()
  if (!expiresAt) return false
  const expiry = Date.parse(expiresAt)
  if (Number.isNaN(expiry)) return false
  // Refresh 60s before the hard expiry to avoid a 401 mid-flight.
  return Date.now() >= expiry - 60_000
}

/** Single-flight token refresh: concurrent 401s share one refresh call. */
async function refreshAccessToken(): Promise<string | null> {
  if (refreshing) return refreshing

  refreshing = (async () => {
    const refreshToken = storage.getRefreshToken()
    if (!refreshToken) return null
    try {
      const res = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })
      const envelope = (await res.json()) as ApiEnvelope<AuthResult>
      if (!res.ok || !envelope.data?.access_token) return null
      storage.setAccessToken(envelope.data.access_token)
      storage.setRefreshToken(envelope.data.refresh_token)
      storage.setExpiresAt(envelope.data.expires_at)
      return envelope.data.access_token
    } catch {
      return null
    } finally {
      refreshing = null
    }
  })()

  return refreshing
}

function sessionExpired(): void {
  storage.clearAuth()
  store.dispatch(clearCredentials())
}

/**
 * Returns a valid access token, refreshing first if the stored one is past
 * its expiry (the proactive path used by the API client and the WebSocket
 * client before a handshake). Returns null — and clears the session — when
 * the refresh itself fails, since an expired session cannot be revived.
 */
export async function ensureFreshAccessToken(): Promise<string | null> {
  if (!isAccessTokenExpired()) return storage.getAccessToken()
  const fresh = await refreshAccessToken()
  if (!fresh) {
    sessionExpired()
    return null
  }
  return fresh
}

async function performEnvelope<T>(
  path: string,
  opts: RequestOptions,
  attempt: 'first' | 'retry',
): Promise<ApiEnvelope<T>> {
  // Proactive refresh: if the access token is already expired, swap it for a
  // fresh one before the request goes out — no 401 round-trip needed.
  if (attempt === 'first' && !opts.skipAuth && isAccessTokenExpired()) {
    const newToken = await refreshAccessToken()
    if (!newToken) {
      sessionExpired()
      throw new ApiError('Your session has expired. Please sign in again.', { status: 401 })
    }
  }

  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (!opts.skipAuth) {
    const token = storage.getAccessToken()
    if (token) headers.Authorization = `Bearer ${token}`
  }

  let res: Response
  try {
    res = await fetch(`${API_BASE}${path}`, {
      method: 'POST',
      headers,
      body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
      signal: opts.signal,
    })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') throw err
    throw new ApiError('Unable to reach the server. Check your connection and try again.', {
      isNetwork: true,
    })
  }

  let envelope: ApiEnvelope<T> | null = null
  try {
    envelope = (await res.json()) as ApiEnvelope<T>
  } catch {
    envelope = null
  }

  // Session expired — try refreshing once, then retry the original request.
  if (res.status === 401 && !opts.skipAuth && attempt === 'first') {
    const newToken = await refreshAccessToken()
    if (newToken) return performEnvelope<T>(path, opts, 'retry')
    sessionExpired()
    throw new ApiError('Your session has expired. Please sign in again.', { status: 401 })
  }

  if (!res.ok || (envelope && typeof envelope.code === 'number' && envelope.code >= 400)) {
    throw ApiError.fromResponse(envelope, res.status)
  }

  return envelope ?? ({ code: res.status, message: 'success' } as ApiEnvelope<T>)
}

/** POST and unwrap the `data` field of the response envelope. */
export async function post<T = unknown>(path: string, body?: unknown, opts?: RequestOptions): Promise<T> {
  const envelope = await performEnvelope<T>(path, { ...opts, body }, 'first')
  return envelope.data as T
}

/** POST a list endpoint and normalize the response into { rows, meta }. */
export async function postList<T>(
  path: string,
  body?: unknown,
  opts?: RequestOptions,
): Promise<ListResult<T>> {
  const envelope = await performEnvelope<unknown>(path, { ...opts, body }, 'first')
  return normalizeList<T>(envelope)
}

export { ApiError }
