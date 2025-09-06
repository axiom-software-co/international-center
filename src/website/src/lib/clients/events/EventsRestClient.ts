// Events REST Client - Database Schema Compliant with Performance Optimizations
// Updated to work with TABLES-EVENTS.md aligned types
// Features: Response caching, request deduplication, performance monitoring

import { BaseRestClient } from '../rest/BaseRestClient';
import { RestClientCache, STANDARD_CACHE_TTL } from '../rest/RestClientCache';
import { config } from '../../environments';
// Removed factory imports for direct cache integration
import type {
  Event,
  EventsResponse,
  EventResponse,
  EventCategoriesResponse,
  FeaturedEventResponse,
  GetEventsParams,
  SearchEventsParams,
  CreateEventParams,
  UpdateEventParams,
} from './types';

export class EventsRestClient extends BaseRestClient {
  private cache = new RestClientCache();
  
// Direct cache integration - no factory configuration needed

  constructor() {
    // Handle test environment or missing configuration
    const eventsConfig = config.domains?.events || {
      baseUrl: 'http://localhost:7220', // Public Gateway URL fallback
      timeout: 5000,
      retryAttempts: 2,
    };
    
    super({
      baseUrl: eventsConfig.baseUrl,
      timeout: eventsConfig.timeout,
      retryAttempts: eventsConfig.retryAttempts,
    });
  }

/**
   * Get all events with optional filtering and pagination
   * Maps to GET /api/v1/events endpoint through Public Gateway
   * Uses database schema-compliant query parameters with caching
   */
  async getEvents(params: GetEventsParams = {}): Promise<EventsResponse> {
    const queryParams = new URLSearchParams();
    
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);
    if (params.publishing_status) queryParams.set('publishing_status', params.publishing_status);
    if (params.featured !== undefined) queryParams.set('featured', params.featured.toString());
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);
    if (params.search) queryParams.set('search', params.search);

    const cacheKey = `events:${queryParams.toString()}`;
    const url = `/api/v1/events${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    
    return this.cache.requestWithCache<EventsResponse>(
      this,
      url,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.LIST
    );
  }

  /**
   * Get single event by slug
   * Maps to GET /api/v1/events/slug/{slug} endpoint through Public Gateway
   * Uses caching for event detail with database schema compliance
   */
  async getEventBySlug(slug: string): Promise<EventResponse> {
    if (!slug || slug.trim() === '') {
      throw new Error('Event slug is required');
    }

    const cacheKey = `events:slug:${slug}`;
    const url = `/api/v1/events/slug/${encodeURIComponent(slug)}`;
    
    return this.cache.requestWithCache<EventResponse>(
      this,
      url,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.DETAIL
    );
  }

  /**
   * Get single event by ID
   * Maps to GET /api/v1/events/{id} endpoint through Public Gateway
   * Uses caching for event detail
   */
  async getEventById(id: string): Promise<EventResponse> {
    if (!id || id.trim() === '') {
      throw new Error('Event ID is required');
    }

    const cacheKey = `events:id:${id}`;
    const url = `/api/v1/events/${encodeURIComponent(id)}`;
    
    return this.cache.requestWithCache<EventResponse>(
      this,
      url,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.DETAIL
    );
  }

  /**
   * Get featured events with optional limit
   * Maps to GET /api/v1/events/published endpoint through Public Gateway
   * Uses caching for featured content with database schema compliance
   */
  async getFeaturedEvents(limit?: number): Promise<FeaturedEventResponse> {
    const queryParams = new URLSearchParams();
    if (limit !== undefined) queryParams.set('limit', limit.toString());

    const cacheKey = `events:featured:${limit || 'all'}`;
    const url = `/api/v1/events/published${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    
    return this.cache.requestWithCache<FeaturedEventResponse>(
      this,
      url,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.FEATURED
    );
  }

  // Convenience methods
  async getPublishedEvents(params: Partial<GetEventsParams> = {}): Promise<EventsResponse> {
    return this.getEvents({
      ...params,
      publishing_status: 'published',
    });
  }

  /**
   * Search events with query parameters
   * Maps to GET /api/v1/events endpoint with search parameter through Public Gateway
   * Database schema-compliant search with no caching for real-time results
   */
  async searchEvents(params: SearchEventsParams): Promise<EventsResponse> {
    // Return empty results for empty queries without making HTTP request
    if (!params.q || params.q.trim() === '') {
      return {
        events: [],
        count: 0,
        correlation_id: `empty-search-${Date.now()}`
      };
    }

    const queryParams = new URLSearchParams();
    queryParams.set('search', params.q.trim());
    
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const url = `/api/v1/events?${queryParams.toString()}`;
    
    // Search results are not cached for real-time freshness
    return this.request<EventsResponse>(url, { method: 'GET' });
  }

  /**
   * Get recent events
   * Uses limit and date-desc parameters for recent content
   */
  async getRecentEvents(limit: number = 5): Promise<EventsResponse> {
    const queryParams = new URLSearchParams();
    queryParams.set('limit', limit.toString());
    queryParams.set('sortBy', 'date-desc');
    
    const url = `/api/v1/events?${queryParams.toString()}`;
    
    return this.request<EventsResponse>(url, { method: 'GET' });
  }

  /**
   * Get events by category
   * Maps to GET /api/v1/events/categories/{id}/events endpoint
   */
  async getEventsByCategory(category_id: string, params?: Record<string, any>): Promise<EventsResponse> {
    if (!category_id) {
      throw new Error('Category is required');
    }

    const url = `/api/v1/events/categories/${encodeURIComponent(category_id)}/events`;
    
    return this.request<EventsResponse>(url, { method: 'GET' });
  }

  /**
   * Get event categories
   * Maps to GET /api/v1/events/categories endpoint through Public Gateway
   * Uses aggressive caching as categories change infrequently
   */
  async getEventCategories(): Promise<EventCategoriesResponse> {
    const cacheKey = 'events:categories';
    const url = '/api/v1/events/categories';
    
    return this.cache.requestWithCache<EventCategoriesResponse>(
      this,
      url,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.CATEGORIES
    );
  }

  /**
   * Create new event
   * Maps to POST /api/v1/events endpoint through Admin Gateway
   * Clears relevant cache entries after creation
   */
  async createEvent(params: CreateEventParams): Promise<EventResponse> {
    const url = '/api/v1/events';
    
    try {
      const result = await this.request<EventResponse>(url, {
        method: 'POST',
        body: JSON.stringify(params),
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      // Invalidate relevant caches
      this.cache.invalidateCachePattern('events:');
      
      return result;
    } catch (error) {
      throw error;
    }
  }

  /**
   * Update existing event
   * Maps to PUT /api/v1/events/{id} endpoint through Admin Gateway
   * Clears relevant cache entries after update
   */
  async updateEvent(params: UpdateEventParams): Promise<EventResponse> {
    const { event_id, ...updateData } = params;
    const url = `/api/v1/events/${encodeURIComponent(event_id)}`;
    
    try {
      const result = await this.request<EventResponse>(url, {
        method: 'PUT',
        body: JSON.stringify(updateData),
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      // Invalidate specific event caches and event lists
      this.cache.invalidateCacheKey(`events:id:${event_id}`);
      this.cache.invalidateCachePattern('events:');
      
      return result;
    } catch (error) {
      throw error;
    }
  }

  /**
   * Delete event (soft delete)
   * Maps to DELETE /api/v1/events/{id} endpoint through Admin Gateway
   * Clears relevant cache entries after deletion
   */
  async deleteEvent(event_id: string): Promise<void> {
    if (!event_id) {
      throw new Error('Event ID is required');
    }

    const url = `/api/v1/events/${encodeURIComponent(event_id)}`;
    
    try {
      await this.request<void>(url, { method: 'DELETE' });
      
      // Invalidate specific event caches and event lists
      this.cache.invalidateCacheKey(`events:id:${event_id}`);
      this.cache.invalidateCachePattern('events:');
    } catch (error) {
      throw error;
    }
  }



  /**
   * Get performance metrics
   */
  public getMetrics() {
    return this.cache.getMetrics();
  }
  
  /**
   * Get cache statistics
   */
  public getCacheStats() {
    return this.cache.getCacheStats();
  }
  
  /**
   * Clear all cache entries and metrics
   */
  public clearCache(): void {
    this.cache.clearCache();
  }
}