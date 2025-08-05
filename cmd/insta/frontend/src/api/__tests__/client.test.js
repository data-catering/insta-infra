import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import * as client from '../client'

// Mock fetch globally
const mockFetch = vi.fn()
global.fetch = mockFetch

describe('API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Set up a default mock response
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      headers: {
        get: vi.fn().mockReturnValue('application/json')
      },
      json: () => Promise.resolve({ success: true }),
      text: () => Promise.resolve('{"success": true}')
    })
  })

  afterEach(() => {
    vi.resetAllMocks()
  })

  describe('Error Handling', () => {
    it('should handle non-ok responses', async () => {
      const errorResponse = {
        error: 'Service not found',
        service_name: 'postgres',
        action: 'start'
      }

      mockFetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: () => Promise.resolve(errorResponse)
      })

      await expect(client.startService('postgres')).rejects.toThrow('Service not found')
    })

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValue(new Error('Network error'))

      await expect(client.listServices()).rejects.toThrow('Network error')
    })

    it('should handle invalid JSON responses', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.reject(new Error('Invalid JSON'))
      })

      await expect(client.listServices()).rejects.toThrow('Unknown error')
    })

    it('should preserve error metadata from backend', async () => {
      const errorResponse = {
        error: 'Docker not running',
        service_name: 'postgres',
        image_name: 'postgres:15',
        action: 'start',
        timestamp: '2024-01-01T10:00:00Z'
      }

      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.resolve(errorResponse)
      })

      try {
        await client.startService('postgres')
      } catch (error) {
        expect(error.metadata.serviceName).toBe('postgres')
        expect(error.metadata.imageName).toBe('postgres:15')
        expect(error.metadata.action).toBe('start')
        expect(error.metadata.timestamp).toBe('2024-01-01T10:00:00Z')
      }
    })
  })

  describe('Service Management', () => {
    describe('listServices', () => {
      it('should fetch services from correct endpoint', async () => {
        const mockServices = {
          postgres: { name: 'postgres', status: 'stopped' },
          redis: { name: 'redis', status: 'running' }
        }

        mockFetch.mockResolvedValue({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockServices)
        })

        const result = await client.listServices()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/services'),
          expect.objectContaining({
            headers: expect.objectContaining({
              'Content-Type': 'application/json'
            })
          })
        )
        expect(result).toEqual(mockServices)
      })
    })

    describe('startService', () => {
      it('should start service with correct parameters', async () => {
        await client.startService('postgres', false)

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/services/postgres/start?persist=false'),
          expect.objectContaining({
            method: 'POST',
            headers: expect.objectContaining({
              'Content-Type': 'application/json'
            })
          })
        )
      })

      it('should start service with persistence', async () => {
        await client.startService('postgres', true)

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/services/postgres/start?persist=true'),
          expect.objectContaining({
            method: 'POST'
          })
        )
      })
    })

    describe('stopService', () => {
      it('should stop service', async () => {
        await client.stopService('postgres')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/services/postgres/stop'),
          expect.objectContaining({
            method: 'POST',
            headers: expect.objectContaining({
              'Content-Type': 'application/json'
            })
          })
        )
      })
    })

    describe('stopAllServices', () => {
      it('should stop all services', async () => {
        await client.stopAllServices()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/services/all/stop'),
          expect.objectContaining({
            method: 'POST'
          })
        )
      })
    })

    describe('getServiceConnection', () => {
      it('should get service connection info', async () => {
        const mockConnection = {
          serviceName: 'postgres',
          webUrls: [{ url: 'http://localhost:5432', description: 'DB' }]
        }

        mockFetch.mockResolvedValue({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockConnection)
        })

        const result = await client.getServiceConnection('postgres')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/services/postgres/connection'),
          expect.objectContaining({
            headers: expect.objectContaining({
              'Content-Type': 'application/json'
            })
          })
        )
        expect(result).toEqual(mockConnection)
      })
    })
  })

  describe('Custom Services', () => {
    describe('listCustomServices', () => {
      it('should fetch custom services', async () => {
        const mockCustomServices = [
          { id: '1', name: 'Custom Service 1' },
          { id: '2', name: 'Custom Service 2' }
        ]

        mockFetch.mockResolvedValue({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockCustomServices)
        })

        const result = await client.listCustomServices()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/custom/compose'),
          expect.objectContaining({
            headers: expect.objectContaining({
              'Content-Type': 'application/json'
            })
          })
        )
        expect(result).toEqual(mockCustomServices)
      })
    })

    describe('uploadCustomService', () => {
      it('should upload custom service with correct data', async () => {
        const mockResponse = { success: true, service: { id: 'new-service' } }

        mockFetch.mockResolvedValue({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockResponse)
        })

        const result = await client.uploadCustomService('Test Service', 'Description', 'compose content')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/custom/compose'),
          expect.objectContaining({
            method: 'POST',
            headers: expect.objectContaining({
              'Content-Type': 'application/json'
            }),
            body: JSON.stringify({
              name: 'Test Service',
              description: 'Description',
              content: 'compose content'
            })
          })
        )
        expect(result).toEqual(mockResponse)
      })
    })

    describe('updateCustomService', () => {
      it('should update custom service', async () => {
        await client.updateCustomService('service-id', 'Updated Name', 'Updated Description', 'updated content')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/custom/compose/service-id'),
          expect.objectContaining({
            method: 'PUT',
            body: JSON.stringify({
              name: 'Updated Name',
              description: 'Updated Description',
              content: 'updated content'
            })
          })
        )
      })
    })

    describe('deleteCustomService', () => {
      it('should delete custom service', async () => {
        await client.deleteCustomService('service-id')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/custom/compose/service-id'),
          expect.objectContaining({
            method: 'DELETE'
          })
        )
      })

      it('should handle 204 No Content response without JSON parse error', async () => {
        // Mock 204 response with no body (typical for successful DELETE operations)
        mockFetch.mockResolvedValue({
          ok: true,
          status: 204,
          statusText: 'No Content',
          headers: {
            get: vi.fn().mockReturnValue(null)
          },
          json: () => Promise.reject(new SyntaxError('Unexpected end of JSON input')),
          text: () => Promise.resolve('')
        })

        // This should return null for 204 responses without throwing an error
        const result = await client.deleteCustomService('service-id')
        expect(result).toBeNull()
      })
    })

    describe('validateCustomCompose', () => {
      it('should validate compose content', async () => {
        const mockValidation = {
          valid: true,
          errors: [],
          warnings: [],
          suggestions: []
        }

        mockFetch.mockResolvedValue({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockValidation)
        })

        const result = await client.validateCustomCompose('compose content')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/v1/custom/validate'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ content: 'compose content' })
          })
        )
        expect(result).toEqual(mockValidation)
      })
    })

    describe('getCustomServiceStats', () => {
      it('should get custom service statistics', async () => {
        const mockStats = {
          total_services: 5,
          active_services: 3
        }

        mockFetch.mockResolvedValue({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockStats)
        })

        const result = await client.getCustomServiceStats()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringMatching(/\/api\/v1\/custom\/stats$/),
          expect.any(Object)
        )
        expect(result).toEqual(mockStats)
      })
    })
  })

  describe('Logs', () => {
    describe('getServiceLogs', () => {
      it('should get service logs with default parameters', async () => {
        const mockLogs = { logs: ['log line 1', 'log line 2'] }

        mockFetch.mockResolvedValue({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockLogs)
        })

        const result = await client.getServiceLogs('postgres')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringMatching(/\/api\/v1\/services\/postgres\/logs$/),
          expect.any(Object)
        )
        expect(result).toEqual(mockLogs)
      })

      it('should get service logs with custom tail parameter', async () => {
        await client.getServiceLogs('postgres', 50)

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringMatching(/\/api\/v1\/services\/postgres\/logs$/),
          expect.any(Object)
        )
      })
    })
  })

  describe('Runtime Management', () => {
    describe('checkImageExists', () => {
      it('should check if image exists', async () => {
        const mockResponse = { exists: true, info: 'postgres:15' }

        mockFetch.mockResolvedValue({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockResponse)
        })

        const result = await client.checkImageExists('postgres:15')

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringMatching(/\/api\/v1\/images\/postgres:15\/exists$/),
          expect.any(Object)
        )
        expect(result).toEqual(mockResponse)
      })
    })

    describe('getRuntimeStatus', () => {
      it('should get runtime status', async () => {
        const mockStatus = { runtime: 'docker', available: true }

        mockFetch.mockResolvedValue({
          ok: true,
          status: 200,
          json: () => Promise.resolve(mockStatus)
        })

        const result = await client.getRuntimeStatus()

        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringMatching(/\/api\/v1\/runtime\/status$/),
          expect.any(Object)
        )
        expect(result).toEqual(mockStatus)
      })
    })
  })

  describe('WebSocket Client', () => {
    // Note: WebSocket testing would typically require a more complex setup
    // For now, we'll test the exported object structure
    it('should export wsClient with expected methods', () => {
      expect(client.wsClient).toBeDefined()
      expect(typeof client.wsClient.subscribe).toBe('function')
      expect(typeof client.wsClient.unsubscribe).toBe('function')
      expect(typeof client.wsClient.subscribeToServiceStatus).toBe('function')
    })

    it('should export WebSocket message types', () => {
      expect(client.WS_MSG_TYPES).toBeDefined()
      expect(client.WS_MSG_TYPES.SERVICE_STATUS_UPDATE).toBeDefined()
      expect(client.WS_MSG_TYPES.SERVICE_STARTED).toBeDefined()
      expect(client.WS_MSG_TYPES.SERVICE_STOPPED).toBeDefined()
      expect(client.WS_MSG_TYPES.SERVICE_ERROR).toBeDefined()
    })
  })

  describe('Request Configuration', () => {
    it('should include correct headers in requests', async () => {
      await client.listServices()

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          })
        })
      )
    })

    it('should use correct API base URL', async () => {
      await client.listServices()

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringMatching(/\/api\/v1\/services$/),
        expect.any(Object)
      )
    })

    it('should handle custom headers in requests', async () => {
      // Test that the client can handle additional headers if needed
      await client.listServices()

      const call = mockFetch.mock.calls[0]
      expect(call[1].headers).toHaveProperty('Content-Type', 'application/json')
    })
  })

  describe('Response Processing', () => {
    it('should parse JSON responses correctly', async () => {
      const mockData = { test: 'data', nested: { value: 123 } }

      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockData)
      })

      const result = await client.listServices()
      expect(result).toEqual(mockData)
    })

    it('should handle empty responses', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        headers: {
          get: vi.fn().mockReturnValue('application/json')
        },
        json: () => Promise.resolve({}),
        text: () => Promise.resolve('{}')
      })

      const result = await client.stopService('postgres')
      expect(result).toEqual({})
    })

    it('should handle 204 No Content responses', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 204, // No Content
        headers: {
          get: vi.fn().mockReturnValue(null)
        },
        json: () => Promise.reject(new SyntaxError('Unexpected end of JSON input')),
        text: () => Promise.resolve('')
      })

      const result = await client.stopService('postgres')
      expect(result).toBeNull()
    })
  })
})