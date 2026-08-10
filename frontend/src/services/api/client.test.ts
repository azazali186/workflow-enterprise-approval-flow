import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { post } from './client'
import { storage } from '@/utils/storage'

const API_BASE = '/api/v1'

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

describe('api client — proactive session expiry', () => {
  const fetchMock = vi.fn()
  const now = Date.now()

  beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock)
    vi.spyOn(Date, 'now').mockReturnValue(now)
    storage.clearAuth()
    fetchMock.mockReset()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('refreshes before the request when the access token is expired', async () => {
    storage.setAccessToken('expired-token')
    storage.setRefreshToken('refresh-token')
    storage.setExpiresAt(new Date(now - 60_000).toISOString()) // already expired

    fetchMock.mockImplementation((url: string) => {
      if (url === `${API_BASE}/auth/refresh`) {
        return Promise.resolve(
          jsonResponse({
            code: 200,
            message: 'success',
            data: {
              user: { id: 'u1' },
              access_token: 'new-access',
              refresh_token: 'new-refresh',
              expires_at: new Date(now + 3_600_000).toISOString(),
            },
          }),
        )
      }
      if (url === `${API_BASE}/me`) {
        return Promise.resolve(jsonResponse({ code: 200, message: 'success', data: { ok: true } }))
      }
      return Promise.reject(new Error(`unexpected url ${url}`))
    })

    const result = await post('/me')

    expect(result).toEqual({ ok: true })
    // Refresh happened first, then the actual request used the new token.
    expect(fetchMock.mock.calls[0][0]).toBe(`${API_BASE}/auth/refresh`)
    const authHeader = fetchMock.mock.calls[1][1].headers.Authorization
    expect(authHeader).toBe('Bearer new-access')
    // New tokens persisted for the next call.
    expect(storage.getAccessToken()).toBe('new-access')
    expect(storage.getRefreshToken()).toBe('new-refresh')
  })

  it('skips the proactive refresh when the token is not yet expired', async () => {
    storage.setAccessToken('valid-token')
    storage.setRefreshToken('refresh-token')
    storage.setExpiresAt(new Date(now + 3_600_000).toISOString())

    fetchMock.mockImplementation((url: string) => {
      if (url === `${API_BASE}/me`) {
        return Promise.resolve(jsonResponse({ code: 200, message: 'success', data: { ok: true } }))
      }
      return Promise.reject(new Error(`unexpected url ${url}`))
    })

    await post('/me')

    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(fetchMock.mock.calls[0][0]).toBe(`${API_BASE}/me`)
  })

  it('clears the session and throws when the refresh fails', async () => {
    storage.setAccessToken('expired-token')
    storage.setRefreshToken('refresh-token')
    storage.setExpiresAt(new Date(now - 60_000).toISOString())

    fetchMock.mockImplementation((url: string) => {
      if (url === `${API_BASE}/auth/refresh`) {
        return Promise.resolve(jsonResponse({ code: 401, message: 'invalid refresh token' }, 401))
      }
      return Promise.reject(new Error(`unexpected url ${url}`))
    })

    await expect(post('/me')).rejects.toThrow('session has expired')
    expect(storage.getAccessToken()).toBeNull()
    expect(storage.getRefreshToken()).toBeNull()
  })
})
