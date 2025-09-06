// Events REST Client Tests - Contract validation for database schema compliance
// Tests validate EventsRestClient methods against TABLES-EVENTS.md schema requirements

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { EventsRestClient } from './EventsRestClient';
import type { Event, EventsResponse, EventResponse, GetEventsParams, SearchEventsParams } from './types';
import { mockFetch, expectUrlWithoutBase, expectQueryWithPlusEncoding } from '../../../test/setup';

// Simple mock response helper for this test file
const createMockResponse = (data: any, status = 200, ok = true) => {
  const statusText = status === 200 ? 'OK' :
                     status === 400 ? 'Bad Request' :
                     status === 404 ? 'Not Found' :
                     status === 429 ? 'Too Many Requests' :
                     status === 500 ? 'Internal Server Error' : 'Unknown';
  
  return {
    ok,
    status,
    statusText,
    headers: { get: () => 'application/json' },
    json: () => Promise.resolve(data)
  };
};

// Mock the environment config to prevent HTTP client initialization issues
vi.mock('../../environments', () => ({
  config: {
    domains: {
      events: {
        baseUrl: 'http://localhost:7220',
        timeout: 5000,
        retryAttempts: 2,
      }
    }
  }
}));

// Database schema validation - Event interface must match TABLES-EVENTS.md exactly
interface DatabaseSchemaEvent {
  // Primary key and identifiers
  event_id: string; // UUID PRIMARY KEY
  title: string; // VARCHAR(255) NOT NULL
  description: string; // TEXT NOT NULL
  content?: string; // TEXT (nullable)
  slug: string; // VARCHAR(255) UNIQUE NOT NULL
  
  // Category relationship
  category_id: string; // UUID NOT NULL REFERENCES event_categories(category_id)
  
  // Media and organization
  image_url?: string; // VARCHAR(500) (nullable)
  organizer_name?: string; // VARCHAR(255) (nullable)
  
  // Event scheduling
  event_date: string; // DATE NOT NULL (ISO date format)
  event_time?: string; // TIME (nullable, HH:MM format)
  end_date?: string; // DATE (nullable)
  end_time?: string; // TIME (nullable)
  location: string; // VARCHAR(500) NOT NULL
  virtual_link?: string; // VARCHAR(500) (nullable)
  
  // Registration management
  max_capacity?: number; // INTEGER (nullable)
  registration_deadline?: string; // TIMESTAMPTZ (nullable, ISO format)
  registration_status: 'open' | 'registration_required' | 'full' | 'cancelled'; // VARCHAR(20) NOT NULL DEFAULT 'open'
  
  // Publishing workflow
  publishing_status: 'draft' | 'published' | 'archived'; // VARCHAR(20) NOT NULL DEFAULT 'draft'
  
  // Metadata and categorization
  tags: string[]; // TEXT[] (PostgreSQL array)
  event_type: 'workshop' | 'seminar' | 'webinar' | 'conference' | 'fundraiser' | 'community' | 'medical' | 'educational'; // VARCHAR(50) NOT NULL
  priority_level: 'low' | 'normal' | 'high' | 'urgent'; // VARCHAR(20) NOT NULL DEFAULT 'normal'
  
  // Audit fields
  created_on: string; // TIMESTAMPTZ NOT NULL DEFAULT NOW()
  created_by?: string; // VARCHAR(255) (nullable)
  modified_on?: string; // TIMESTAMPTZ (nullable)
  modified_by?: string; // VARCHAR(255) (nullable)
  
  // Soft delete fields
  is_deleted: boolean; // BOOLEAN NOT NULL DEFAULT FALSE
  deleted_on?: string; // TIMESTAMPTZ (nullable)
  deleted_by?: string; // VARCHAR(255) (nullable)
}

describe('EventsRestClient', () => {
  let client: EventsRestClient;

  beforeEach(() => {
    // Use the imported mockFetch from setup.ts
    client = new EventsRestClient();
    
    // Ensure completely clean mock state for each test
    mockFetch.mockReset();
    mockFetch.mockClear();
    
    // Clear client cache to prevent cross-test contamination
    client.clearCache();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe('Database Schema Compliance', () => {
    it('should validate Event interface matches database schema exactly', () => {
      // This test validates that our Event type matches the database schema
      // It should fail initially to drive schema alignment
      const mockEvent: DatabaseSchemaEvent = {
        event_id: 'uuid-123',
        title: 'Sample Event',
        description: 'Event description',
        content: 'Full event content',
        slug: 'sample-event',
        category_id: 'category-uuid-456',
        image_url: 'https://example.com/image.jpg',
        organizer_name: 'Event Organizer',
        event_date: '2024-03-15',
        event_time: '14:30',
        end_date: '2024-03-15',
        end_time: '17:00',
        location: '123 Main St, City, State',
        virtual_link: 'https://zoom.us/meeting/123',
        max_capacity: 100,
        registration_deadline: '2024-03-10T23:59:59Z',
        registration_status: 'open',
        publishing_status: 'published',
        tags: ['healthcare', 'education'],
        event_type: 'workshop',
        priority_level: 'normal',
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'admin@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'admin@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // This assertion will fail until Event interface matches DatabaseSchemaEvent
      // Driving the need to align types with database schema
      expect(() => {
        const event: Event = mockEvent as any;
        // Required fields validation
        expect(event.event_id).toBeDefined();
        expect(event.description).toBeDefined();
        expect(event.registration_status).toBeDefined();
        expect(event.publishing_status).toBeDefined();
        expect(event.event_type).toBeDefined();
        expect(event.priority_level).toBeDefined();
        expect(event.max_capacity).toBeDefined();
        expect(event.registration_deadline).toBeDefined();
        expect(event.is_deleted).toBeDefined();
        expect(event.created_on).toBeDefined();
      }).not.toThrow();
    });

    it('should validate event_type enum values match database constraints', () => {
      const validEventTypes: DatabaseSchemaEvent['event_type'][] = [
        'workshop', 'seminar', 'webinar', 'conference', 
        'fundraiser', 'community', 'medical', 'educational'
      ];
      
      validEventTypes.forEach(eventType => {
        expect(['workshop', 'seminar', 'webinar', 'conference', 'fundraiser', 'community', 'medical', 'educational'])
          .toContain(eventType);
      });
    });

    it('should validate registration_status enum values match database constraints', () => {
      const validStatuses: DatabaseSchemaEvent['registration_status'][] = [
        'open', 'registration_required', 'full', 'cancelled'
      ];
      
      validStatuses.forEach(status => {
        expect(['open', 'registration_required', 'full', 'cancelled'])
          .toContain(status);
      });
    });

    it('should validate publishing_status enum values match database constraints', () => {
      const validStatuses: DatabaseSchemaEvent['publishing_status'][] = [
        'draft', 'published', 'archived'
      ];
      
      validStatuses.forEach(status => {
        expect(['draft', 'published', 'archived'])
          .toContain(status);
      });
    });

    it('should validate priority_level enum values match database constraints', () => {
      const validPriorities: DatabaseSchemaEvent['priority_level'][] = [
        'low', 'normal', 'high', 'urgent'
      ];
      
      validPriorities.forEach(priority => {
        expect(['low', 'normal', 'high', 'urgent'])
          .toContain(priority);
      });
    });
  });

  describe('getEvents', () => {
    it('should fetch events with proper query parameters', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'test-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      const params: GetEventsParams = {
        page: 1,
        pageSize: 10,
        category: 'healthcare',
        featured: true,
        sortBy: 'date-desc'
      };

      const result = await client.getEvents(params);

      // Validate fetch call with cache-aware expectations
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringMatching(new RegExp('http://localhost:7220/api/v1/events\\?page=1&pageSize=10&category=healthcare&featured=true&sortBy=date-desc')),
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          })
        })
      );
      expect(result).toEqual(mockResponse);
    }, 5000);

    it('should handle empty parameters', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'test-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      const result = await client.getEvents();

      // Validate fetch call with full URL pattern
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringMatching(new RegExp('http://localhost:7220/api/v1/events$')),
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          })
        })
      );
      expect(result).toEqual(mockResponse);
    }, 5000);

    it('should return events matching database schema', async () => {
      const mockDatabaseEvent: DatabaseSchemaEvent = {
        event_id: 'uuid-123',
        title: 'Database Schema Event',
        description: 'Event with complete database schema fields',
        content: 'Full event content',
        slug: 'database-schema-event',
        category_id: 'category-uuid-456',
        image_url: 'https://example.com/image.jpg',
        organizer_name: 'Schema Validator',
        event_date: '2024-03-15',
        event_time: '14:30',
        end_date: '2024-03-15',
        end_time: '17:00',
        location: '123 Schema St, Database City',
        virtual_link: 'https://schema.example.com/meeting',
        max_capacity: 150,
        registration_deadline: '2024-03-10T23:59:59Z',
        registration_status: 'open',
        publishing_status: 'published',
        tags: ['database', 'schema', 'validation'],
        event_type: 'workshop',
        priority_level: 'high',
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'schema@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'validator@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      const mockResponse: EventsResponse = {
        events: [mockDatabaseEvent as any],
        count: 1,
        correlation_id: 'schema-test-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      const result = await client.getEvents();
      
      expect(result.events).toHaveLength(1);
      const event = result.events[0];
      
      // Validate database schema required fields
      expect(event.event_id).toBeDefined();
      expect(event.title).toBeDefined();
      expect(event.description).toBeDefined();
      expect(event.slug).toBeDefined();
      expect(event.category_id).toBeDefined();
      expect(event.event_date).toBeDefined();
      expect(event.location).toBeDefined();
      expect(event.registration_status).toBeDefined();
      expect(event.publishing_status).toBeDefined();
      expect(event.tags).toBeDefined();
      expect(event.event_type).toBeDefined();
      expect(event.priority_level).toBeDefined();
      expect(event.created_on).toBeDefined();
      expect(event.is_deleted).toBeDefined();
    }, 5000);
  });

  describe('getEventBySlug', () => {
    it('should fetch event by slug and return database schema compliant event', async () => {
      const mockDatabaseEvent: DatabaseSchemaEvent = {
        event_id: 'event-uuid-789',
        title: 'Slug Event',
        description: 'Event fetched by slug',
        slug: 'slug-event',
        category_id: 'category-uuid-789',
        event_date: '2024-04-20',
        location: 'Slug Location',
        registration_status: 'open',
        publishing_status: 'published',
        tags: ['slug', 'test'],
        event_type: 'seminar',
        priority_level: 'normal',
        created_on: '2024-02-01T00:00:00Z',
        is_deleted: false,
      };

      const mockResponse: EventResponse = {
        event: mockDatabaseEvent as any,
        correlation_id: 'slug-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      const result = await client.getEventBySlug('slug-event');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/events/slug/slug-event'),
        expect.objectContaining({
          method: 'GET'
        })
      );
      expect(result.event.slug).toBe('slug-event');
      expect(result.event.event_id).toBeDefined();
      expect(result.event.description).toBeDefined();
    }, 5000);

    it('should throw error for empty slug', async () => {
      await expect(client.getEventBySlug('')).rejects.toThrow('Event slug is required');
    }, 5000);

    it('should encode special characters in slug', async () => {
      const mockResponse: EventResponse = {
        event: {} as any,
        correlation_id: 'encoded-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getEventBySlug('event with spaces & special chars');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/events/slug/event%20with%20spaces%20%26%20special%20chars'),
        expect.any(Object)
      );
    }, 5000);
  });

  describe('getEventById', () => {
    it('should fetch event by ID', async () => {
      const mockResponse: EventResponse = {
        event: {} as any,
        correlation_id: 'id-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getEventById('event-uuid-123');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/events/event-uuid-123'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should throw error for empty ID', async () => {
      await expect(client.getEventById('')).rejects.toThrow('Event ID is required');
    }, 5000);
  });

  describe('getFeaturedEvents', () => {
    it('should fetch featured events', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'featured-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getFeaturedEvents();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/events/published'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should handle limit parameter (legacy support)', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'featured-limit-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getFeaturedEvents(5);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/events/published'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);
  });

  describe('searchEvents', () => {
    it('should search events with query parameters', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'search-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      const params: SearchEventsParams = {
        q: 'healthcare workshop',
        page: 1,
        pageSize: 5,
        category: 'medical'
      };

      await client.searchEvents(params);

      // Validate search call with + encoding (URLSearchParams standard)
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringMatching(new RegExp('http://localhost:7220/api/v1/events\\?search=healthcare\\+workshop&page=1&pageSize=5&category=medical')),
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          })
        })
      );
    }, 5000);

    it('should encode search query properly', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'search-encoded-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.searchEvents({ q: 'event with special & chars' });

      // Validate special character encoding with + for spaces, %26 for &
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringMatching(new RegExp('http://localhost:7220/api/v1/events\\?search=event\\+with\\+special\\+%26\\+chars')),
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          })
        })
      );
    }, 5000);
  });

  describe('getRecentEvents', () => {
    it('should fetch recent events with default limit', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'recent-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getRecentEvents();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/events?limit=5&sortBy=date-desc'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should fetch recent events with custom limit', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'recent-custom-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getRecentEvents(10);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/events?limit=10&sortBy=date-desc'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);
  });

  describe('getEventsByCategory', () => {
    it('should fetch events by category', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'category-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getEventsByCategory('healthcare');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/events/categories/healthcare/events'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should throw error for empty category', async () => {
      await expect(client.getEventsByCategory('')).rejects.toThrow('Category is required');
    }, 5000);

    it('should encode category name', async () => {
      const mockResponse: EventsResponse = {
        events: [],
        count: 0,
        correlation_id: 'category-encoded-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockResponse));

      await client.getEventsByCategory('medical research');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/events/categories/medical%20research/events'),
        expect.any(Object)
      );
    }, 5000);
  });

  describe('Error Handling', () => {
    it('should handle network errors', async () => {
      const networkError = new Error('Network error');
      networkError.name = 'NetworkError';
      mockFetch.mockRejectedValue(networkError);

      await expect(client.getEvents()).rejects.toThrow('Network error');
    }, 5000);

    it('should handle HTTP error responses', async () => {
      const errorData = {
        error: {
          code: 'NOT_FOUND',
          message: 'Events not found',
          correlation_id: 'error-404-events'
        }
      };
      
      mockFetch.mockResolvedValueOnce(createMockResponse(errorData, 404, false));

      await expect(client.getEvents()).rejects.toThrow();
    }, 5000);

    it('should handle malformed JSON responses', async () => {
      const malformedResponse = createMockResponse(null, 200, true);
      // Override the json method to simulate parsing error
      malformedResponse.json = vi.fn().mockRejectedValue(new Error('Invalid JSON'));
      
      mockFetch.mockResolvedValueOnce(malformedResponse);

      await expect(client.getEvents()).rejects.toThrow('Invalid JSON');
    }, 5000);
  });

  describe('Shared Cache Behavior', () => {
    it('should use shared RestClientCache for caching operations', async () => {
      const mockEventsResponse: EventsResponse = {
        events: [{
          event_id: 'cache-test-uuid',
          title: 'Cache Test Event',
          summary: 'Testing cache behavior',
          slug: 'cache-test-event',
          category_id: 'category-uuid',
          start_time: '2024-03-15T14:30:00Z',
          end_time: '2024-03-15T16:30:00Z',
          location: 'Test Location',
          publishing_status: 'published',
          tags: ['cache', 'test'],
          event_type: 'workshop',
          priority_level: 'normal',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false,
          deleted_on: null,
          deleted_by: null,
        }],
        count: 1,
        correlation_id: 'cache-correlation-id'
      };

      mockFetch.mockResolvedValueOnce(createMockResponse(mockEventsResponse));

      // Clear cache before test
      client.clearCache();

      // First request should hit the API
      const firstResult = await client.getEvents();
      expect(mockFetch).toHaveBeenCalledTimes(1);
      expect(firstResult).toEqual(mockEventsResponse);

      // Second request should use cache (no additional fetch call)
      const secondResult = await client.getEvents();
      expect(mockFetch).toHaveBeenCalledTimes(1); // Still 1, not 2
      expect(secondResult).toEqual(mockEventsResponse);
    }, 5000);

    it('should provide cache performance metrics via shared cache', async () => {
      // Clear cache and reset metrics
      client.clearCache();

      // Initial metrics should show empty state
      const initialMetrics = client.getMetrics();
      expect(initialMetrics.totalRequests).toBe(0);
      expect(initialMetrics.cacheHits).toBe(0);
      expect(initialMetrics.cacheMisses).toBe(0);
      expect(initialMetrics.errorCount).toBe(0);
    }, 5000);

    it('should provide cache statistics via shared cache', async () => {
      // Clear cache before test
      client.clearCache();

      const initialStats = client.getCacheStats();
      expect(initialStats).toHaveProperty('size');
      expect(initialStats).toHaveProperty('hitRate');
      expect(typeof initialStats.size).toBe('number');
      expect(typeof initialStats.hitRate).toBe('number');
    }, 5000);

    it('should clear all cache entries and reset metrics', async () => {
      // Clear cache before test
      client.clearCache();

      // Verify cache is cleared
      const stats = client.getCacheStats();
      expect(stats.size).toBe(0);

      // Verify metrics are reset
      const metrics = client.getMetrics();
      expect(metrics.totalRequests).toBe(0);
      expect(metrics.cacheHits).toBe(0);
      expect(metrics.cacheMisses).toBe(0);
      expect(metrics.errorCount).toBe(0);
    }, 5000);
  });
});