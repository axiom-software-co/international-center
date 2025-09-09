import { describe, it, expect, beforeEach } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useServicesStore } from './services';
import type { Service, ServiceCategory } from '../lib/clients/services/types';

describe('ServicesStore', () => {
  beforeEach(() => {
    // Create fresh pinia instance for each test
    setActivePinia(createPinia());
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

  describe('State Mutation Methods', () => {
    it('should set services with pagination data', () => {
      const store = useServicesStore();
      const mockServices: Service[] = [
        {
          service_id: 'service-1',
          title: 'Test Service',
          description: 'Test Description',
          slug: 'test-service',
          category_id: 'cat-1',
          publishing_status: 'published',
          delivery_mode: 'outpatient_service',
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.setServices(mockServices, 100, 2, 20);
      
      expect(store.services).toEqual(mockServices);
      expect(store.total).toBe(100);
      expect(store.page).toBe(2);
      expect(store.pageSize).toBe(20);
    });

    it('should set featured services', () => {
      const store = useServicesStore();
      const mockFeatured: Service[] = [
        {
          service_id: 'featured-1',
          title: 'Featured Service',
          description: 'Featured Description',
          slug: 'featured-service',
          category_id: 'cat-1',
          publishing_status: 'published',
          delivery_mode: 'inpatient_service',
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.$patch({ featuredServices: mockFeatured });
      
      expect(store.featuredServices).toEqual(mockFeatured);
    });

    it('should set service categories', () => {
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
      
      store.$patch({ categories: mockCategories });
      
      expect(store.categories).toEqual(mockCategories);
    });

    it('should set search results with total', () => {
      const store = useServicesStore();
      const mockResults: Service[] = [
        {
          service_id: 'search-1',
          title: 'Search Result',
          description: 'Search Description',
          slug: 'search-result',
          category_id: 'cat-1',
          publishing_status: 'published',
          delivery_mode: 'outpatient_service',
          order_number: 1,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.$patch({ searchResults: mockResults, searchTotal: 1 });
      
      expect(store.searchResults).toEqual(mockResults);
      expect(store.searchTotal).toBe(1);
    });

    it('should clear search results', () => {
      const store = useServicesStore();
      
      // Set initial search data
      store.$patch({ searchResults: [{ service_id: 'test' } as Service], searchTotal: 1 });
      
      // Clear search results
      store.$patch({ searchResults: [], searchTotal: 0 });
      
      expect(store.searchResults).toEqual([]);
      expect(store.searchTotal).toBe(0);
    });
  });

  describe('Cache Management Methods', () => {
    it('should have cache invalidation method', () => {
      const store = useServicesStore();
      
      // Set some cached data
      store.$patch({ services: [{ service_id: 'cached' } as Service] });
      
      // Cache invalidation should be available as method
      expect(typeof store.invalidateCache).toBe('function');
      
      // Call invalidate cache
      store.invalidateCache();
      
      // State should remain but cache should be cleared internally
      expect(store.services).toBeDefined();
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

});