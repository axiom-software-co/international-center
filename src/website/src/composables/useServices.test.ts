import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, nextTick } from 'vue';
import { useServices, useService, useFeaturedServices, useServiceCategories } from './useServices';
import { servicesClient } from '../lib/clients';
import { RestError } from '../lib/clients/rest/BaseRestClient';

// Mock the services client
vi.mock('../lib/clients', () => ({
  servicesClient: {
    getServices: vi.fn(),
    getServiceBySlug: vi.fn(),
    getFeaturedServices: vi.fn(),
    getServiceCategories: vi.fn(),
  }
}));

describe('useServices composables', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('useServices', () => {
    it('should initialize with correct default values', () => {
      const { services, loading, error, total, page, pageSize, totalPages } = useServices({ enabled: false });

      expect(services.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      expect(totalPages.value).toBe(0);
    });

    it('should fetch services with backend response format including content', async () => {
      const mockBackendResponse = {
        services: [
          {
            service_id: '123',
            title: 'Cardiology Services',
            description: 'Heart care services',
            slug: 'cardiology',
            publishing_status: 'published',
            category_id: '456',
            delivery_mode: 'outpatient_service',
            content: '<h2>Comprehensive Heart Care</h2><p>Our cardiology team provides advanced diagnostics and treatment.</p>',
            image_url: 'https://storage.azure.com/images/cardiology.jpg',
            order_number: 1
          }
        ],
        count: 1,
        correlation_id: 'services-correlation-123'
      };

      (servicesClient.getServices as any).mockResolvedValueOnce(mockBackendResponse);

      const { services, loading, error, total, refetch } = useServices({ 
        page: 1, 
        pageSize: 10,
        immediate: false 
      });

      expect(loading.value).toBe(false);
      
      await refetch();
      await nextTick();

      expect(services.value).toEqual(mockBackendResponse.services);
      expect(total.value).toBe(1);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(servicesClient.getServices).toHaveBeenCalledWith({
        page: 1,
        pageSize: 10
      });
    });

    it('should handle search parameter correctly', async () => {
      const mockSearchResponse = {
        services: [
          {
            service_id: '789',
            title: 'Cardiac Surgery',
            description: 'Advanced heart procedures',
            slug: 'cardiac-surgery',
            publishing_status: 'published'
          }
        ],
        count: 1,
        correlation_id: 'search-correlation-789'
      };

      (servicesClient.getServices as any).mockResolvedValueOnce(mockSearchResponse);

      const { services, refetch } = useServices({ 
        search: 'cardiac',
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(servicesClient.getServices).toHaveBeenCalledWith({
        search: 'cardiac'
      });
      expect(services.value).toEqual(mockSearchResponse.services);
    });

    it('should handle category filtering', async () => {
      const mockCategoryResponse = {
        services: [
          {
            service_id: '456',
            title: 'Primary Care Checkup',
            category_id: 'primary-care-id'
          }
        ],
        count: 1,
        correlation_id: 'category-correlation-456'
      };

      (servicesClient.getServices as any).mockResolvedValueOnce(mockCategoryResponse);

      const { services, refetch } = useServices({ 
        category: 'primary-care',
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(servicesClient.getServices).toHaveBeenCalledWith({
        category: 'primary-care'
      });
    });

    it('should handle API errors with correlation_id', async () => {
      const mockError = new RestError(
        'Services not found',
        404,
        { error: { code: 'NOT_FOUND', message: 'Services not found' } },
        'error-correlation-404'
      );

      (servicesClient.getServices as any).mockRejectedValueOnce(mockError);

      const { services, loading, error, refetch } = useServices({ immediate: false });

      await refetch();
      await nextTick();

      expect(services.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe('Services not found');
    });

    it('should handle rate limiting errors', async () => {
      const mockRateLimitError = new RestError(
        'Rate limit exceeded: Too many requests',
        429,
        { error: { code: 'RATE_LIMIT_EXCEEDED', message: 'Too many requests' } },
        'rate-limit-correlation-429'
      );

      (servicesClient.getServices as any).mockRejectedValueOnce(mockRateLimitError);

      const { error, refetch } = useServices({ immediate: false });

      await refetch();
      await nextTick();

      expect(error.value).toBe('Rate limit exceeded: Too many requests');
    });
  });

  describe('useService', () => {
    it('should fetch service by slug with backend format', async () => {
      const mockServiceResponse = {
        service: {
          service_id: '123',
          title: 'Cardiology Services',
          description: 'Comprehensive heart care',
          slug: 'cardiology',
          publishing_status: 'published',
          category_id: '456',
          delivery_mode: 'outpatient_service',
          content: '<h2>Comprehensive Heart Care Services</h2><p>Our cardiology team provides advanced diagnostics and treatment for heart conditions including:</p><ul><li>ECG and stress testing</li><li>Echocardiogram imaging</li><li>Cardiac catheterization</li><li>Heart surgery coordination</li></ul><p>Contact us to schedule your consultation.</p>'
        },
        correlation_id: 'service-correlation-123'
      };

      (servicesClient.getServiceBySlug as any).mockResolvedValueOnce(mockServiceResponse);

      const { service, loading, error, refetch } = useService(ref('cardiology'));

      // Wait for initial fetch
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(service.value).toEqual(mockServiceResponse.service);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(servicesClient.getServiceBySlug).toHaveBeenCalledWith('cardiology');
    });

    it('should handle null slug gracefully', async () => {
      const { service, loading } = useService(ref(null));

      await nextTick();

      expect(service.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(servicesClient.getServiceBySlug).not.toHaveBeenCalled();
    });

    it('should react to slug changes', async () => {
      const mockService1 = {
        service: { service_id: '1', title: 'Service 1', slug: 'service-1' },
        correlation_id: 'correlation-1'
      };
      const mockService2 = {
        service: { service_id: '2', title: 'Service 2', slug: 'service-2' },
        correlation_id: 'correlation-2'
      };

      (servicesClient.getServiceBySlug as any)
        .mockResolvedValueOnce(mockService1)
        .mockResolvedValueOnce(mockService2);

      const slugRef = ref('service-1');
      const { service } = useService(slugRef);

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(service.value).toEqual(mockService1.service);

      // Change slug
      slugRef.value = 'service-2';
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(service.value).toEqual(mockService2.service);
      expect(servicesClient.getServiceBySlug).toHaveBeenCalledTimes(2);
    });
  });

  describe('useFeaturedServices', () => {
    it('should fetch published services for featured display', async () => {
      const mockFeaturedResponse = {
        services: [
          {
            service_id: '789',
            title: 'Featured Cardiology',
            publishing_status: 'published',
            featured: true
          },
          {
            service_id: '101',
            title: 'Featured Orthopedics',
            publishing_status: 'published',
            featured: true
          }
        ],
        count: 2,
        correlation_id: 'featured-correlation-789'
      };

      (servicesClient.getFeaturedServices as any).mockResolvedValueOnce(mockFeaturedResponse);

      const { services, loading, error } = useFeaturedServices();

      // Wait for mount and fetch
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(services.value).toEqual(mockFeaturedResponse.services);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(servicesClient.getFeaturedServices).toHaveBeenCalledWith(undefined);
    });

    it('should handle limit parameter', async () => {
      const mockLimitedResponse = {
        services: [
          { service_id: '1', title: 'Service 1', publishing_status: 'published' }
        ],
        count: 1,
        correlation_id: 'limited-correlation-123'
      };

      (servicesClient.getFeaturedServices as any).mockResolvedValueOnce(mockLimitedResponse);

      const { services } = useFeaturedServices(5);

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(servicesClient.getFeaturedServices).toHaveBeenCalledWith(5);
    });
  });

  describe('useServiceCategories', () => {
    it('should fetch categories with backend format', async () => {
      const mockCategoriesResponse = {
        categories: [
          {
            category_id: '456',
            name: 'Primary Care',
            slug: 'primary-care',
            order_number: 1,
            is_default_unassigned: false
          },
          {
            category_id: '789',
            name: 'Specialty Care',
            slug: 'specialty-care',
            order_number: 2,
            is_default_unassigned: false
          }
        ],
        count: 2,
        correlation_id: 'categories-correlation-456'
      };

      (servicesClient.getServiceCategories as any).mockResolvedValueOnce(mockCategoriesResponse);

      const { categories, loading, error } = useServiceCategories();

      // Wait for mount and fetch
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(categories.value).toEqual(mockCategoriesResponse.categories);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(servicesClient.getServiceCategories).toHaveBeenCalled();
    });

    it('should handle empty categories response', async () => {
      const mockEmptyResponse = {
        categories: [],
        count: 0,
        correlation_id: 'empty-categories-correlation'
      };

      (servicesClient.getServiceCategories as any).mockResolvedValueOnce(mockEmptyResponse);

      const { categories } = useServiceCategories();

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(categories.value).toEqual([]);
    });
  });

  describe('error handling across composables', () => {
    it('should handle network errors consistently', async () => {
      const networkError = new Error('Network connection failed');
      
      (servicesClient.getServices as any).mockRejectedValueOnce(networkError);

      const { error, refetch } = useServices({ immediate: false });

      await refetch();
      await nextTick();

      expect(error.value).toBe('Network connection failed');
    });

    it('should handle timeout errors', async () => {
      const timeoutError = new RestError('Request timeout after 5000ms', 408);
      
      (servicesClient.getServiceBySlug as any).mockRejectedValueOnce(timeoutError);

      const { error } = useService(ref('test-service'));

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(error.value).toBe('Request timeout after 5000ms');
    });

    it('should reset error state on successful refetch', async () => {
      const mockError = new RestError('Temporary error', 500);
      const mockSuccessResponse = {
        services: [{ service_id: '1', title: 'Test Service' }],
        count: 1,
        correlation_id: 'success-correlation'
      };

      (servicesClient.getServices as any)
        .mockRejectedValueOnce(mockError)
        .mockResolvedValueOnce(mockSuccessResponse);

      const { error, refetch } = useServices({ immediate: false });

      // First call fails
      await refetch();
      await nextTick();
      expect(error.value).toBe('Temporary error');

      // Second call succeeds
      await refetch();
      await nextTick();
      expect(error.value).toBe(null);
    });
  });
});