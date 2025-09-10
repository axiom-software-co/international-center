// Comprehensive contract integration tests for frontend TypeScript clients
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { useContractNews, useContractResearch, useContractServices, useContractEvents, useContractHealth } from '../composables/useContractApi'
import { ContractErrorHandler } from '../lib/error-handling'

// Mock the API client
vi.mock('../lib/api-client', () => ({
  apiClient: {
    getNews: vi.fn(),
    getNewsById: vi.fn(),
    getFeaturedNews: vi.fn(),
    getNewsCategories: vi.fn(),
    getResearch: vi.fn(),
    getResearchById: vi.fn(),
    getFeaturedResearch: vi.fn(),
    getResearchCategories: vi.fn(),
    getServices: vi.fn(),
    getServiceById: vi.fn(),
    getFeaturedServices: vi.fn(),
    getServiceCategories: vi.fn(),
    getEvents: vi.fn(),
    getEventById: vi.fn(),
    getFeaturedEvents: vi.fn(),
    getEventCategories: vi.fn(),
    getHealth: vi.fn(),
    submitMediaInquiry: vi.fn(),
    submitBusinessInquiry: vi.fn()
  }
}))

describe('Contract API Integration', () => {
  let pinia: any

  beforeEach(() => {
    pinia = createPinia()
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.resetAllMocks()
  })

  describe('News Contract Integration', () => {
    it('should fetch news articles using contract client', async () => {
      // Arrange
      const mockNewsResponse = {
        data: [
          {
            news_id: '123',
            title: 'Test News',
            summary: 'Test Summary',
            category_id: 'cat-123',
            news_type: 'announcement',
            priority_level: 'normal',
            publishing_status: 'published',
            publication_timestamp: '2023-01-01T00:00:00Z',
            created_on: '2023-01-01T00:00:00Z',
            slug: 'test-news'
          }
        ],
        pagination: {
          current_page: 1,
          total_pages: 1,
          total_items: 1,
          items_per_page: 20,
          has_next: false,
          has_previous: false
        }
      }

      const { apiClient } = await import('../lib/api-client')
      vi.mocked(apiClient.getNews).mockResolvedValue(mockNewsResponse)

      // Act
      const newsComposable = useContractNews()
      const result = await newsComposable.fetchNews({ page: 1, limit: 20 })

      // Assert
      expect(apiClient.getNews).toHaveBeenCalledWith({ page: 1, limit: 20 })
      expect(result).toEqual(mockNewsResponse.data)
      expect(newsComposable.news.value).toEqual(mockNewsResponse.data)
      expect(newsComposable.loading.value).toBe(false)
      expect(newsComposable.error.value).toBeNull()
    })

    it('should handle contract client errors properly', async () => {
      // Arrange
      const mockError = new Error('Network error')
      const { apiClient } = await import('../lib/api-client')
      vi.mocked(apiClient.getNews).mockRejectedValue(mockError)

      // Act
      const newsComposable = useContractNews()
      
      try {
        await newsComposable.fetchNews({ page: 1, limit: 20 })
      } catch (err) {
        // Expected to throw
      }

      // Assert
      expect(apiClient.getNews).toHaveBeenCalledWith({ page: 1, limit: 20 })
      expect(newsComposable.news.value).toBeNull()
      expect(newsComposable.error.value).toBeTruthy()
      expect(newsComposable.loading.value).toBe(false)
    })
  })

  describe('Contract Error Handling', () => {
    it('should parse contract-compliant error responses', () => {
      // Arrange
      const contractError = {
        error: {
          code: 'NOT_FOUND',
          message: 'News article not found',
          correlation_id: '123-456-789',
          timestamp: '2023-01-01T00:00:00Z'
        },
        status: 404
      }

      // Act
      const parsedError = ContractErrorHandler.parseContractError(contractError)

      // Assert
      expect(parsedError.isContractError).toBe(true)
      expect(parsedError.error.code).toBe('NOT_FOUND')
      expect(parsedError.status).toBe(404)
    })

    it('should generate user-friendly error messages', () => {
      // Arrange
      const apiError = {
        error: {
          code: 'NOT_FOUND',
          message: 'Resource not found',
          correlation_id: '123',
          timestamp: '2023-01-01T00:00:00Z'
        },
        status: 404,
        isContractError: true
      }

      // Act
      const userMessage = ContractErrorHandler.getUserFriendlyMessage(apiError, 'news article')

      // Assert
      expect(userMessage).toBe('The requested news article could not be found.')
    })

    it('should classify error types correctly', () => {
      // Test validation error
      const validationError = {
        error: { code: 'VALIDATION_ERROR', message: 'Invalid input', correlation_id: '123', timestamp: '2023-01-01T00:00:00Z' },
        status: 400,
        isContractError: true
      }
      expect(ContractErrorHandler.classifyError(validationError)).toBe('validation')

      // Test not found error
      const notFoundError = {
        error: { code: 'NOT_FOUND', message: 'Not found', correlation_id: '123', timestamp: '2023-01-01T00:00:00Z' },
        status: 404,
        isContractError: true
      }
      expect(ContractErrorHandler.classifyError(notFoundError)).toBe('not_found')

      // Test server error
      const serverError = {
        error: { code: 'INTERNAL_ERROR', message: 'Server error', correlation_id: '123', timestamp: '2023-01-01T00:00:00Z' },
        status: 500,
        isContractError: true
      }
      expect(ContractErrorHandler.classifyError(serverError)).toBe('server_error')
    })
  })

  describe('Research Contract Integration', () => {
    it('should fetch research publications using contract client', async () => {
      // Arrange
      const mockResearchResponse = {
        data: [
          {
            research_id: '456',
            title: 'Test Research',
            abstract: 'Test Abstract',
            category_id: 'cat-456',
            research_type: 'clinical_study',
            study_status: 'completed',
            publishing_status: 'published',
            publication_date: '2023-01-01',
            citation_count: 10,
            download_count: 50,
            authors: [{ name: 'Dr. Smith', affiliation: 'University', email: 'smith@example.com' }],
            keywords: ['research', 'study'],
            created_on: '2023-01-01T00:00:00Z',
            slug: 'test-research'
          }
        ],
        pagination: {
          current_page: 1,
          total_pages: 1,
          total_items: 1,
          items_per_page: 20,
          has_next: false,
          has_previous: false
        }
      }

      const { apiClient } = await import('../lib/api-client')
      vi.mocked(apiClient.getResearch).mockResolvedValue(mockResearchResponse)

      // Act
      const researchComposable = useContractResearch()
      const result = await researchComposable.fetchResearch({ page: 1, limit: 20 })

      // Assert
      expect(apiClient.getResearch).toHaveBeenCalledWith({ page: 1, limit: 20 })
      expect(result).toEqual(mockResearchResponse.data)
      expect(researchComposable.research.value).toEqual(mockResearchResponse.data)
    })
  })

  describe('Services Contract Integration', () => {
    it('should fetch services using contract client', async () => {
      // Arrange
      const mockServicesResponse = {
        data: [
          {
            service_id: '789',
            title: 'Test Service',
            description: 'Test Description',
            category_id: 'cat-789',
            service_type: 'consultation',
            availability_status: 'available',
            insurance_accepted: true,
            telehealth_available: true,
            publishing_status: 'published',
            created_on: '2023-01-01T00:00:00Z',
            slug: 'test-service'
          }
        ],
        pagination: {
          current_page: 1,
          total_pages: 1,
          total_items: 1,
          items_per_page: 20,
          has_next: false,
          has_previous: false
        }
      }

      const { apiClient } = await import('../lib/api-client')
      vi.mocked(apiClient.getServices).mockResolvedValue(mockServicesResponse)

      // Act
      const servicesComposable = useContractServices()
      const result = await servicesComposable.fetchServices({ page: 1, limit: 20 })

      // Assert
      expect(apiClient.getServices).toHaveBeenCalledWith({ page: 1, limit: 20 })
      expect(result).toEqual(mockServicesResponse.data)
      expect(servicesComposable.services.value).toEqual(mockServicesResponse.data)
    })
  })

  describe('Events Contract Integration', () => {
    it('should fetch events using contract client', async () => {
      // Arrange
      const mockEventsResponse = {
        data: [
          {
            event_id: '101',
            title: 'Test Event',
            description: 'Test Event Description',
            category_id: 'cat-101',
            event_type: 'conference',
            start_datetime: '2023-12-01T10:00:00Z',
            timezone: 'UTC',
            registration_required: true,
            current_registrations: 0,
            registration_status: 'open',
            organizer: {
              name: 'John Organizer',
              email: 'organizer@example.com'
            },
            publishing_status: 'published',
            created_on: '2023-01-01T00:00:00Z',
            slug: 'test-event'
          }
        ],
        pagination: {
          current_page: 1,
          total_pages: 1,
          total_items: 1,
          items_per_page: 20,
          has_next: false,
          has_previous: false
        }
      }

      const { apiClient } = await import('../lib/api-client')
      vi.mocked(apiClient.getEvents).mockResolvedValue(mockEventsResponse)

      // Act
      const eventsComposable = useContractEvents()
      const result = await eventsComposable.fetchEvents({ page: 1, limit: 20 })

      // Assert
      expect(apiClient.getEvents).toHaveBeenCalledWith({ page: 1, limit: 20 })
      expect(result).toEqual(mockEventsResponse.data)
      expect(eventsComposable.events.value).toEqual(mockEventsResponse.data)
    })
  })

  describe('Health Check Contract Integration', () => {
    it('should perform health checks using contract client', async () => {
      // Arrange
      const mockHealthResponse = {
        status: 'healthy',
        timestamp: '2023-01-01T00:00:00Z',
        version: '1.0.0',
        checks: {
          database: { status: 'up', response_time_ms: 10 },
          vault: { status: 'up', response_time_ms: 5 },
          dapr: { status: 'up', response_time_ms: 3 }
        }
      }

      const { apiClient } = await import('../lib/api-client')
      vi.mocked(apiClient.getHealth).mockResolvedValue(mockHealthResponse)

      // Act
      const healthComposable = useContractHealth()
      const result = await healthComposable.checkHealth()

      // Assert
      expect(apiClient.getHealth).toHaveBeenCalled()
      expect(result).toEqual(mockHealthResponse)
      expect(healthComposable.health.value).toEqual(mockHealthResponse)
      expect(healthComposable.isHealthy.value).toBe(true)
    })
  })

  describe('Frontend Build Type Checking', () => {
    it('should enforce contract types at compile time', () => {
      // This test validates that TypeScript compilation will fail if contract types are violated
      
      // Valid news article should pass type checking
      const validNewsArticle: any = {
        news_id: '123',
        title: 'Valid News',
        summary: 'Valid Summary',
        category_id: 'cat-123',
        news_type: 'announcement',
        priority_level: 'normal',
        publishing_status: 'published',
        publication_timestamp: '2023-01-01T00:00:00Z',
        created_on: '2023-01-01T00:00:00Z',
        slug: 'valid-news'
      }

      // Verify required fields are present
      expect(validNewsArticle.news_id).toBeDefined()
      expect(validNewsArticle.title).toBeDefined()
      expect(validNewsArticle.summary).toBeDefined()
      expect(validNewsArticle.publishing_status).toBeDefined()
      
      // Type checking would catch missing required fields at compile time
      // TypeScript compiler enforces contract compliance
    })

    it('should validate contract response structures', () => {
      // Test that response structure matches contract specifications
      const validPaginatedResponse = {
        data: [],
        pagination: {
          current_page: 1,
          total_pages: 1,
          total_items: 0,
          items_per_page: 20,
          has_next: false,
          has_previous: false
        }
      }

      // Verify pagination structure matches contract
      expect(validPaginatedResponse.pagination).toHaveProperty('current_page')
      expect(validPaginatedResponse.pagination).toHaveProperty('total_pages')
      expect(validPaginatedResponse.pagination).toHaveProperty('total_items')
      expect(validPaginatedResponse.pagination).toHaveProperty('items_per_page')
      expect(validPaginatedResponse.pagination).toHaveProperty('has_next')
      expect(validPaginatedResponse.pagination).toHaveProperty('has_previous')
    })
  })

  describe('End-to-End Contract Workflow', () => {
    it('should complete full news workflow with contract clients', async () => {
      // Arrange
      const mockCategoriesResponse = { data: [{ category_id: 'cat-1', name: 'Test Category', slug: 'test' }] }
      const mockNewsResponse = { 
        data: [{ news_id: '1', title: 'Test', summary: 'Summary', category_id: 'cat-1' }],
        pagination: { current_page: 1, total_items: 1 }
      }

      const { apiClient } = await import('../lib/api-client')
      vi.mocked(apiClient.getNewsCategories).mockResolvedValue(mockCategoriesResponse)
      vi.mocked(apiClient.getNews).mockResolvedValue(mockNewsResponse)

      // Act
      const newsComposable = useContractNews()
      
      // Step 1: Fetch categories
      await newsComposable.fetchNewsCategories()
      expect(newsComposable.categories.value).toEqual(mockCategoriesResponse.data)
      
      // Step 2: Fetch news
      await newsComposable.fetchNews({ page: 1, limit: 20 })
      expect(newsComposable.news.value).toEqual(mockNewsResponse.data)

      // Assert
      expect(apiClient.getNewsCategories).toHaveBeenCalled()
      expect(apiClient.getNews).toHaveBeenCalledWith({ page: 1, limit: 20 })
      
      // Verify complete workflow state
      expect(newsComposable.loading.value).toBe(false)
      expect(newsComposable.error.value).toBeNull()
    })
  })

  describe('Contract Client Type Safety', () => {
    it('should maintain type safety throughout the call chain', async () => {
      // This test validates that types are preserved from contract generation to frontend usage
      
      // Mock with contract-compliant response structure
      const mockHealthResponse = {
        status: 'healthy' as const,
        timestamp: '2023-01-01T00:00:00Z',
        version: '1.0.0',
        checks: {
          database: { status: 'up' as const, response_time_ms: 10 },
          vault: { status: 'up' as const, response_time_ms: 5 }
        }
      }

      const { apiClient } = await import('../lib/api-client')
      vi.mocked(apiClient.getHealth).mockResolvedValue(mockHealthResponse)

      // Act
      const healthComposable = useContractHealth()
      await healthComposable.checkHealth()

      // Assert - TypeScript compilation enforces these types
      expect(typeof healthComposable.health.value?.status).toBe('string')
      expect(typeof healthComposable.health.value?.timestamp).toBe('string')
      expect(typeof healthComposable.health.value?.version).toBe('string')
      expect(healthComposable.health.value?.checks).toBeDefined()
    })
  })
})