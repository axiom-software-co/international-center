// Events Composables Tests - State management and API integration validation
// Tests validate useEvents composables with database schema-compliant reactive state

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ref, nextTick, defineComponent } from 'vue';
import { mount } from '@vue/test-utils';
import { useEvents, useEvent, useFeaturedEvents, useSearchEvents } from './useEvents';
import type { Event, EventsResponse, EventResponse, GetEventsParams, SearchEventsParams } from '../lib/clients/events/types';

// Mock the EventsRestClient with hoisted functions
const mockGetEvents = vi.fn();
const mockGetEventBySlug = vi.fn();
const mockGetFeaturedEvents = vi.fn();
const mockSearchEvents = vi.fn();

const MockedEventsRestClient = vi.fn().mockImplementation(() => ({
  getEvents: mockGetEvents,
  getEventBySlug: mockGetEventBySlug,
  getFeaturedEvents: mockGetFeaturedEvents,
  searchEvents: mockSearchEvents,
}));

vi.mock('../lib/clients/events/EventsRestClient', () => ({
  EventsRestClient: MockedEventsRestClient
}));

// Database schema-compliant mock event for testing
const createMockDatabaseEvent = (overrides: Partial<any> = {}): any => ({
  event_id: 'event-uuid-123',
  title: 'Mock Database Event',
  description: 'Event description from database',
  content: 'Full event content',
  slug: 'mock-database-event',
  category_id: 'category-uuid-456',
  image_url: 'https://example.com/event-image.jpg',
  organizer_name: 'Event Organizer',
  event_date: '2024-03-15',
  event_time: '14:30',
  end_date: '2024-03-15',
  end_time: '17:00',
  location: '123 Database St, Schema City',
  virtual_link: 'https://virtual.example.com/event',
  max_capacity: 100,
  registration_deadline: '2024-03-10T23:59:59Z',
  registration_status: 'open' as const,
  publishing_status: 'published' as const,
  tags: ['database', 'schema', 'event'],
  event_type: 'workshop' as const,
  priority_level: 'normal' as const,
  created_on: '2024-01-01T00:00:00Z',
  created_by: 'admin@example.com',
  modified_on: '2024-01-02T00:00:00Z',
  modified_by: 'admin@example.com',
  is_deleted: false,
  deleted_on: null,
  deleted_by: null,
  ...overrides,
});

describe('useEvents Composables', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('useEvents', () => {
    it('should initialize with proper default state', () => {
      mockGetEvents.mockResolvedValue({
        events: [],
        count: 0,
        correlation_id: 'test-correlation-id'
      });

      const { events, loading, error, total, page, pageSize, totalPages } = useEvents({
        enabled: false,
        immediate: false
      });

      expect(events.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      expect(totalPages.value).toBe(0);
    }, 5000);

    it('should fetch events with database schema-compliant data', async () => {
      const mockDatabaseEvents = [
        createMockDatabaseEvent(),
        createMockDatabaseEvent({
          event_id: 'event-uuid-124',
          title: 'Second Database Event',
          slug: 'second-database-event',
          event_type: 'seminar' as const,
          priority_level: 'high' as const,
        })
      ];

      const mockResponse: EventsResponse = {
        events: mockDatabaseEvents,
        count: 2,
        correlation_id: 'events-correlation-id'
      };

      mockGetEvents.mockResolvedValue(mockResponse);

      const { events, loading, error, total, totalPages, refetch } = useEvents({
        enabled: false,
        immediate: false
      });

      expect(loading.value).toBe(false);

      await refetch();

      await nextTick();

      expect(mockGetEvents).toHaveBeenCalledTimes(1);
      expect(events.value).toHaveLength(2);
      expect(total.value).toBe(2);
      expect(totalPages.value).toBe(1);
      expect(error.value).toBe(null);
      expect(loading.value).toBe(false);

      // Validate database schema fields are present
      const firstEvent = events.value[0];
      expect(firstEvent.event_id).toBeDefined();
      expect(firstEvent.description).toBeDefined();
      expect(firstEvent.registration_status).toBeDefined();
      expect(firstEvent.publishing_status).toBeDefined();
      expect(firstEvent.event_type).toBeDefined();
      expect(firstEvent.priority_level).toBeDefined();
      expect(firstEvent.max_capacity).toBeDefined();
      expect(firstEvent.is_deleted).toBeDefined();
      expect(firstEvent.created_on).toBeDefined();
    }, 5000);

    it('should handle API errors gracefully', async () => {
      mockGetEvents.mockRejectedValue(new Error('API Error'));

      const { events, loading, error, refetch } = useEvents({
        enabled: false,
        immediate: false
      });

      await refetch();
      await nextTick();

      expect(error.value).toBe('API Error');
      expect(events.value).toEqual([]);
      expect(loading.value).toBe(false);
    }, 5000);

    it('should handle query parameters correctly', async () => {
      mockGetEvents.mockResolvedValue({
        events: [],
        count: 0,
        correlation_id: 'params-correlation-id'
      });

      const params: GetEventsParams = {
        page: 2,
        pageSize: 20,
        category: 'healthcare',
        featured: true,
        sortBy: 'date-desc'
      };

      const { refetch } = useEvents({
        enabled: false,
        immediate: false,
        ...params
      });

      await refetch();

      expect(mockGetEvents).toHaveBeenCalledWith(
        expect.objectContaining(params)
      );
    }, 5000);

    it('should handle pagination calculations correctly', async () => {
      mockGetEvents.mockResolvedValue({
        events: Array(15).fill(null).map((_, i) => createMockDatabaseEvent({
          event_id: `event-${i}`,
          title: `Event ${i}`,
          slug: `event-${i}`
        })),
        count: 150,
        correlation_id: 'pagination-correlation-id'
      });

      const { total, pageSize, totalPages, refetch } = useEvents({
        enabled: false,
        immediate: false,
        pageSize: 15
      });

      await refetch();
      await nextTick();

      expect(total.value).toBe(150);
      expect(pageSize.value).toBe(15);
      expect(totalPages.value).toBe(10); // 150 / 15 = 10
    }, 5000);
  });

  describe('useEvent', () => {
    it('should fetch single event by slug', async () => {
      const mockEvent = createMockDatabaseEvent({
        slug: 'single-event-test'
      });

      const mockResponse: EventResponse = {
        event: mockEvent,
        correlation_id: 'single-event-correlation-id'
      };

      mockGetEventBySlug.mockResolvedValue(mockResponse);

      const slugRef = ref('single-event-test');
      const { event, loading, error } = useEvent(slugRef);

      await nextTick();

      expect(mockGetEventBySlug).toHaveBeenCalledWith('single-event-test');
      expect(event.value).toEqual(mockEvent);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);

      // Validate database schema compliance
      expect(event.value?.event_id).toBeDefined();
      expect(event.value?.description).toBeDefined();
      expect(event.value?.registration_status).toBeDefined();
    }, 5000);

    it('should handle slug changes reactively', async () => {
      mockGetEventBySlug.mockResolvedValue({
        event: createMockDatabaseEvent(),
        correlation_id: 'reactive-correlation-id'
      });

      const slugRef = ref('initial-slug');
      const { refetch } = useEvent(slugRef);

      await nextTick();

      expect(mockGetEventBySlug).toHaveBeenCalledWith('initial-slug');

      // Change slug
      slugRef.value = 'updated-slug';
      await nextTick();

      expect(mockGetEventBySlug).toHaveBeenCalledWith('updated-slug');
      expect(mockGetEventBySlug).toHaveBeenCalledTimes(2);
    }, 5000);

    it('should handle empty slug gracefully', async () => {

      const { event, loading } = useEvent(ref(null));

      await nextTick();

      expect(mockGetEventBySlug).not.toHaveBeenCalled();
      expect(event.value).toBe(null);
      expect(loading.value).toBe(false);
    }, 5000);

    it('should handle API errors', async () => {
      mockGetEventBySlug.mockRejectedValue(new Error('Event not found'));

      const { event, error, loading } = useEvent('non-existent-slug');

      await nextTick();

      expect(error.value).toBe('Event not found');
      expect(event.value).toBe(null);
      expect(loading.value).toBe(false);
    }, 5000);
  });

  describe('useFeaturedEvents', () => {
    it('should fetch featured events', async () => {
      const mockFeaturedEvents = [
        createMockDatabaseEvent({ title: 'Featured Event 1' }),
        createMockDatabaseEvent({ 
          event_id: 'event-uuid-125',
          title: 'Featured Event 2',
          slug: 'featured-event-2',
          priority_level: 'high' as const
        })
      ];

      const mockResponse: EventsResponse = {
        events: mockFeaturedEvents,
        count: 2,
        correlation_id: 'featured-correlation-id'
      };

      mockGetFeaturedEvents.mockResolvedValue(mockResponse);

      const { events, loading, error } = useFeaturedEvents();

      await nextTick();

      expect(mockGetFeaturedEvents).toHaveBeenCalledWith(undefined);
      expect(events.value).toHaveLength(2);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);

      // Validate database schema compliance
      expect(events.value[0].event_id).toBeDefined();
      expect(events.value[0].publishing_status).toBeDefined();
    }, 5000);

    it('should handle limit parameter', async () => {
      mockGetFeaturedEvents.mockResolvedValue({
        events: [],
        count: 0,
        correlation_id: 'featured-limit-correlation-id'
      });

      const limitRef = ref(5);
      useFeaturedEvents(limitRef);

      await nextTick();

      expect(mockGetFeaturedEvents).toHaveBeenCalledWith(5);
    }, 5000);

    it('should handle limit changes reactively', async () => {
      mockGetFeaturedEvents.mockResolvedValue({
        events: [],
        count: 0,
        correlation_id: 'featured-reactive-correlation-id'
      });

      const limitRef = ref(3);
      useFeaturedEvents(limitRef);

      await nextTick();
      expect(mockGetFeaturedEvents).toHaveBeenCalledWith(3);

      limitRef.value = 7;
      await nextTick();
      expect(mockGetFeaturedEvents).toHaveBeenCalledWith(7);
      expect(mockGetFeaturedEvents).toHaveBeenCalledTimes(2);
    }, 5000);
  });

  describe('useSearchEvents', () => {
    it('should search events with query', async () => {
      const mockSearchResults = [
        createMockDatabaseEvent({ title: 'Search Result 1' }),
        createMockDatabaseEvent({
          event_id: 'event-uuid-126',
          title: 'Search Result 2',
          slug: 'search-result-2',
          event_type: 'conference' as const
        })
      ];

      const mockResponse: EventsResponse = {
        events: mockSearchResults,
        count: 2,
        correlation_id: 'search-correlation-id'
      };

      mockSearchEvents.mockResolvedValue(mockResponse);

      const { results, loading, error, total, search } = useSearchEvents();

      await search('healthcare workshop', {
        page: 1,
        pageSize: 10,
        category: 'medical'
      });

      expect(mockSearchEvents).toHaveBeenCalledWith({
        q: 'healthcare workshop',
        page: 1,
        pageSize: 10,
        category: 'medical'
      });
      expect(results.value).toHaveLength(2);
      expect(total.value).toBe(2);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);

      // Validate database schema compliance
      expect(results.value[0].event_id).toBeDefined();
      expect(results.value[0].event_type).toBeDefined();
    }, 5000);

    it('should handle empty search queries', async () => {
      const { results, total, totalPages, search } = useSearchEvents();

      await search('');

      expect(results.value).toEqual([]);
      expect(total.value).toBe(0);
      expect(totalPages.value).toBe(0);
    }, 5000);

    it('should handle search errors', async () => {
      mockSearchEvents.mockRejectedValue(new Error('Search failed'));

      const { results, error, loading, search } = useSearchEvents();

      await search('test query');

      expect(error.value).toBe('Search failed');
      expect(results.value).toEqual([]);
      expect(loading.value).toBe(false);
    }, 5000);

    it('should calculate pagination correctly for search results', async () => {
      mockSearchEvents.mockResolvedValue({
        events: Array(5).fill(null).map((_, i) => createMockDatabaseEvent({
          event_id: `search-result-${i}`,
          title: `Search Result ${i}`,
          slug: `search-result-${i}`
        })),
        count: 50,
        correlation_id: 'search-pagination-correlation-id'
      });

      const { total, page, pageSize, totalPages, search } = useSearchEvents();

      await search('test query', {
        page: 2,
        pageSize: 5
      });

      expect(total.value).toBe(50);
      expect(page.value).toBe(2);
      expect(pageSize.value).toBe(5);
      expect(totalPages.value).toBe(10); // 50 / 5 = 10
    }, 5000);

    it('should handle search options correctly', async () => {
      mockSearchEvents.mockResolvedValue({
        events: [],
        count: 0,
        correlation_id: 'search-options-correlation-id'
      });

      const { search } = useSearchEvents();

      const searchOptions: Partial<SearchEventsParams> = {
        page: 3,
        pageSize: 25,
        category: 'healthcare',
        sortBy: 'date-desc'
      };

      await search('medical conference', searchOptions);

      expect(mockSearchEvents).toHaveBeenCalledWith({
        q: 'medical conference',
        page: 3,
        pageSize: 25,
        category: 'healthcare',
        sortBy: 'date-desc'
      });
    }, 5000);
  });

  describe('Reactive State Management', () => {
    it('should maintain proper loading states during transitions', async () => {
      
      // Simulate slow API call
      let resolvePromise: (value: EventsResponse) => void;
      const slowPromise = new Promise<EventsResponse>((resolve) => {
        resolvePromise = resolve;
      });
      mockGetEvents.mockReturnValue(slowPromise);

      const { loading, refetch } = useEvents({
        enabled: false,
        immediate: false
      });

      expect(loading.value).toBe(false);

      const fetchPromise = refetch();
      
      // Should be loading during fetch
      expect(loading.value).toBe(true);

      // Resolve the promise
      resolvePromise!({
        events: [],
        count: 0,
        correlation_id: 'loading-test-correlation-id'
      });

      await fetchPromise;
      await nextTick();

      // Should not be loading after fetch completes
      expect(loading.value).toBe(false);
    }, 5000);

    it('should properly clear errors when making new requests', async () => {
      
      // First call fails
      mockGetEvents.mockRejectedValueOnce(new Error('First error'));
      
      const { error, refetch } = useEvents({
        enabled: false,
        immediate: false
      });

      await refetch();
      await nextTick();

      expect(error.value).toBe('First error');

      // Second call succeeds
      mockGetEvents.mockResolvedValueOnce({
        events: [],
        count: 0,
        correlation_id: 'error-clear-correlation-id'
      });

      await refetch();
      await nextTick();

      expect(error.value).toBe(null);
    }, 5000);
  });
});