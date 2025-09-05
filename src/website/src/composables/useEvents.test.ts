import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ref, nextTick, defineComponent } from 'vue';
import { mount } from '@vue/test-utils';
import { useEvents, useEvent, useFeaturedEvents, useEventCategories } from './useEvents';
import { useEventsStore } from '../stores/events';
import { RestError } from '../lib/clients/rest/BaseRestClient';

// Mock the events store
vi.mock('../stores/events', () => ({
  useEventsStore: vi.fn()
}));

let mockStore: any;

describe('useEvents composables', () => {
  beforeEach(() => {
    // Complete reset of all mocks and timers
    vi.clearAllMocks();
    vi.clearAllTimers();
    vi.resetAllMocks();
    
    // Create completely fresh mock store for each test with new refs
    mockStore = {
      events: ref([]),
      categories: ref([]),
      featuredEvents: ref([]),
      searchResults: ref([]),
      loading: ref(false),
      error: ref(null),
      total: ref(0),
      totalPages: ref(0),
      searchTotal: ref(0),
      fetchEvents: vi.fn(),
      fetchEvent: vi.fn(),
      fetchFeaturedEvents: vi.fn(),
      fetchEventCategories: vi.fn(),
      searchEvents: vi.fn(),
      setSearchResults: vi.fn()
    };
    
    // Setup mock store return with fresh implementation
    (useEventsStore as any).mockImplementation(() => mockStore);
  });
  
  afterEach(() => {
    // Additional cleanup after each test
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  describe('useEvents', () => {
    it('should initialize with correct default values', () => {
      const { events, loading, error, total, page, pageSize, totalPages } = useEvents({ enabled: false });

      expect(events.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      expect(totalPages.value).toBe(0);
    });

    it('should fetch events with backend response format including content', async () => {
      const mockEvents = [
        {
          event_id: '123',
          title: 'Medical Conference 2024',
          description: 'Annual medical conference',
          slug: 'medical-conference-2024',
          publishing_status: 'published',
          category_id: '456',
          event_type: 'conference',
          content: '<h2>Medical Conference 2024</h2><p>Join us for the annual medical conference featuring leading experts in healthcare.</p>',
          image_url: 'https://storage.azure.com/images/medical-conference.jpg',
          event_date: '2024-06-15',
          event_time: '09:00',
          location: 'Convention Center, New York'
        }
      ];

      // Pre-populate mock store state before component mounts
      mockStore.events.value = mockEvents;
      mockStore.total.value = 1;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      // Mock store behavior to maintain state
      mockStore.fetchEvents.mockImplementation(async () => {
        // State is already set above
      });

      const TestComponent = defineComponent({
        setup() {
          return useEvents({ 
            page: 1, 
            pageSize: 10,
            immediate: false 
          });
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      const { events, loading, error, total, refetch } = (wrapper.vm as any);

      expect(loading).toBe(false);
      
      await refetch();
      await nextTick();

      expect(events).toEqual(mockEvents);
      expect(total).toBe(1);
      expect(loading).toBe(false);
      expect(error).toBe(null);
      expect(mockStore.fetchEvents).toHaveBeenCalledWith(
        { page: 1, pageSize: 10 }, 
        { useCache: true }
      );
    });

    it('should handle search parameter correctly', async () => {
      const mockEvents = [
        {
          event_id: '789',
          title: 'Medical Workshop',
          description: 'Advanced medical procedures',
          slug: 'medical-workshop',
          publishing_status: 'published'
        }
      ];

      // Pre-populate mock store state
      mockStore.events.value = mockEvents;
      mockStore.total.value = 1;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      // Mock store behavior
      mockStore.fetchEvents.mockImplementation(async () => {
        // State is already set above
      });

      const TestComponent = defineComponent({
        setup() {
          return useEvents({ 
            search: 'medical',
            immediate: false 
          });
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      const { events, refetch } = (wrapper.vm as any);

      await refetch();
      await nextTick();

      expect(mockStore.fetchEvents).toHaveBeenCalledWith(
        { search: 'medical' }, 
        { useCache: true }
      );
      expect(events).toEqual(mockEvents);
    });

    it('should handle category filtering', async () => {
      const mockEvents = [
        {
          event_id: '456',
          title: 'Healthcare Seminar',
          category_id: 'healthcare-id'
        }
      ];

      // Pre-populate mock store state
      mockStore.events.value = mockEvents;
      mockStore.total.value = 1;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      // Mock store behavior
      mockStore.fetchEvents.mockImplementation(async () => {
        // State is already set above
      });

      const TestComponent = defineComponent({
        setup() {
          return useEvents({ 
            category: 'healthcare',
            immediate: false 
          });
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      const { events, refetch } = (wrapper.vm as any);

      await refetch();
      await nextTick();

      expect(mockStore.fetchEvents).toHaveBeenCalledWith(
        { category: 'healthcare' }, 
        { useCache: true }
      );
    });

    it('should handle API errors with correlation_id', async () => {
      const errorMessage = 'Events not found';

      // Pre-populate mock store state for error
      mockStore.events.value = [];
      mockStore.total.value = 0;
      mockStore.loading.value = false;
      mockStore.error.value = errorMessage;

      // Mock store behavior
      mockStore.fetchEvents.mockImplementation(async () => {
        // State is already set above
      });

      const TestComponent = defineComponent({
        setup() {
          return useEvents({ immediate: false });
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      const { events, loading, error, refetch } = (wrapper.vm as any);

      await refetch();
      await nextTick();

      expect(events).toEqual([]);
      expect(loading).toBe(false);
      expect(error).toBe(errorMessage);
    });

    it('should handle rate limiting errors', async () => {
      const errorMessage = 'Rate limit exceeded: Too many requests';

      // Pre-populate mock store state for rate limit error
      mockStore.events.value = [];
      mockStore.total.value = 0;
      mockStore.loading.value = false;
      mockStore.error.value = errorMessage;

      // Mock store behavior
      mockStore.fetchEvents.mockImplementation(async () => {
        // State is already set above
      });

      const TestComponent = defineComponent({
        setup() {
          return useEvents({ immediate: false });
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      const { error, refetch } = (wrapper.vm as any);

      await refetch();
      await nextTick();

      expect(error).toBe(errorMessage);
    });
  });

  describe('useEvent', () => {
    it('should fetch event by slug with backend format', async () => {
      const mockEvent = {
        event_id: '123',
        title: 'Medical Conference 2024',
        description: 'Comprehensive medical conference',
        slug: 'medical-conference',
        publishing_status: 'published',
        category_id: '456',
        event_type: 'conference',
        content: '<h2>Medical Conference 2024</h2><p>Join us for comprehensive medical training including:</p><ul><li>Advanced diagnostic techniques</li><li>Treatment protocols</li><li>Research presentations</li><li>Networking sessions</li></ul><p>Contact us to register.</p>'
      };

      // Pre-populate mock store state
      mockStore.loading.value = false;
      mockStore.error.value = null;

      // Mock store behavior
      mockStore.fetchEvent.mockResolvedValueOnce(mockEvent);

      const TestComponent = defineComponent({
        setup() {
          return useEvent(ref('medical-conference'));
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      // Wait for initial fetch
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      const { event, loading, error, refetch } = (wrapper.vm as any);

      expect(event).toEqual(mockEvent);
      expect(loading).toBe(false);
      expect(error).toBe(null);
      expect(mockStore.fetchEvent).toHaveBeenCalledWith('medical-conference');
    });

    it('should handle null slug gracefully', async () => {
      // Pre-populate mock store state
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const TestComponent = defineComponent({
        setup() {
          return useEvent(ref(null));
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      await nextTick();

      const { event, loading } = (wrapper.vm as any);

      expect(event).toBe(null);
      expect(loading).toBe(false);
      expect(mockStore.fetchEvent).not.toHaveBeenCalled();
    });

    it('should react to slug changes', async () => {
      const mockEvent1 = { event_id: '1', title: 'Event 1', slug: 'event-1' };
      const mockEvent2 = { event_id: '2', title: 'Event 2', slug: 'event-2' };

      // Pre-populate mock store state
      mockStore.loading.value = false;
      mockStore.error.value = null;

      // Mock store behavior for sequential calls
      mockStore.fetchEvent
        .mockResolvedValueOnce(mockEvent1)
        .mockResolvedValueOnce(mockEvent2);

      const slugRef = ref('event-1');
      
      const TestComponent = defineComponent({
        setup() {
          return useEvent(slugRef);
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      const { event } = (wrapper.vm as any);
      expect(event).toEqual(mockEvent1);

      // Change slug
      slugRef.value = 'event-2';
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(event).toEqual(mockEvent2);
      expect(mockStore.fetchEvent).toHaveBeenCalledTimes(2);
    });
  });

  describe('useFeaturedEvents', () => {
    it('should fetch published events for featured display', async () => {
      const mockFeaturedEvents = [
        {
          event_id: '789',
          title: 'Featured Medical Seminar',
          publishing_status: 'published',
          featured: true
        },
        {
          event_id: '101',
          title: 'Featured Healthcare Workshop',
          publishing_status: 'published',
          featured: true
        }
      ];

      // Pre-populate mock store state
      mockStore.featuredEvents.value = mockFeaturedEvents;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      // Mock store behavior
      mockStore.fetchFeaturedEvents.mockImplementation(async () => {
        // State is already set above
      });

      const TestComponent = defineComponent({
        setup() {
          return useFeaturedEvents();
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      const { events, loading, error } = (wrapper.vm as any);

      expect(events).toEqual(mockFeaturedEvents);
      expect(loading).toBe(false);
      expect(error).toBe(null);
      expect(mockStore.fetchFeaturedEvents).toHaveBeenCalledWith(undefined);
    });

    it('should handle limit parameter', async () => {
      const mockLimitedEvents = [
        { event_id: '1', title: 'Event 1', publishing_status: 'published' }
      ];

      // Pre-populate mock store state
      mockStore.featuredEvents.value = mockLimitedEvents;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      // Mock store behavior
      mockStore.fetchFeaturedEvents.mockImplementation(async () => {
        // State is already set above
      });

      const { events } = useFeaturedEvents(5);

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(mockStore.fetchFeaturedEvents).toHaveBeenCalledWith(5);
    });
  });

  describe('useEventCategories', () => {
    it('should fetch categories with backend format', async () => {
      const mockCategories = [
        {
          category_id: '456',
          name: 'Medical Events',
          slug: 'medical-events',
          order_number: 1,
          is_default_unassigned: false
        },
        {
          category_id: '789',
          name: 'Educational Events',
          slug: 'educational-events',
          order_number: 2,
          is_default_unassigned: false
        }
      ];

      // Pre-populate mock store state
      mockStore.categories.value = mockCategories;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      // Mock store behavior
      mockStore.fetchEventCategories.mockImplementation(async () => {
        // State is already set above
      });

      const TestComponent = defineComponent({
        setup() {
          return useEventCategories();
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      const { categories, loading, error } = (wrapper.vm as any);

      expect(categories).toEqual(mockCategories);
      expect(loading).toBe(false);
      expect(error).toBe(null);
      expect(mockStore.fetchEventCategories).toHaveBeenCalled();
    });

    it('should handle empty categories response', async () => {
      // Pre-populate mock store state for empty response
      mockStore.categories.value = [];
      mockStore.loading.value = false;
      mockStore.error.value = null;

      // Mock store behavior
      mockStore.fetchEventCategories.mockImplementation(async () => {
        // State is already set above
      });

      const { categories } = useEventCategories();

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(categories.value).toEqual([]);
    });
  });

  describe('error handling across composables', () => {
    it('should handle network errors consistently', async () => {
      const errorMessage = 'Network connection failed';
      
      // Pre-populate mock store state for network error
      mockStore.events.value = [];
      mockStore.total.value = 0;
      mockStore.loading.value = false;
      mockStore.error.value = errorMessage;

      // Mock store behavior
      mockStore.fetchEvents.mockImplementation(async () => {
        // State is already set above
      });

      const { error, refetch } = useEvents({ immediate: false });

      await refetch();
      await nextTick();

      expect(error.value).toBe(errorMessage);
    });

    it('should handle timeout errors', async () => {
      const errorMessage = 'Request timeout after 5000ms';
      
      // Pre-populate mock store state for timeout error
      mockStore.loading.value = false;
      mockStore.error.value = errorMessage;

      // Mock store behavior
      mockStore.fetchEvent.mockImplementation(async () => {
        // State is already set above
        return null;
      });

      const { error } = useEvent(ref('test-event'));

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(error.value).toBe(errorMessage);
    });

    it('should reset error state on successful refetch', async () => {
      const mockEvents = [{ event_id: '1', title: 'Test Event' }];

      // Mock store behavior for error then success
      mockStore.fetchEvents
        .mockImplementationOnce(async () => {
          mockStore.error.value = 'Temporary error';
          mockStore.events.value = [];
          mockStore.loading.value = false;
        })
        .mockImplementationOnce(async () => {
          mockStore.error.value = null;
          mockStore.events.value = mockEvents;
          mockStore.loading.value = false;
        });

      const { error, refetch } = useEvents({ immediate: false });

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