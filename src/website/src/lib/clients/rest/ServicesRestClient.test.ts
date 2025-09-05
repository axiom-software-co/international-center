import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ServicesRestClient } from './ServicesRestClient';
import { RestError } from './BaseRestClient';

// Mock function defined at module level for proper hoisting
const createMockFetchResponse = (data: any, ok = true, status = 200) => {
  const response = {
    ok,
    status,
    statusText: ok ? 'OK' : 'Error',
    json: vi.fn().mockResolvedValue(data),
    text: vi.fn().mockResolvedValue(JSON.stringify(data)),
    headers: new Map([['content-type', 'application/json']]),
    clone: vi.fn().mockReturnThis()
  };
  
  return Promise.resolve(response as Response);
};

// Mock fetch with proper vitest mock setup
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('ServicesRestClient', () => {
  let client: ServicesRestClient;
  
  beforeEach(() => {
    client = new ServicesRestClient();
    mockFetch.mockClear();
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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(expectedResponse)
      );

      const result = await client.getServices();
      
      expect(result).toEqual(expectedResponse);
    });

    it('should handle search parameter correctly', async () => {
      const mockResponse = {
        services: [],
        count: 0,
        correlation_id: 'search-correlation-456'
      };

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(expectedResponse)
      );

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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(expectedResponse)
      );

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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(errorResponse, false, 404)
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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(errorResponse, false, 400)
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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(errorResponse, false, 429)
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

      mockFetch.mockResolvedValueOnce(
        createMockFetchResponse(errorResponse, false, 500)
      );

      await expect(client.getServices()).rejects.toThrow(RestError);
      await expect(client.getServices()).rejects.toThrow('Server error');
    });
  });

  describe('timeout and retry behavior', () => {
    it('should timeout after configured duration', async () => {
      // Mock fetch to never resolve
      mockFetch.mockImplementation(() => 
        new Promise(() => {}) // Never resolves
      );

      const start = Date.now();
      await expect(client.getServices()).rejects.toThrow('Request timeout');
      const elapsed = Date.now() - start;
      
      // Should timeout around the configured timeout (5000ms)
      expect(elapsed).toBeGreaterThan(4900);
      expect(elapsed).toBeLessThan(5200);
    }, 15000); // Set test timeout to 15 seconds

    it('should retry on 500 errors', async () => {
      const errorResponse = {
        error: {
          code: 'INTERNAL_ERROR',
          message: 'Temporary server error',
          correlation_id: 'retry-correlation-500'
        }
      };

      // First call fails, second succeeds (client retries once for 500 errors)
      (global.fetch as any)
        .mockResolvedValueOnce(createMockFetchResponse(errorResponse, false, 500))
        .mockResolvedValueOnce(createMockFetchResponse({
          services: [],
          count: 0,
          correlation_id: 'success-after-retry'
        }));

      const result = await client.getServices();
      
      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result.correlation_id).toBe('success-after-retry');
    });
  });
});