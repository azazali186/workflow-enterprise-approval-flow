const ACCESS_KEY = 'af.access_token'
const REFRESH_KEY = 'af.refresh_token'

export const storage = {
  getAccessToken(): string | null {
    return localStorage.getItem(ACCESS_KEY)
  },
  setAccessToken(token: string): void {
    localStorage.setItem(ACCESS_KEY, token)
  },
  getRefreshToken(): string | null {
    return localStorage.getItem(REFRESH_KEY)
  },
  setRefreshToken(token: string): void {
    localStorage.setItem(REFRESH_KEY, token)
  },
  clearAuth(): void {
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
  },
}
