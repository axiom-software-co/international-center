import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, computed } from 'vue';
import { useEvents, useEvent, useFeaturedEvents, useEventCategories } from './useEvents';
import { useEventsStore } from '../stores/events';

// Mock the events store
vi.mock('../stores/events', () => ({
  useEventsStore: vi.fn()
}));

let mockStore: any;

describe('useEvents composables', () => {
  beforeEach(() => {
    // Complete reset of all mocks
    vi.clearAllMocks();
    
    // Create fresh mock store for each test with new refs
    const totalRef = ref(0);
    const pageSizeRef = ref(10);
    
    mockStore = {
      events: ref([]),
      categories: ref([]),
      featuredEvents: ref([]),
      searchResults: ref([]),
      event: ref(null),
      loading: ref(false),
      error: ref(null),
      total: totalRef,
      totalPages: computed(() => Math.ceil(totalRef.value / pageSizeRef.value) || 0),
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

  describe('useEvents', () => {
    it('should initialize with correct default values', () => {
      const { events, loading, error, total, page, pageSize, totalPages, refetch } = useEvents({ enabled: false });

      expect(events.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      
      // Contract: composable should expose reactive properties and functions
      expect(totalPages).toBeTruthy();
      expect(typeof refetch).toBe('function');
    });

    it('should expose store state reactively', () => {
      const mockEvents = [
        {
          event_id: '123',
          title: 'Medical Conference 2024',
          description: 'Annual medical conference',
          slug: 'medical-conference-2024',
          publishing_status: 'published',
          category_id: '456',
          event_type: 'conference'
        }
      ];

      // Pre-populate mock store state
      mockStore.events.value = mockEvents;
      mockStore.total.value = 1;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { events, loading, error, total } = useEvents({ immediate: false });

      // Composable should expose store state reactively
      expect(events.value).toEqual(mockEvents);
      expect(total.value).toBe(1);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should handle search parameter correctly', () => {
      const { refetch } = useEvents({ 
        search: 'medical',
        immediate: false 
      });

      // Call refetch and verify store method was called with search parameter
      refetch();
      
      expect(mockStore.fetchEvents).toHaveBeenCalledWith(
        expect.objectContaining({
          search: 'medical'
        })
      );
    });

    it('should handle category filtering', () => {
      const { refetch } = useEvents({ 
        category: 'healthcare',
        immediate: false 
      });

      // Call refetch and verify store method was called with category parameter
      refetch();
      
      expect(mockStore.fetchEvents).toHaveBeenCalledWith(
        expect.objectContaining({
          category: 'healthcare'
        })
      );
    });

    it('should expose error state from store', () => {
      const errorMessage = 'Events not found';

      // Pre-populate mock store state for error
      mockStore.events.value = [];
      mockStore.total.value = 0;
      mockStore.loading.value = false;
      mockStore.error.value = errorMessage;

      const { events, loading, error } = useEvents({ immediate: false });

      expect(events.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(errorMessage);
    });

    it('should provide pagination parameters correctly', () => {
      const { page, pageSize } = useEvents({ 
        page: 2, 
        pageSize: 20,
        immediate: false 
      });

      expect(page.value).toBe(2);
      expect(pageSize.value).toBe(20);
      
      // Verify composable exposes expected pagination properties
      expect(typeof page.value).toBe('number');
      expect(typeof pageSize.value).toBe('number');
    });
  });

  describe('useEvent', () => {
    it('should call store fetchEvent with slug parameter', () => {
      mockStore.fetchEvent.mockClear();
      const slugRef = ref('medical-conference');
      
      useEvent(slugRef);

      expect(mockStore.fetchEvent).toHaveBeenCalledWith('medical-conference');
    });

    it('should handle null slug gracefully', () => {
      mockStore.fetchEvent.mockClear();
      const slugRef = ref(null);
      
      const { event } = useEvent(slugRef);

      // Should not call store fetchEvent when slug is null
      expect(mockStore.fetchEvent).not.toHaveBeenCalled();
      expect(event.value).toBe(null);
    });

    it('should expose store event state reactively', () => {
      const mockEvent = { event_id: '1', title: 'Event 1', slug: 'event-1' };
      const slugRef = ref('event-1');
      
      // Pre-populate mock store state
      mockStore.event.value = mockEvent;
      
      const { event, loading, error } = useEvent(slugRef);

      expect(event.value).toEqual(mockEvent);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });
  });

  describe('useFeaturedEvents', () => {
    it('should expose featured events from store', () => {
      const mockFeaturedEvents = [
        {
          event_id: '789',
          title: 'Featured Medical Seminar',
          publishing_status: 'published',
          featured: true
        }
      ];

      // Pre-populate mock store state
      mockStore.featuredEvents.value = mockFeaturedEvents;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { events, loading, error } = useFeaturedEvents();

      expect(events.value).toEqual(mockFeaturedEvents);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockStore.fetchFeaturedEvents).toHaveBeenCalled();
    });

    it('should call store fetchFeaturedEvents with limit parameter', () => {
      useFeaturedEvents(5);

      expect(mockStore.fetchFeaturedEvents).toHaveBeenCalledWith(5);
    });
  });

  describe('useEventCategories', () => {
    it('should expose categories from store', () => {
      const mockCategories = [
        {
          category_id: '456',
          name: 'Medical Events',
          slug: 'medical-events',
          order_number: 1,
          is_default_unassigned: false
        }
      ];

      // Pre-populate mock store state
      mockStore.categories.value = mockCategories;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { categories, loading, error } = useEventCategories();

      expect(categories.value).toEqual(mockCategories);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockStore.fetchEventCategories).toHaveBeenCalled();
    });

    it('should handle empty categories response', () => {
      // Pre-populate mock store state for empty response
      mockStore.categories.value = [];

      const { categories } = useEventCategories();

      expect(categories.value).toEqual([]);
    });
  });

  describe('composable reactivity', () => {
    it('should provide refetch function from useEvents', () => {
      const { refetch } = useEvents({ immediate: false });
      
      expect(typeof refetch).toBe('function');
      
      refetch();
      expect(mockStore.fetchEvents).toHaveBeenCalled();
    });

    it('should provide refetch function from useEvent', () => {
      const { refetch } = useEvent(ref('test-slug'));
      
      expect(typeof refetch).toBe('function');
    });

    it('should handle immediate parameter correctly', () => {
      // When immediate is false, should not call store method on initialization
      mockStore.fetchEvents.mockClear();
      
      useEvents({ immediate: false });
      
      // Should not have been called during initialization
      expect(mockStore.fetchEvents).not.toHaveBeenCalled();
    });
  });
});