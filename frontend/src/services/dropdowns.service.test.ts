import { describe, it, expect, vi, beforeEach } from 'vitest'
import { dropdownsService } from './dropdowns.service'

// Mock the API client
vi.mock('./api/client', () => ({
  post: vi.fn(),
}))

import { post } from './api/client'
const mockPost = vi.mocked(post)

describe('DropdownsService', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('sends correct request payload', async () => {
      mockPost.mockResolvedValue({ users: [] })

      await dropdownsService.list({ entities: ['users'] })

      expect(mockPost).toHaveBeenCalledWith('/dropdowns', { entities: ['users'] })
    })

    it('includes optional parameters when provided', async () => {
      mockPost.mockResolvedValue({ workflows: [] })

      await dropdownsService.list({
        entities: ['workflows'],
        include_inactive: true,
        statuses: ['submitted', 'approved'],
      })

      expect(mockPost).toHaveBeenCalledWith('/dropdowns', {
        entities: ['workflows'],
        include_inactive: true,
        statuses: ['submitted', 'approved'],
      })
    })

    it('handles empty response', async () => {
      mockPost.mockResolvedValue({})

      const result = await dropdownsService.list({ entities: ['users'] })

      expect(result).toEqual({})
    })

    it('handles network failure', async () => {
      mockPost.mockRejectedValue(new Error('Network error'))

      await expect(
        dropdownsService.list({ entities: ['users'] })
      ).rejects.toThrow('Network error')
    })

    it('handles API error response', async () => {
      const error = new Error('Invalid entity type')
      mockPost.mockRejectedValue(error)

      await expect(
        dropdownsService.list({ entities: ['invalid'] })
      ).rejects.toThrow('Invalid entity type')
    })

    it('handles timeout error', async () => {
      const error = new Error('Request timeout')
      mockPost.mockRejectedValue(error)

      await expect(
        dropdownsService.list({ entities: ['users'] })
      ).rejects.toThrow('Request timeout')
    })
  })

  describe('convenience methods', () => {
    it('users() returns user options', async () => {
      mockPost.mockResolvedValue({
        users: [{ id: '1', name: 'Alice' }],
      })

      const result = await dropdownsService.users()

      expect(result).toEqual([{ id: '1', name: 'Alice' }])
      expect(mockPost).toHaveBeenCalledWith('/dropdowns', { entities: ['users'] })
    })

    it('users() returns empty array when no users', async () => {
      mockPost.mockResolvedValue({})

      const result = await dropdownsService.users()

      expect(result).toEqual([])
    })

    it('workflows() handles include_inactive flag', async () => {
      mockPost.mockResolvedValue({
        workflows: [{ id: '1', name: 'Test' }],
      })

      await dropdownsService.workflows(true)

      expect(mockPost).toHaveBeenCalledWith('/dropdowns', {
        entities: ['workflows'],
        include_inactive: true,
      })
    })

    it('workflows() defaults to active only', async () => {
      mockPost.mockResolvedValue({ workflows: [] })

      await dropdownsService.workflows()

      expect(mockPost).toHaveBeenCalledWith('/dropdowns', {
        entities: ['workflows'],
        include_inactive: false,
      })
    })

    it('applications() uses default status', async () => {
      mockPost.mockResolvedValue({ applications: [] })

      await dropdownsService.applications()

      expect(mockPost).toHaveBeenCalledWith('/dropdowns', {
        entities: ['applications'],
        statuses: ['submitted'],
      })
    })

    it('applications() uses custom statuses', async () => {
      mockPost.mockResolvedValue({ applications: [] })

      await dropdownsService.applications(['approved', 'completed'])

      expect(mockPost).toHaveBeenCalledWith('/dropdowns', {
        entities: ['applications'],
        statuses: ['approved', 'completed'],
      })
    })

    it('handles null response fields gracefully', async () => {
      mockPost.mockResolvedValue({
        users: null,
        workflows: undefined,
      })

      const users = await dropdownsService.users()
      expect(users).toEqual([])

      const workflows = await dropdownsService.workflows()
      expect(workflows).toEqual([])
    })
  })

  describe('multiple', () => {
    it('fetches multiple entity types in one request', async () => {
      mockPost.mockResolvedValue({
        users: [{ id: '1', name: 'Alice' }],
        workflows: [{ id: '2', name: 'Test' }],
      })

      const result = await dropdownsService.multiple(['users', 'workflows'])

      expect(result).toEqual({
        users: [{ id: '1', name: 'Alice' }],
        workflows: [{ id: '2', name: 'Test' }],
      })
      expect(mockPost).toHaveBeenCalledTimes(1)
    })

    it('passes options correctly', async () => {
      mockPost.mockResolvedValue({})

      await dropdownsService.multiple(['workflows', 'templates'], {
        include_inactive: true,
        statuses: ['submitted'],
      })

      expect(mockPost).toHaveBeenCalledWith('/dropdowns', {
        entities: ['workflows', 'templates'],
        include_inactive: true,
        statuses: ['submitted'],
      })
    })

    it('handles partial response', async () => {
      mockPost.mockResolvedValue({
        users: [{ id: '1', name: 'Alice' }],
        // workflows missing
      })

      const result = await dropdownsService.multiple(['users', 'workflows'])

      expect(result.users).toEqual([{ id: '1', name: 'Alice' }])
      expect(result.workflows).toBeUndefined()
    })
  })

  describe('error handling', () => {
    it('propagates AbortError for cancelled requests', async () => {
      const abortError = new DOMException('Aborted', 'AbortError')
      mockPost.mockRejectedValue(abortError)

      await expect(
        dropdownsService.list({ entities: ['users'] })
      ).rejects.toThrow('Aborted')
    })

    it('handles JSON parse error in response', async () => {
      mockPost.mockRejectedValue(new SyntaxError('Unexpected token'))

      await expect(
        dropdownsService.list({ entities: ['users'] })
      ).rejects.toThrow('Unexpected token')
    })

    it('handles server 500 error', async () => {
      mockPost.mockRejectedValue(new Error('Internal Server Error'))

      await expect(
        dropdownsService.list({ entities: ['users'] })
      ).rejects.toThrow('Internal Server Error')
    })

    it('handles server 400 error', async () => {
      mockPost.mockRejectedValue(new Error('Bad Request'))

      await expect(
        dropdownsService.list({ entities: ['invalid'] })
      ).rejects.toThrow('Bad Request')
    })

    it('handles server 401 unauthorized', async () => {
      mockPost.mockRejectedValue(new Error('Unauthorized'))

      await expect(
        dropdownsService.list({ entities: ['users'] })
      ).rejects.toThrow('Unauthorized')
    })

    it('handles server 403 forbidden', async () => {
      mockPost.mockRejectedValue(new Error('Forbidden'))

      await expect(
        dropdownsService.list({ entities: ['users'] })
      ).rejects.toThrow('Forbidden')
    })
  })

  describe('edge cases', () => {
    it('handles empty entities array', async () => {
      mockPost.mockResolvedValue({})

      const result = await dropdownsService.list({ entities: [] })

      expect(result).toEqual({})
      expect(mockPost).toHaveBeenCalledWith('/dropdowns', { entities: [] })
    })

    it('handles very large entity list', async () => {
      const entities = Array.from({ length: 100 }, (_, i) => `entity_${i}`)
      mockPost.mockResolvedValue({})

      await dropdownsService.list({ entities })

      expect(mockPost).toHaveBeenCalledWith('/dropdowns', { entities })
    })

    it('handles special characters in entity names', async () => {
      mockPost.mockResolvedValue({
        users: [{ id: '1', name: 'O\'Brien & Sons' }],
      })

      const result = await dropdownsService.users()

      expect(result[0].name).toBe("O'Brien & Sons")
    })

    it('handles unicode in response', async () => {
      mockPost.mockResolvedValue({
        users: [{ id: '1', name: '日本語テスト' }],
      })

      const result = await dropdownsService.users()

      expect(result[0].name).toBe('日本語テスト')
    })

    it('handles concurrent calls', async () => {
      mockPost.mockResolvedValue({ users: [{ id: '1', name: 'Alice' }] })

      const promises = Array.from({ length: 10 }, () => dropdownsService.users())
      const results = await Promise.all(promises)

      expect(results).toHaveLength(10)
      results.forEach((result) => {
        expect(result).toEqual([{ id: '1', name: 'Alice' }])
      })
    })
  })
})
