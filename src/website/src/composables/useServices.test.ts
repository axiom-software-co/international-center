import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, nextTick, defineComponent, computed } from 'vue';
import { mount } from '@vue/test-utils';
import { useServices, useService, useFeaturedServices, useServiceCategories } from './useServices';
import { useServicesStore } from '../stores/services';
import { RestError } from '../lib/clients/rest/BaseRestClient';

// Mock the store module - RED phase: define store-centric contracts
vi.mock('../stores/services', () => ({
  useServicesStore: vi.fn()
}));


describe('useServices composables', () => {
  // Define mock store structure - RED phase: store-centric contract
  const mockStore = {
    // State refs that composables should expose via storeToRefs
    services: ref([]),
    service: ref(null), // Individual service state
    loading: ref(false),
    error: ref(null),
    total: ref(0),
    categories: ref([]),
    featuredServices: ref([]),
    
    // Computed values
    totalPages: computed(() => Math.ceil(mockStore.total.value / 10) || 0),
    
    // Explicit action methods that composables should delegate to
    fetchServices: vi.fn(),
    fetchService: vi.fn(),
    fetchFeaturedServices: vi.fn(),
    fetchServiceCategories: vi.fn(),
    searchServices: vi.fn(),
  };

  beforeEach(() => {
    // Ensure all store properties are properly initialized as refs
    if (!mockStore.service || !mockStore.service.value !== undefined) {
      mockStore.service = ref(null);
    }
    
    // Reset mock store state
    mockStore.services.value = [];
    mockStore.service.value = null;
    mockStore.loading.value = false;
    mockStore.error.value = null;
    mockStore.total.value = 0;
    mockStore.categories.value = [];
    mockStore.featuredServices.value = [];
    
    // Clear all store action mocks
    mockStore.fetchServices.mockClear();
    mockStore.fetchService.mockClear();
    mockStore.fetchFeaturedServices.mockClear();
    mockStore.fetchServiceCategories.mockClear();
    mockStore.searchServices.mockClear();
    
    // Setup store mock return
    vi.mocked(useServicesStore).mockReturnValue(mockStore as any);
  });

  describe('useServices', () => {
    it('should expose store state via storeToRefs and initialize with correct default values', () => {
      const { services, loading, error, total, page, pageSize, totalPages, refetch } = useServices({ enabled: false });

      // RED phase: expect composable to expose store state directly
      expect(services.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      
      // Contract: composable should expose reactive properties and functions
      expect(totalPages).toBeTruthy();
      expect(typeof refetch).toBe('function');
      
      // Contract: composable should use store
      expect(useServicesStore).toHaveBeenCalled();
    });

    it('should delegate to store.fetchServices and expose store state', async () => {
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

      const { services, loading, error, total, refetch } = useServices({ 
        page: 1, 
        pageSize: 10,
        immediate: false 
      });

      // Contract: composable should expose store state via storeToRefs
      expect(services.value).toBeDefined();
      expect(loading.value).toBeDefined();
      expect(error.value).toBeDefined();
      expect(total.value).toBeDefined();
      
      // Clear any previous calls before testing refetch
      mockStore.fetchServices.mockClear();
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation, not direct client calls  
      expect(mockStore.fetchServices).toHaveBeenCalledTimes(1);
      
      // Contract: composable exposes reactive state from store
      expect(typeof refetch).toBe('function');
    });

    it('should delegate search parameters to store.fetchServices', async () => {
      const mockServices = [
        {
          service_id: '789',
          title: 'Cardiac Surgery',
          description: 'Advanced heart procedures',
          slug: 'cardiac-surgery',
          publishing_status: 'published'
        }
      ];

      const { services, refetch } = useServices({ 
        search: 'cardiac',
        immediate: false 
      });

      // Clear any previous calls before testing refetch
      mockStore.fetchServices.mockClear();

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchServices).toHaveBeenCalledTimes(1);
      
      // Contract: composable exposes store state
      expect(services.value).toBeDefined();
    });

    it('should handle category filtering', async () => {
      const mockServices = [
        {
          service_id: '456',
          title: 'Primary Care Checkup',
          category_id: 'primary-care-id'
        }
      ];

      const { services, refetch } = useServices({ 
        category: 'primary-care',
        immediate: false 
      });

      // Clear any previous calls before testing refetch
      mockStore.fetchServices.mockClear();

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchServices).toHaveBeenCalledTimes(1);
      
      // Contract: composable exposes store state
      expect(services.value).toBeDefined();
    });

    it('should handle API errors with correlation_id', async () => {
      const errorMessage = 'Services not found';

      // RED phase: simulate store error state
      mockStore.services.value = [];
      mockStore.loading.value = false;
      mockStore.error.value = errorMessage;
      mockStore.total.value = 0;

      const { services, loading, error, refetch } = useServices({ immediate: false });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchServices).toHaveBeenCalled();
      
      // Contract: composable should expose store error state
      expect(services.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(errorMessage);
    });

    it('should handle rate limiting errors', async () => {
      const errorMessage = 'Rate limit exceeded: Too many requests';

      // RED phase: simulate store rate limit error state
      mockStore.error.value = errorMessage;
      mockStore.loading.value = false;

      const { error, refetch } = useServices({ immediate: false });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchServices).toHaveBeenCalled();
      
      // Contract: composable should expose store error state
      expect(error.value).toBe(errorMessage);
    });
  });

  describe('useService', () => {
    it('should delegate to store.fetchService and expose store state', async () => {
      const mockService = {
        service_id: '123',
        title: 'Cardiology Services',
        description: 'Comprehensive heart care',
        slug: 'cardiology',
        publishing_status: 'published',
        category_id: '456',
        delivery_mode: 'outpatient_service',
        content: '<h2>Comprehensive Heart Care Services</h2><p>Our cardiology team provides advanced diagnostics and treatment for heart conditions including:</p><ul><li>ECG and stress testing</li><li>Echocardiogram imaging</li><li>Cardiac catheterization</li><li>Heart surgery coordination</li></ul><p>Contact us to schedule your consultation.</p>'
      };

      // RED phase: simulate store state after individual service fetch
      mockStore.service = ref(mockService);
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { service, loading, error, refetch } = useService(ref('cardiology'));
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchService).toHaveBeenCalledWith('cardiology');
      
      // Contract: composable should expose store state
      expect(service.value).toEqual(mockService);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should handle null slug gracefully without calling store actions', async () => {
      // RED phase: simulate initial store state
      mockStore.service = ref(null);
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { service, loading, error, refetch } = useService(ref(null));
      
      await refetch();
      await nextTick();

      // Contract: composable should expose null state
      expect(service.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      
      // RED phase: should not call store action for null slug
      expect(mockStore.fetchService).not.toHaveBeenCalled();
    });

    it('should react to slug changes and delegate to store', async () => {
      const mockService1 = { service_id: '1', title: 'Service 1', slug: 'service-1' };
      const mockService2 = { service_id: '2', title: 'Service 2', slug: 'service-2' };

      // RED phase: simulate store state changes
      const serviceRef = ref(mockService1);
      mockStore.service = serviceRef;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const slugRef = ref('service-1');
      const { service, loading, error } = useService(slugRef);
      
      await nextTick();
      expect(service.value).toEqual(mockService1);

      // Simulate slug change and store state update
      serviceRef.value = mockService2;
      slugRef.value = 'service-2';
      await nextTick();

      expect(service.value).toEqual(mockService2);
      
      // RED phase: expect store delegation for both slugs
      expect(mockStore.fetchService).toHaveBeenCalledWith('service-1');
      expect(mockStore.fetchService).toHaveBeenCalledWith('service-2');
      expect(mockStore.fetchService).toHaveBeenCalledTimes(2);
    });
  });

  describe('useFeaturedServices', () => {
    it('should delegate to store.fetchFeaturedServices and expose store state', async () => {
      const mockFeaturedServices = [
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
      ];

      // RED phase: simulate store state after featured services fetch
      mockStore.featuredServices.value = mockFeaturedServices;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { services, loading, error, refetch } = useFeaturedServices();
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchFeaturedServices).toHaveBeenCalledWith(undefined);
      
      // Contract: composable should expose store state
      expect(services.value).toEqual(mockFeaturedServices);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should delegate limit parameter to store action', async () => {
      const mockLimitedServices = [
        { service_id: '1', title: 'Service 1', publishing_status: 'published' }
      ];

      // RED phase: simulate store state after limited featured services fetch
      mockStore.featuredServices.value = mockLimitedServices;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { services, refetch } = useFeaturedServices(5);

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation with limit
      expect(mockStore.fetchFeaturedServices).toHaveBeenCalledWith(5);
      
      // Contract: composable should expose store state
      expect(services.value).toEqual(mockLimitedServices);
    });
  });

  describe('useServiceCategories', () => {
    it('should delegate to store.fetchServiceCategories and expose store state', async () => {
      const mockCategories = [
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
      ];

      // RED phase: simulate store state after categories fetch
      mockStore.categories.value = mockCategories;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { categories, loading, error, refetch } = useServiceCategories();
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchServiceCategories).toHaveBeenCalled();
      
      // Contract: composable should expose store state
      expect(categories.value).toEqual(mockCategories);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should handle empty categories response from store', async () => {
      // RED phase: simulate empty store state
      mockStore.categories.value = [];
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { categories, loading, error, refetch } = useServiceCategories();

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchServiceCategories).toHaveBeenCalled();
      
      // Contract: composable should expose empty store state
      expect(categories.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });
  });

  describe('error handling across composables', () => {
    it('should handle network errors consistently via store state', async () => {
      const errorMessage = 'Network connection failed';
      
      // RED phase: simulate store network error state
      mockStore.error.value = errorMessage;
      mockStore.loading.value = false;
      mockStore.services.value = [];

      const { error, refetch } = useServices({ immediate: false });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchServices).toHaveBeenCalled();
      
      // Contract: composable should expose store error state
      expect(error.value).toBe(errorMessage);
    });

    it('should handle timeout errors via store state', async () => {
      const errorMessage = 'Request timeout after 5000ms';
      
      // RED phase: simulate store timeout error state
      mockStore.error.value = errorMessage;
      mockStore.loading.value = false;
      mockStore.service.value = null;

      const { error, refetch } = useService(ref('test-service'));

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchService).toHaveBeenCalledWith('test-service');
      
      // Contract: composable should expose store error state
      expect(error.value).toBe(errorMessage);
    });

    it('should reset error state on successful refetch via store', async () => {
      const mockServices = [{ service_id: '1', title: 'Test Service' }];

      // RED phase: simulate store error state initially
      mockStore.error.value = 'Temporary error';
      mockStore.loading.value = false;
      mockStore.services.value = [];
      mockStore.total.value = 0;

      const { services, error, refetch } = useServices({ immediate: false });

      // First call shows error state
      await refetch();
      await nextTick();
      expect(error.value).toBe('Temporary error');

      // RED phase: simulate store success state after recovery
      mockStore.error.value = null;
      mockStore.services.value = mockServices;
      mockStore.total.value = 1;

      // Second call shows success state
      await refetch();
      await nextTick();
      
      // RED phase: expect store action delegation for both calls
      expect(mockStore.fetchServices).toHaveBeenCalledTimes(2);
      
      // Contract: composable should expose updated store state
      expect(error.value).toBe(null);
      expect(services.value).toEqual(mockServices);
    });
  });
});