import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useServicesStore } from './services';
import type { Service, ServiceCategory, GetServicesParams, SearchServicesParams } from '../lib/clients/services/types';

// Mock the services client
vi.mock('../lib/clients', () => ({
  servicesClient: {
    getServices: vi.fn(),
    getServiceBySlug: vi.fn(),
    getFeaturedServices: vi.fn(),
    searchServices: vi.fn(),
    getServiceCategories: vi.fn(),
  }
}));

describe('ServicesStore', () => {
  beforeEach(() => {
    // Create fresh pinia instance for each test
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Initial State', () => {
    it('should initialize with empty state and default values', () => {
      const store = useServicesStore();
      
      expect(store.services).toEqual([]);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(store.total).toBe(0);
      expect(store.page).toBe(1);
      expect(store.pageSize).toBe(10);
      expect(store.categories).toEqual([]);
      expect(store.featuredServices).toEqual([]);
      expect(store.searchResults).toEqual([]);
    });

    it('should provide totalPages getter based on total and pageSize', () => {
      const store = useServicesStore();
      
      // Initially should be 0
      expect(store.totalPages).toBe(0);
      
      // Set some data to test calculation
      store.$patch({
        total: 25,
        pageSize: 10
      });
      
      expect(store.totalPages).toBe(3); // Math.ceil(25/10) = 3
    });
  });

  describe('State Management', () => {
    it('should manage loading state during operations', () => {
      const store = useServicesStore();
      
      expect(store.loading).toBe(false);
      
      // Should be able to set loading state
      store.setLoading(true);
      expect(store.loading).toBe(true);
      
      store.setLoading(false);
      expect(store.loading).toBe(false);
    });

    it('should manage error state with proper clearing', () => {
      const store = useServicesStore();
      
      expect(store.error).toBeNull();
      
      // Should set error
      store.setError('Network error occurred');
      expect(store.error).toBe('Network error occurred');
      
      // Should clear error
      store.clearError();
      expect(store.error).toBeNull();
    });

    it('should update services data and pagination state', () => {
      const store = useServicesStore();
      const mockServices: Service[] = [
        {
          service_id: 'service-1',
          title: 'Test Service 1',
          description: 'Service Description',
          slug: 'test-service-1',
          category_id: 'cat-1',
          publishing_status: 'published',
          delivery_mode: 'outpatient_service',
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.setServices(mockServices, 25, 2, 10);
      
      expect(store.services).toEqual(mockServices);
      expect(store.total).toBe(25);
      expect(store.page).toBe(2);
      expect(store.pageSize).toBe(10);
    });
  });

  describe('Actions - Services Operations', () => {
    it('should fetch services with parameters and update state accordingly', async () => {
      const { servicesClient } = await import('../lib/clients');
      const store = useServicesStore();
      
      const mockResponse = {
        services: [
          {
            service_id: 'service-1',
            title: 'Test Service',
            description: 'Test Description',
            slug: 'test-service',
            category_id: 'cat-1',
            publishing_status: 'published' as const,
            delivery_mode: 'outpatient_service' as const,
            order_number: 1,
            created_on: '2024-01-01T00:00:00Z',
            is_deleted: false
          }
        ],
        count: 1,
        correlation_id: 'test-correlation-id'
      };
      
      vi.mocked(servicesClient.getServices).mockResolvedValueOnce(mockResponse);
      
      const params: GetServicesParams = { page: 1, pageSize: 10 };
      await store.fetchServices(params);
      
      expect(servicesClient.getServices).toHaveBeenCalledWith(params);
      expect(store.services).toEqual(mockResponse.services);
      expect(store.total).toBe(mockResponse.count);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    it('should handle services fetch errors properly', async () => {
      const { servicesClient } = await import('../lib/clients');
      const store = useServicesStore();
      
      const errorMessage = 'Failed to fetch services';
      vi.mocked(servicesClient.getServices).mockRejectedValueOnce(new Error(errorMessage));
      
      await store.fetchServices();
      
      expect(store.error).toBe(errorMessage);
      expect(store.loading).toBe(false);
      expect(store.services).toEqual([]);
    });

    it('should fetch single service by slug', async () => {
      const { servicesClient } = await import('../lib/clients');
      const store = useServicesStore();
      
      const mockService: Service = {
        service_id: 'service-1',
        title: 'Single Service',
        description: 'Service Description',
        slug: 'single-service',
        category_id: 'cat-1',
        publishing_status: 'published',
        delivery_mode: 'mobile_service',
        order_number: 1,
        created_on: '2024-01-01T00:00:00Z',
        is_deleted: false,
        content: 'Full service content'
      };
      
      const mockResponse = {
        service: mockService,
        correlation_id: 'service-correlation-id'
      };
      
      vi.mocked(servicesClient.getServiceBySlug).mockResolvedValueOnce(mockResponse);
      
      const result = await store.fetchService('single-service');
      
      expect(servicesClient.getServiceBySlug).toHaveBeenCalledWith('single-service');
      expect(result).toEqual(mockService);
    });
  });

  describe('Actions - Featured Services', () => {
    it('should fetch featured services and cache in store', async () => {
      const { servicesClient } = await import('../lib/clients');
      const store = useServicesStore();
      
      const mockFeaturedServices = [
        {
          service_id: 'featured-1',
          title: 'Featured Service',
          description: 'Featured Description',
          slug: 'featured-service',
          category_id: 'cat-1',
          publishing_status: 'published' as const,
          delivery_mode: 'inpatient_service' as const,
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      const mockResponse = {
        services: mockFeaturedServices,
        count: 1,
        correlation_id: 'featured-correlation-id'
      };
      
      vi.mocked(servicesClient.getFeaturedServices).mockResolvedValueOnce(mockResponse);
      
      await store.fetchFeaturedServices(5);
      
      expect(servicesClient.getFeaturedServices).toHaveBeenCalledWith(5);
      expect(store.featuredServices).toEqual(mockFeaturedServices);
    });
  });

  describe('Actions - Search Operations', () => {
    it('should perform services search and store results separately', async () => {
      const { servicesClient } = await import('../lib/clients');
      const store = useServicesStore();
      
      const mockSearchResults = [
        {
          service_id: 'search-1',
          title: 'Search Result',
          description: 'Search Description',
          slug: 'search-result',
          category_id: 'cat-1',
          publishing_status: 'published' as const,
          delivery_mode: 'outpatient_service' as const,
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      const mockResponse = {
        services: mockSearchResults,
        count: 1,
        correlation_id: 'search-correlation-id'
      };
      
      vi.mocked(servicesClient.searchServices).mockResolvedValueOnce(mockResponse);
      
      const searchParams: SearchServicesParams = {
        q: 'test query',
        page: 1,
        pageSize: 10
      };
      
      await store.searchServices(searchParams);
      
      expect(servicesClient.searchServices).toHaveBeenCalledWith(searchParams);
      expect(store.searchResults).toEqual(mockSearchResults);
      expect(store.searchTotal).toBe(1);
    });

    it('should clear search results when query is empty', async () => {
      const store = useServicesStore();
      
      // Set some initial search results
      store.$patch({
        searchResults: [{ 
          service_id: 'test',
          title: 'Test',
          description: 'Test',
          slug: 'test',
          category_id: 'cat-1',
          publishing_status: 'published' as const,
          delivery_mode: 'outpatient_service' as const,
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }],
        searchTotal: 1
      });
      
      await store.searchServices({ q: '', page: 1, pageSize: 10 });
      
      expect(store.searchResults).toEqual([]);
      expect(store.searchTotal).toBe(0);
    });
  });

  describe('Actions - Categories', () => {
    it('should fetch and cache service categories', async () => {
      const { servicesClient } = await import('../lib/clients');
      const store = useServicesStore();
      
      const mockCategories: ServiceCategory[] = [
        {
          category_id: 'cat-1',
          name: 'Healthcare Services',
          slug: 'healthcare-services',
          order_number: 1,
          is_default_unassigned: false,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      const mockResponse = {
        categories: mockCategories,
        count: 1,
        correlation_id: 'categories-correlation-id'
      };
      
      vi.mocked(servicesClient.getServiceCategories).mockResolvedValueOnce(mockResponse);
      
      await store.fetchServiceCategories();
      
      expect(servicesClient.getServiceCategories).toHaveBeenCalled();
      expect(store.categories).toEqual(mockCategories);
    });
  });

  describe('Getters', () => {
    it('should provide computed values for pagination', () => {
      const store = useServicesStore();
      
      store.$patch({
        total: 0,
        pageSize: 10
      });
      expect(store.totalPages).toBe(0);
      
      store.$patch({
        total: 15,
        pageSize: 5
      });
      expect(store.totalPages).toBe(3);
    });

    it('should provide hasServices getter for conditional rendering', () => {
      const store = useServicesStore();
      
      expect(store.hasServices).toBe(false);
      
      store.$patch({
        services: [{
          service_id: 'test',
          title: 'Test',
          description: 'Test',
          slug: 'test',
          category_id: 'cat-1',
          publishing_status: 'published',
          delivery_mode: 'outpatient_service',
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }]
      });
      
      expect(store.hasServices).toBe(true);
    });

    it('should provide services grouped by category', () => {
      const store = useServicesStore();
      
      const mockServices = [
        {
          service_id: 'service-1',
          title: 'Service 1',
          description: 'Description 1',
          slug: 'service-1',
          category_id: 'cat-1',
          publishing_status: 'published' as const,
          delivery_mode: 'outpatient_service' as const,
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        },
        {
          service_id: 'service-2',
          title: 'Service 2',
          description: 'Description 2',
          slug: 'service-2',
          category_id: 'cat-2',
          publishing_status: 'published' as const,
          delivery_mode: 'mobile_service' as const,
          order_number: 2,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];

      store.$patch({ services: mockServices });
      
      const grouped = store.servicesByCategory;
      expect(grouped).toHaveProperty('cat-1');
      expect(grouped).toHaveProperty('cat-2');
      expect(grouped['cat-1']).toHaveLength(1);
      expect(grouped['cat-2']).toHaveLength(1);
    });
  });

  describe('Cache Management', () => {
    it('should cache services data and avoid duplicate fetches', async () => {
      const { servicesClient } = await import('../lib/clients');
      const store = useServicesStore();
      
      const mockResponse = {
        services: [{
          service_id: 'cached-1',
          title: 'Cached Service',
          description: 'Cached Description',
          slug: 'cached-service',
          category_id: 'cat-1',
          publishing_status: 'published' as const,
          delivery_mode: 'outpatient_service' as const,
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }],
        count: 1,
        correlation_id: 'cache-test-id'
      };
      
      vi.mocked(servicesClient.getServices).mockResolvedValueOnce(mockResponse);
      
      // First fetch should call API
      await store.fetchServices({ page: 1, pageSize: 10 });
      expect(servicesClient.getServices).toHaveBeenCalledTimes(1);
      
      // Second fetch with same params should use cache
      await store.fetchServices({ page: 1, pageSize: 10 }, { useCache: true });
      expect(servicesClient.getServices).toHaveBeenCalledTimes(1); // Still only called once
    });

    it('should invalidate cache and refetch when requested', async () => {
      const { servicesClient } = await import('../lib/clients');
      const store = useServicesStore();
      
      const mockResponse = {
        services: [],
        count: 0,
        correlation_id: 'invalidate-test-id'
      };
      
      vi.mocked(servicesClient.getServices).mockResolvedValue(mockResponse);
      
      // First fetch
      await store.fetchServices({ page: 1, pageSize: 10 });
      expect(servicesClient.getServices).toHaveBeenCalledTimes(1);
      
      // Invalidate cache and fetch again
      store.invalidateCache();
      await store.fetchServices({ page: 1, pageSize: 10 });
      expect(servicesClient.getServices).toHaveBeenCalledTimes(2);
    });
  });
});