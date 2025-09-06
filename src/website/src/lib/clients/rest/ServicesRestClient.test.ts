import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ServicesRestClient } from './ServicesRestClient';
import { RestError } from './BaseRestClient';
import { mockFetch } from '../../../test/setup';

// Simple mock response helper for this test file
const createMockResponse = (data: any, status = 200, ok = true) => {
  const statusText = status === 200 ? 'OK' :
                     status === 400 ? 'Bad Request' :
                     status === 404 ? 'Not Found' :
                     status === 429 ? 'Too Many Requests' :
                     status === 500 ? 'Internal Server Error' : 'Unknown';
  
  return {
    ok,
    status,
    statusText,
    headers: { get: () => 'application/json' },
    json: () => Promise.resolve(data)
  };
};

describe('ServicesRestClient', () => {
  let client: ServicesRestClient;
  
  beforeEach(() => {
    client = new ServicesRestClient();
    
    // Ensure completely clean mock state for each test
    mockFetch.mockReset();
    mockFetch.mockClear();
    
    // Clear cache for complete test isolation
    client.clearCache();
  });

  describe('getServices', () => {
    it('should call correct backend endpoint with v1 API path', async () => {
      const mockResponse = {
        services: [
          {
            service_id: '123',
            title: 'Test Service',
            description: 'Test Description',
            slug: 'test-service',
            publishing_status: 'published'
          }
        ],
        count: 1,
        correlation_id: 'test-correlation-123'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getServices({ page: 1, pageSize: 10 });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/services?page=1&pageSize=10',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          })
        })
      );
    });

    it('should return backend response format with services array and correlation_id', async () => {
      const expectedResponse = {
        services: [
          {
            service_id: '123',
            title: 'Test Service',
            description: 'Test Description',
            slug: 'test-service',
            publishing_status: 'published'
          }
        ],
        count: 1,
        correlation_id: 'test-correlation-123'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(expectedResponse));

      const result = await client.getServices();
      
      expect(result).toEqual(expectedResponse);
    });

    it('should handle search parameter correctly', async () => {
      const mockResponse = {
        services: [],
        count: 0,
        correlation_id: 'search-correlation-456'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getServices({ search: 'cardiac' });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/services?search=cardiac',
        expect.any(Object)
      );
    });

    it('should handle category filter correctly', async () => {
      const mockResponse = {
        services: [],
        count: 0,
        correlation_id: 'category-correlation-789'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getServices({ category: 'primary-care' });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/services/categories/primary-care/services',
        expect.any(Object)
      );
    });
  });

  describe('getServiceBySlug', () => {
    it('should call correct v1 API endpoint for service by slug', async () => {
      const mockResponse = {
        service: {
          service_id: '123',
          title: 'Cardiology Services',
          slug: 'cardiology',
          description: 'Heart care services',
          publishing_status: 'published',
          category_id: '456',
          delivery_mode: 'outpatient_service',
          content: '<p>Comprehensive cardiac care including diagnostics, treatment planning, and ongoing patient management.</p>',
          image_url: 'https://storage.azure.com/images/cardiology-hero.jpg',
          order_number: 1
        },
        correlation_id: 'slug-correlation-abc'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getServiceBySlug('cardiology');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/services/slug/cardiology',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });

    it('should return service with content field from PostgreSQL storage', async () => {
      const expectedResponse = {
        service: {
          service_id: '123',
          title: 'Cardiology Services',
          slug: 'cardiology',
          description: 'Heart care services',
          publishing_status: 'published',
          category_id: '456',
          delivery_mode: 'outpatient_service',
          content: '<h2>Advanced Heart Care</h2>\n<p>Our cardiology department provides comprehensive heart care services including:</p>\n<ul>\n<li>Diagnostic testing (ECG, echo, stress testing)</li>\n<li>Interventional procedures</li>\n<li>Cardiac rehabilitation</li>\n<li>Preventive care and education</li>\n</ul>',
          image_url: 'https://storage.azure.com/images/cardiology-hero.jpg',
          order_number: 1
        },
        correlation_id: 'slug-correlation-abc'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(expectedResponse));

      const result = await client.getServiceBySlug('cardiology');
      
      expect(result).toEqual(expectedResponse);
      expect(result.service.content).toContain('Advanced Heart Care');
      expect(result.service.content).toContain('<ul>');
    });

    it('should handle services without content field', async () => {
      const expectedResponse = {
        service: {
          service_id: '124',
          title: 'Basic Consultation',
          slug: 'consultation',
          description: 'General consultation services',
          publishing_status: 'published',
          category_id: '456',
          delivery_mode: 'outpatient_service',
          // content field is null/undefined
          image_url: 'https://storage.azure.com/images/consultation.jpg',
          order_number: 2
        },
        correlation_id: 'consultation-correlation-def'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(expectedResponse));

      const result = await client.getServiceBySlug('consultation');
      
      expect(result).toEqual(expectedResponse);
      expect(result.service.content).toBeUndefined();
    });

    it('should throw validation error for empty slug', async () => {
      await expect(client.getServiceBySlug('')).rejects.toThrow('Service slug is required');
    });
  });

  describe('getServiceCategories', () => {
    it('should call correct v1 API endpoint for categories', async () => {
      const mockResponse = {
        categories: [
          {
            category_id: '456',
            name: 'Primary Care',
            slug: 'primary-care',
            order_number: 1
          }
        ],
        count: 1,
        correlation_id: 'categories-correlation-def'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getServiceCategories();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/services/categories',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });
  });

  describe('getFeaturedServices', () => {
    it('should call correct v1 API endpoint for published services', async () => {
      const mockResponse = {
        services: [
          {
            service_id: '789',
            title: 'Featured Service',
            publishing_status: 'published'
          }
        ],
        count: 1,
        correlation_id: 'published-correlation-ghi'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getFeaturedServices();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/services/published',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });
  });

  describe('getFeaturedCategories', () => {
    it('should call correct v1 API endpoint for featured categories', async () => {
      const mockResponse = {
        featured_categories: [
          {
            featured_category_id: 'featured-1',
            category_id: '456',
            feature_position: 1
          }
        ],
        count: 1,
        correlation_id: 'featured-correlation-jkl'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getFeaturedCategories();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/services/featured',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });
  });

  describe('error handling', () => {
    it('should handle 404 errors with correlation_id', async () => {
      const errorResponse = {
        error: {
          code: 'NOT_FOUND',
          message: 'Service not found',
          correlation_id: 'error-correlation-404'
        }
      };

      mockFetch.mockImplementation(() => 
        createMockResponse(errorResponse, 404, false)
      );

      await expect(client.getServiceBySlug('nonexistent')).rejects.toThrow(RestError);
      await expect(client.getServiceBySlug('nonexistent')).rejects.toThrow('Not found');
    });

    it('should handle 400 validation errors', async () => {
      const errorResponse = {
        error: {
          code: 'VALIDATION_ERROR',
          message: 'Invalid service parameters',
          correlation_id: 'error-correlation-400'
        }
      };

      mockFetch.mockImplementation(() =>
        createMockResponse(errorResponse, 400, false)
      );

      await expect(client.getServices({ page: -1 })).rejects.toThrow(RestError);
      await expect(client.getServices({ page: -1 })).rejects.toThrow('Bad request');
    });

    it('should handle 429 rate limit errors', async () => {
      const errorResponse = {
        error: {
          code: 'RATE_LIMIT_EXCEEDED',
          message: 'Too many requests',
          correlation_id: 'error-correlation-429'
        }
      };

      mockFetch.mockImplementation(() =>
        createMockResponse(errorResponse, 429, false)
      );

      await expect(client.getServices()).rejects.toThrow(RestError);
      await expect(client.getServices()).rejects.toThrow('Rate limit exceeded');
    });

    it('should handle 500 server errors', async () => {
      const errorResponse = {
        error: {
          code: 'INTERNAL_ERROR',
          message: 'Database connection failed',
          correlation_id: 'error-correlation-500'
        }
      };

      mockFetch.mockImplementation(() =>
        createMockResponse(errorResponse, 500, false)
      );

      await expect(client.getServices()).rejects.toThrow(RestError);
      await expect(client.getServices()).rejects.toThrow('Server error');
    });
  });

  describe('timeout and retry behavior', () => {
    it('should timeout after configured duration', async () => {
      // Mock fetch to reject with timeout error
      const timeoutError = new Error('Request timeout');
      timeoutError.name = 'AbortError';
      mockFetch.mockRejectedValue(timeoutError);

      const start = Date.now();
      await expect(client.getServices()).rejects.toThrow('Request timeout');
      const elapsed = Date.now() - start;
      
      // Should timeout quickly since we're mocking the timeout
      expect(elapsed).toBeLessThan(100);
    }, 5000);

    it('should retry on 500 errors', async () => {
      const errorResponse = {
        error: {
          code: 'INTERNAL_ERROR',
          message: 'Temporary server error',
          correlation_id: 'retry-correlation-500'
        }
      };

      // First call fails, second succeeds (client retries once for 500 errors)
      mockFetch
        .mockResolvedValueOnce(createMockResponse(errorResponse, 500, false))
        .mockResolvedValueOnce(createMockResponse({
          services: [],
          count: 0,
          correlation_id: 'success-after-retry'
        }));

      const result = await client.getServices();
      
      expect(mockFetch).toHaveBeenCalledTimes(2);
      expect(result.correlation_id).toBe('success-after-retry');
    });
  });

  describe('Shared Cache Behavior', () => {
    it('should use shared RestClientCache for caching operations', async () => {
      const mockServicesResponse: BackendServicesResponse = {
        services: [{
          service_id: 'cache-test-uuid',
          title: 'Cache Test Service',
          summary: 'Testing cache behavior',
          slug: 'cache-test-service',
          category_id: 'category-uuid',
          publishing_status: 'published',
          tags: ['cache', 'test'],
          service_type: 'consultation',
          priority_level: 'normal',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false,
          deleted_on: null,
          deleted_by: null,
        }],
        count: 1,
        correlation_id: 'cache-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockServicesResponse));

      // Clear cache before test
      client.clearCache();

      // First request should hit the API
      const firstResult = await client.getServices();
      expect(mockFetch).toHaveBeenCalledTimes(1);
      expect(firstResult).toEqual(mockServicesResponse);

      // Second request should use cache (no additional fetch call)
      const secondResult = await client.getServices();
      expect(mockFetch).toHaveBeenCalledTimes(1); // Still 1, not 2
      expect(secondResult).toEqual(mockServicesResponse);
    }, 5000);

    it('should provide cache performance metrics via shared cache', async () => {
      // Clear cache and reset metrics
      client.clearCache();

      // Initial metrics should show empty state
      const initialMetrics = client.getMetrics();
      expect(initialMetrics.totalRequests).toBe(0);
      expect(initialMetrics.cacheHits).toBe(0);
      expect(initialMetrics.cacheMisses).toBe(0);
      expect(initialMetrics.errorCount).toBe(0);
    }, 5000);

    it('should provide cache statistics via shared cache', async () => {
      // Clear cache before test
      client.clearCache();

      const initialStats = client.getCacheStats();
      expect(initialStats).toHaveProperty('size');
      expect(initialStats).toHaveProperty('hitRate');
      expect(typeof initialStats.size).toBe('number');
      expect(typeof initialStats.hitRate).toBe('number');
    }, 5000);

    it('should clear all cache entries and reset metrics', async () => {
      // Clear cache before test
      client.clearCache();

      // Verify cache is cleared
      const stats = client.getCacheStats();
      expect(stats.size).toBe(0);

      // Verify metrics are reset
      const metrics = client.getMetrics();
      expect(metrics.totalRequests).toBe(0);
      expect(metrics.cacheHits).toBe(0);
      expect(metrics.cacheMisses).toBe(0);
      expect(metrics.errorCount).toBe(0);
    }, 5000);
  });
});