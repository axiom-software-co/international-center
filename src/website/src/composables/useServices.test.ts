import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, nextTick, defineComponent, computed } from 'vue';
import { mount } from '@vue/test-utils';
import { useServices, useService, useFeaturedServices, useServiceCategories } from './useServices';
import { useServicesStore } from '../stores/services';
import { RestError } from '../lib/clients/rest/BaseRestClient';

// Mock the services client with hoisted functions
const {
  mockGetServices,
  mockGetServiceBySlug,
  mockGetFeaturedServices,
  mockGetServiceCategories
} = vi.hoisted(() => {
  const mockGetServicesFunc = vi.fn();
  const mockGetServiceBySlugFunc = vi.fn();
  const mockGetFeaturedServicesFunc = vi.fn();
  const mockGetServiceCategoriesFunc = vi.fn();
  
  return {
    mockGetServices: mockGetServicesFunc,
    mockGetServiceBySlug: mockGetServiceBySlugFunc,
    mockGetFeaturedServices: mockGetFeaturedServicesFunc,
    mockGetServiceCategories: mockGetServiceCategoriesFunc,
  };
});

vi.mock('../lib/clients', () => ({
  servicesClient: {
    getServices: mockGetServices,
    getServiceBySlug: mockGetServiceBySlug,
    getFeaturedServices: mockGetFeaturedServices,
    getServiceCategories: mockGetServiceCategories,
  },
  // Pass through any types that might be imported
  ...vi.importActual('../lib/clients')
}));


describe('useServices composables', () => {
  beforeEach(() => {
    // Clear all client mocks
    mockGetServices.mockClear();
    mockGetServiceBySlug.mockClear();
    mockGetFeaturedServices.mockClear();
    mockGetServiceCategories.mockClear();
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
      const mockServices = [
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
      ];

      // Mock client response
      mockGetServices.mockResolvedValue({
        services: mockServices,
        count: 1
      });

      // Direct composable call
      const { services, loading, error, total, refetch } = useServices({ 
        page: 1, 
        pageSize: 10,
        immediate: false 
      });

      expect(loading.value).toBe(false);
      
      await refetch();
      await nextTick();

      expect(services.value).toEqual(mockServices);
      expect(total.value).toBe(1);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockGetServices).toHaveBeenCalledWith({
        page: 1,
        pageSize: 10
      });
    });

    it('should handle search parameter correctly', async () => {
      const mockServices = [
        {
          service_id: '789',
          title: 'Cardiac Surgery',
          description: 'Advanced heart procedures',
          slug: 'cardiac-surgery',
          publishing_status: 'published'
        }
      ];

      // Mock client response
      mockGetServices.mockResolvedValue({
        services: mockServices,
        count: 1
      });

      // Direct composable call
      const { services, refetch } = useServices({ 
        search: 'cardiac',
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(mockGetServices).toHaveBeenCalledWith({
        search: 'cardiac'
      });
      expect(services.value).toEqual(mockServices);
    });

    it('should handle category filtering', async () => {
      const mockServices = [
        {
          service_id: '456',
          title: 'Primary Care Checkup',
          category_id: 'primary-care-id'
        }
      ];

      // Mock client response
      mockGetServices.mockResolvedValue({
        services: mockServices,
        count: 1
      });

      // Direct composable call
      const { services, refetch } = useServices({ 
        category: 'primary-care',
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(mockGetServices).toHaveBeenCalledWith({
        category: 'primary-care'
      });
    });

    it('should handle API errors with correlation_id', async () => {
      const errorMessage = 'Services not found';

      // Mock client to throw error
      mockGetServices.mockRejectedValue(new Error(errorMessage));

      // Direct composable call
      const { services, loading, error, refetch } = useServices({ immediate: false });

      await refetch();
      await nextTick();

      expect(services.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(errorMessage);
    });

    it('should handle rate limiting errors', async () => {
      const errorMessage = 'Rate limit exceeded: Too many requests';

      // Mock client to throw rate limit error
      mockGetServices.mockRejectedValue(new Error(errorMessage));

      // Direct composable call
      const { error, refetch } = useServices({ immediate: false });

      await refetch();
      await nextTick();

      expect(error.value).toBe(errorMessage);
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
        }
      };

      mockGetServiceBySlug.mockResolvedValue(mockServiceResponse);

      // Direct composable call - factory functions work independently
      const { service, loading, error, refetch } = useService(ref('cardiology'));
      
      // Wait for async operations to complete
      await new Promise(resolve => setTimeout(resolve, 100));

      expect(mockGetServiceBySlug).toHaveBeenCalledWith('cardiology');
      expect(service.value).toEqual(mockServiceResponse.service);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should handle null slug gracefully', async () => {
      // Direct composable call with null slug
      const { service, loading, error, refetch } = useService(ref(null));
      
      await new Promise(resolve => setTimeout(resolve, 50));

      expect(service.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockGetServiceBySlug).not.toHaveBeenCalled();
    });

    it('should react to slug changes', async () => {
      const mockServiceResponse1 = { service: { service_id: '1', title: 'Service 1', slug: 'service-1' } };
      const mockServiceResponse2 = { service: { service_id: '2', title: 'Service 2', slug: 'service-2' } };

      // Mock sequential API responses
      mockGetServiceBySlug
        .mockResolvedValueOnce(mockServiceResponse1)
        .mockResolvedValueOnce(mockServiceResponse2);

      const slugRef = ref('service-1');
      const { service, loading, error } = useService(slugRef);
      
      // Wait for initial fetch
      await new Promise(resolve => setTimeout(resolve, 50));
      expect(service.value).toEqual(mockServiceResponse1.service);

      // Change slug and wait for refetch
      slugRef.value = 'service-2';
      await new Promise(resolve => setTimeout(resolve, 50));

      expect(service.value).toEqual(mockServiceResponse2.service);
      expect(mockGetServiceBySlug).toHaveBeenCalledTimes(2);
      expect(mockGetServiceBySlug).toHaveBeenCalledWith('service-1');
      expect(mockGetServiceBySlug).toHaveBeenCalledWith('service-2');
    });
  });

  describe('useFeaturedServices', () => {
    it('should fetch published services for featured display', async () => {
      const mockFeaturedServicesResponse = {
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
        ]
      };

      mockGetFeaturedServices.mockResolvedValue(mockFeaturedServicesResponse);

      // Direct composable call - factory functions work independently
      const { services, loading, error, refetch } = useFeaturedServices();
      
      // Wait for async operations to complete
      await new Promise(resolve => setTimeout(resolve, 100));

      expect(mockGetFeaturedServices).toHaveBeenCalledWith(undefined);
      expect(services.value).toEqual(mockFeaturedServicesResponse.services);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should handle limit parameter', async () => {
      const mockLimitedServicesResponse = {
        services: [{ service_id: '1', title: 'Service 1', publishing_status: 'published' }]
      };

      mockGetFeaturedServices.mockResolvedValue(mockLimitedServicesResponse);

      // Direct composable call with limit - factory functions work independently
      const { services, loading, error } = useFeaturedServices(5);

      // Wait for async operations to complete
      await new Promise(resolve => setTimeout(resolve, 100));

      expect(mockGetFeaturedServices).toHaveBeenCalledWith(5);
      expect(services.value).toEqual(mockLimitedServicesResponse.services);
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
        ]
      };

      mockGetServiceCategories.mockResolvedValue(mockCategoriesResponse);

      // Direct composable call - factory functions work independently
      const { categories, loading, error, refetch } = useServiceCategories();
      
      // Wait for async operations to complete
      await new Promise(resolve => setTimeout(resolve, 100));

      expect(mockGetServiceCategories).toHaveBeenCalled();
      expect(categories.value).toEqual(mockCategoriesResponse.categories);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should handle empty categories response', async () => {
      const mockEmptyCategoriesResponse = { categories: [] };

      mockGetServiceCategories.mockResolvedValue(mockEmptyCategoriesResponse);

      // Direct composable call - factory functions work independently
      const { categories, loading, error } = useServiceCategories();

      // Wait for async operations to complete
      await new Promise(resolve => setTimeout(resolve, 100));

      expect(categories.value).toEqual([]);
      expect(loading.value).toBe(false);
    });
  });

  describe('error handling across composables', () => {
    it('should handle network errors consistently', async () => {
      const errorMessage = 'Network connection failed';
      
      // Mock client to throw network error
      mockGetServices.mockRejectedValue(new Error(errorMessage));

      // Direct composable call
      const { error, refetch } = useServices({ immediate: false });

      await refetch();
      await nextTick();

      expect(error.value).toBe(errorMessage);
    });

    it('should handle timeout errors', async () => {
      const errorMessage = 'Request timeout after 5000ms';
      
      // Mock client to throw timeout error
      mockGetServiceBySlug.mockRejectedValue(new Error(errorMessage));

      const { error, refetch } = useService(ref('test-service'));

      await refetch();
      await nextTick();

      expect(error.value).toBe(errorMessage);
    });

    it('should reset error state on successful refetch', async () => {
      const mockServices = [{ service_id: '1', title: 'Test Service' }];

      // Mock client behavior for error then success
      mockGetServices
        .mockRejectedValueOnce(new Error('Temporary error'))
        .mockResolvedValueOnce({
          services: mockServices,
          count: 1
        });

      const { services, error, refetch } = useServices({ immediate: false });

      // First call fails
      await refetch();
      await nextTick();
      expect(error.value).toBeTruthy(); // Error should be set

      // Second call succeeds
      await refetch();
      await nextTick();
      expect(error.value).toBe(null);
      expect(services.value).toEqual(mockServices);
    });
  });
});