import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useEventsStore } from './events';
import type { Event, EventCategory, GetEventsParams, SearchEventsParams } from '../lib/clients/events/types';

// Mock the events client with direct mock functions
vi.mock('../lib/clients', () => ({
  eventsClient: {
    getEvents: vi.fn(),
    getEventBySlug: vi.fn(),
    getFeaturedEvents: vi.fn(),
    searchEvents: vi.fn(),
    getEventCategories: vi.fn(),
  }
}));

// Import the mocked client for direct access in tests
import { eventsClient } from '../lib/clients';

describe('EventsStore', () => {
  beforeEach(() => {
    // Create fresh pinia instance for each test
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Initial State', () => {
    it('should initialize with empty state and default values', () => {
      const store = useEventsStore();
      
      expect(store.events).toEqual([]);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(store.total).toBe(0);
      expect(store.page).toBe(1);
      expect(store.pageSize).toBe(10);
      expect(store.categories).toEqual([]);
      expect(store.featuredEvents).toEqual([]);
      expect(store.searchResults).toEqual([]);
      expect(store.upcomingEvents).toEqual([]);
      expect(store.pastEvents).toEqual([]);
    });

    it('should provide totalPages getter based on total and pageSize', () => {
      const store = useEventsStore();
      
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
      const store = useEventsStore();
      
      expect(store.loading).toBe(false);
      
      // Should be able to set loading state
      store.setLoading(true);
      expect(store.loading).toBe(true);
      
      store.setLoading(false);
      expect(store.loading).toBe(false);
    });

    it('should manage error state with proper clearing', () => {
      const store = useEventsStore();
      
      expect(store.error).toBeNull();
      
      // Should set error
      store.setError('Network error occurred');
      expect(store.error).toBe('Network error occurred');
      
      // Should clear error
      store.clearError();
      expect(store.error).toBeNull();
    });

    it('should update events data and pagination state', () => {
      const store = useEventsStore();
      const mockEvents: Event[] = [
        {
          event_id: 'event-1',
          title: 'Test Event 1',
          description: 'Event Description',
          slug: 'test-event-1',
          category_id: 'cat-1',
          event_date: '2024-06-01',
          location: 'Test Location',
          registration_status: 'open',
          publishing_status: 'published',
          event_type: 'workshop',
          priority_level: 'normal',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.setEvents(mockEvents, 25, 2, 10);
      
      expect(store.events).toEqual(mockEvents);
      expect(store.total).toBe(25);
      expect(store.page).toBe(2);
      expect(store.pageSize).toBe(10);
    });
  });

  describe('Actions - Events Operations', () => {
    it('should fetch events with parameters and update state accordingly', async () => {
      const store = useEventsStore();
      
      const mockResponse = {
        events: [
          {
            event_id: 'event-1',
            title: 'Test Event',
            description: 'Test Description',
            slug: 'test-event',
            category_id: 'cat-1',
            event_date: '2024-06-01',
            location: 'Test Location',
            registration_status: 'open' as const,
            publishing_status: 'published' as const,
            event_type: 'seminar' as const,
            priority_level: 'high' as const,
            created_on: '2024-01-01T00:00:00Z',
            is_deleted: false
          }
        ],
        count: 1,
        correlation_id: 'test-correlation-id'
      };
      
      const { eventsClient } = await import('../lib/clients');
      (eventsClient.getEvents as any).mockResolvedValueOnce(mockResponse);
      
      const params: GetEventsParams = { page: 1, pageSize: 10 };
      await store.fetchEvents(params);
      
      expect(eventsClient.getEvents).toHaveBeenCalledWith(params);
      expect(store.events).toEqual(mockResponse.events);
      expect(store.total).toBe(mockResponse.count);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    it('should handle events fetch errors properly', async () => {
      const store = useEventsStore();
      
      const errorMessage = 'Failed to fetch events';
      (eventsClient.getEvents as any).mockRejectedValueOnce(new Error(errorMessage));
      
      await store.fetchEvents();
      
      expect(store.error).toBe(errorMessage);
      expect(store.loading).toBe(false);
      expect(store.events).toEqual([]);
    });

    it('should fetch single event by slug', async () => {
      const store = useEventsStore();
      
      const mockEvent: Event = {
        event_id: 'event-1',
        title: 'Single Event',
        description: 'Event Description',
        slug: 'single-event',
        category_id: 'cat-1',
        event_date: '2024-06-01',
        event_time: '14:00:00',
        location: 'Main Conference Hall',
        registration_status: 'open',
        publishing_status: 'published',
        event_type: 'conference',
        priority_level: 'urgent',
        created_on: '2024-01-01T00:00:00Z',
        is_deleted: false,
        content: 'Full event content'
      };
      
      const mockResponse = {
        event: mockEvent,
        correlation_id: 'event-correlation-id'
      };
      
      (eventsClient.getEventBySlug as any).mockResolvedValueOnce(mockResponse);
      
      const result = await store.fetchEvent('single-event');
      
      expect(eventsClient.getEventBySlug).toHaveBeenCalledWith('single-event');
      expect(result).toEqual(mockEvent);
    });
  });

  describe('Actions - Featured Events', () => {
    it('should fetch featured events and cache in store', async () => {
      const store = useEventsStore();
      
      const mockFeaturedEvents = [
        {
          event_id: 'featured-1',
          title: 'Featured Event',
          description: 'Featured Description',
          slug: 'featured-event',
          category_id: 'cat-1',
          event_date: '2024-07-01',
          location: 'Featured Venue',
          registration_status: 'open' as const,
          publishing_status: 'published' as const,
          event_type: 'fundraiser' as const,
          priority_level: 'urgent' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      const mockResponse = {
        events: mockFeaturedEvents,
        count: 1,
        correlation_id: 'featured-correlation-id'
      };
      
      const { eventsClient } = await import('../lib/clients');
      (eventsClient.getFeaturedEvents as any).mockResolvedValueOnce(mockResponse);
      
      await store.fetchFeaturedEvents(5);
      
      expect(eventsClient.getFeaturedEvents).toHaveBeenCalledWith(5);
      expect(store.featuredEvents).toEqual(mockFeaturedEvents);
    });
  });

  describe('Actions - Search Operations', () => {
    it('should perform events search and store results separately', async () => {
      const store = useEventsStore();
      
      const mockSearchResults = [
        {
          event_id: 'search-1',
          title: 'Search Result',
          description: 'Search Description',
          slug: 'search-result',
          category_id: 'cat-1',
          event_date: '2024-08-01',
          location: 'Search Location',
          registration_status: 'open' as const,
          publishing_status: 'published' as const,
          event_type: 'webinar' as const,
          priority_level: 'normal' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      const mockResponse = {
        events: mockSearchResults,
        count: 1,
        correlation_id: 'search-correlation-id'
      };
      
      const { eventsClient } = await import('../lib/clients');
      (eventsClient.searchEvents as any).mockResolvedValueOnce(mockResponse);
      
      const searchParams: SearchEventsParams = {
        q: 'test query',
        page: 1,
        pageSize: 10
      };
      
      await store.searchEvents(searchParams);
      
      expect(eventsClient.searchEvents).toHaveBeenCalledWith(searchParams);
      expect(store.searchResults).toEqual(mockSearchResults);
      expect(store.searchTotal).toBe(1);
    });

    it('should clear search results when query is empty', async () => {
      const store = useEventsStore();
      
      // Set some initial search results
      store.$patch({
        searchResults: [{ 
          event_id: 'test',
          title: 'Test',
          description: 'Test',
          slug: 'test',
          category_id: 'cat-1',
          event_date: '2024-06-01',
          location: 'Test Location',
          registration_status: 'open' as const,
          publishing_status: 'published' as const,
          event_type: 'workshop' as const,
          priority_level: 'normal' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }],
        searchTotal: 1
      });
      
      await store.searchEvents({ q: '', page: 1, pageSize: 10 });
      
      expect(store.searchResults).toEqual([]);
      expect(store.searchTotal).toBe(0);
    });
  });

  describe('Actions - Categories', () => {
    it('should fetch and cache event categories', async () => {
      const store = useEventsStore();
      
      const mockCategories: EventCategory[] = [
        {
          category_id: 'cat-1',
          name: 'Medical Events',
          slug: 'medical-events',
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
      
      (eventsClient.getEventCategories as any).mockResolvedValueOnce(mockResponse);
      
      await store.fetchEventCategories();
      
      expect(eventsClient.getEventCategories).toHaveBeenCalled();
      expect(store.categories).toEqual(mockCategories);
    });
  });

  describe('Getters', () => {
    it('should provide computed values for pagination', () => {
      const store = useEventsStore();
      
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

    it('should provide hasEvents getter for conditional rendering', () => {
      const store = useEventsStore();
      
      expect(store.hasEvents).toBe(false);
      
      store.$patch({
        events: [{
          event_id: 'test',
          title: 'Test',
          description: 'Test',
          slug: 'test',
          category_id: 'cat-1',
          event_date: '2024-06-01',
          location: 'Test Location',
          registration_status: 'open',
          publishing_status: 'published',
          event_type: 'workshop',
          priority_level: 'normal',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }]
      });
      
      expect(store.hasEvents).toBe(true);
    });

    it('should provide events grouped by date (upcoming vs past)', () => {
      const store = useEventsStore();
      const today = new Date().toISOString().split('T')[0];
      const futureDate = '2025-06-01';
      const pastDate = '2023-06-01';
      
      const mockEvents = [
        {
          event_id: 'upcoming-1',
          title: 'Upcoming Event',
          description: 'Future event',
          slug: 'upcoming-event',
          category_id: 'cat-1',
          event_date: futureDate,
          location: 'Future Location',
          registration_status: 'open' as const,
          publishing_status: 'published' as const,
          event_type: 'workshop' as const,
          priority_level: 'normal' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        },
        {
          event_id: 'past-1',
          title: 'Past Event',
          description: 'Past event',
          slug: 'past-event',
          category_id: 'cat-1',
          event_date: pastDate,
          location: 'Past Location',
          registration_status: 'open' as const,
          publishing_status: 'published' as const,
          event_type: 'seminar' as const,
          priority_level: 'normal' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];

      store.$patch({ events: mockEvents });
      
      // Should separate upcoming and past events
      expect(store.upcomingEvents).toHaveLength(1);
      expect(store.pastEvents).toHaveLength(1);
      expect(store.upcomingEvents[0].event_id).toBe('upcoming-1');
      expect(store.pastEvents[0].event_id).toBe('past-1');
    });

    it('should provide events grouped by category', () => {
      const store = useEventsStore();
      
      const mockEvents = [
        {
          event_id: 'event-1',
          title: 'Event 1',
          description: 'Description 1',
          slug: 'event-1',
          category_id: 'cat-1',
          event_date: '2024-06-01',
          location: 'Location 1',
          registration_status: 'open' as const,
          publishing_status: 'published' as const,
          event_type: 'workshop' as const,
          priority_level: 'normal' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        },
        {
          event_id: 'event-2',
          title: 'Event 2',
          description: 'Description 2',
          slug: 'event-2',
          category_id: 'cat-2',
          event_date: '2024-07-01',
          location: 'Location 2',
          registration_status: 'open' as const,
          publishing_status: 'published' as const,
          event_type: 'seminar' as const,
          priority_level: 'high' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];

      store.$patch({ events: mockEvents });
      
      const grouped = store.eventsByCategory;
      expect(grouped).toHaveProperty('cat-1');
      expect(grouped).toHaveProperty('cat-2');
      expect(grouped['cat-1']).toHaveLength(1);
      expect(grouped['cat-2']).toHaveLength(1);
    });
  });

  describe('Event Registration Management', () => {
    it('should track event registration counts', () => {
      const store = useEventsStore();
      
      const eventWithRegistrations = {
        event_id: 'event-1',
        title: 'Event with Registrations',
        description: 'Event Description',
        slug: 'event-with-registrations',
        category_id: 'cat-1',
        event_date: '2024-06-01',
        location: 'Event Location',
        registration_status: 'open' as const,
        publishing_status: 'published' as const,
        event_type: 'workshop' as const,
        priority_level: 'normal' as const,
        max_capacity: 100,
        created_on: '2024-01-01T00:00:00Z',
        is_deleted: false
      };

      store.$patch({ events: [eventWithRegistrations] });
      
      // Should provide registration availability info
      expect(store.events[0].max_capacity).toBe(100);
      expect(store.events[0].registration_status).toBe('open');
    });
  });

  describe('Cache Management', () => {
    it('should cache events data and avoid duplicate fetches', async () => {
      const store = useEventsStore();
      
      const mockResponse = {
        events: [{
          event_id: 'cached-1',
          title: 'Cached Event',
          description: 'Cached Description',
          slug: 'cached-event',
          category_id: 'cat-1',
          event_date: '2024-06-01',
          location: 'Cached Location',
          registration_status: 'open' as const,
          publishing_status: 'published' as const,
          event_type: 'workshop' as const,
          priority_level: 'normal' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }],
        count: 1,
        correlation_id: 'cache-test-id'
      };
      
      const { eventsClient } = await import('../lib/clients');
      (eventsClient.getEvents as any).mockResolvedValueOnce(mockResponse);
      
      // First fetch should call API
      await store.fetchEvents({ page: 1, pageSize: 10 });
      expect(eventsClient.getEvents).toHaveBeenCalledTimes(1);
      
      // Second fetch with same params should use cache
      await store.fetchEvents({ page: 1, pageSize: 10 }, { useCache: true });
      expect(eventsClient.getEvents).toHaveBeenCalledTimes(1); // Still only called once
    });

    it('should invalidate cache and refetch when requested', async () => {
      const store = useEventsStore();
      
      const mockResponse = {
        events: [],
        count: 0,
        correlation_id: 'invalidate-test-id'
      };
      
      const { eventsClient } = await import('../lib/clients');
      (eventsClient.getEvents as any).mockResolvedValue(mockResponse);
      
      // First fetch
      await store.fetchEvents({ page: 1, pageSize: 10 });
      expect(eventsClient.getEvents).toHaveBeenCalledTimes(1);
      
      // Invalidate cache and fetch again
      store.invalidateCache();
      await store.fetchEvents({ page: 1, pageSize: 10 });
      expect(eventsClient.getEvents).toHaveBeenCalledTimes(2);
    });
  });
});