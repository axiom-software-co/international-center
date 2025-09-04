// Events REST Client - Database Schema Compliant with Performance Optimizations
// Updated to work with TABLES-EVENTS.md aligned types
// Features: Response caching, request deduplication, performance monitoring

import { BaseRestClient } from '../rest/BaseRestClient';
import { config } from '../../environments';
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

// Cache interface for response caching
interface CacheEntry<T> {
  data: T;
  timestamp: number;
  ttl: number;
}

// Performance metrics tracking
interface RequestMetrics {
  totalRequests: number;
  cacheHits: number;
  cacheMisses: number;
  averageResponseTime: number;
  errorCount: number;
}

export class EventsRestClient extends BaseRestClient {
  private cache = new Map<string, CacheEntry<any>>();
  private pendingRequests = new Map<string, Promise<any>>();
  private metrics: RequestMetrics = {
    totalRequests: 0,
    cacheHits: 0,
    cacheMisses: 0,
    averageResponseTime: 0,
    errorCount: 0,
  };
  
  // Cache TTL values in milliseconds
  private static readonly CACHE_TTL = {
    CATEGORIES: 15 * 60 * 1000, // 15 minutes - categories change infrequently
    FEATURED: 5 * 60 * 1000, // 5 minutes - featured events change occasionally  
    EVENT_DETAIL: 2 * 60 * 1000, // 2 minutes - individual events may be updated
    EVENT_LIST: 30 * 1000, // 30 seconds - lists change more frequently
  };

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

    // Clear expired cache entries every 5 minutes
    setInterval(() => this.clearExpiredCache(), 5 * 60 * 1000);
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
    if (params.category_id) queryParams.set('category_id', params.category_id);
    if (params.event_type) queryParams.set('event_type', params.event_type);
    if (params.publishing_status) queryParams.set('publishing_status', params.publishing_status);
    if (params.registration_status) queryParams.set('registration_status', params.registration_status);
    if (params.priority_level) queryParams.set('priority_level', params.priority_level);
    if (params.event_date_from) queryParams.set('event_date_from', params.event_date_from);
    if (params.event_date_to) queryParams.set('event_date_to', params.event_date_to);
    if (params.organizer_name) queryParams.set('organizer_name', params.organizer_name);
    if (params.featured !== undefined) queryParams.set('featured', params.featured.toString());
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const endpoint = `/api/v1/events${queryParams.toString() ? `?${queryParams}` : ''}`;
    const cacheKey = `events:${endpoint}`;
    
    return this.requestWithCache<EventsResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, EventsRestClient.CACHE_TTL.EVENT_LIST);
  }

  /**
   * Get event by slug
   * Maps to GET /api/v1/events/slug/{slug} endpoint through Public Gateway
   * Cached for performance
   */
  async getEventBySlug(slug: string): Promise<EventResponse> {
    if (!slug) {
      throw new Error('Event slug is required');
    }

    const endpoint = `/api/v1/events/slug/${encodeURIComponent(slug)}`;
    const cacheKey = `event:slug:${slug}`;
    
    return this.requestWithCache<EventResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, EventsRestClient.CACHE_TTL.EVENT_DETAIL);
  }

  /**
   * Get event by event_id (database primary key)
   * Maps to GET /api/v1/events/{event_id} endpoint through Public Gateway
   * Cached for performance
   */
  async getEventById(event_id: string): Promise<EventResponse> {
    if (!event_id) {
      throw new Error('Event ID is required');
    }

    const endpoint = `/api/v1/events/${encodeURIComponent(event_id)}`;
    const cacheKey = `event:id:${event_id}`;
    
    return this.requestWithCache<EventResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, EventsRestClient.CACHE_TTL.EVENT_DETAIL);
  }

  /**
   * Get featured events
   * Maps to GET /api/v1/events/featured endpoint through Public Gateway
   * Cached with medium TTL since featured events change occasionally
   */
  async getFeaturedEvents(): Promise<FeaturedEventResponse> {
    const endpoint = '/api/v1/events/featured';
    const cacheKey = 'events:featured';
    
    return this.requestWithCache<FeaturedEventResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, EventsRestClient.CACHE_TTL.FEATURED);
  }

  /**
   * Get published events
   * Uses getEvents with publishing_status filter
   */
  async getPublishedEvents(params: Partial<GetEventsParams> = {}): Promise<EventsResponse> {
    return this.getEvents({
      ...params,
      publishing_status: 'published',
    });
  }

  /**
   * Search events
   * Uses GET /api/v1/events/search endpoint with database schema-compliant parameters
   */
  async searchEvents(params: SearchEventsParams): Promise<EventsResponse> {
    const queryParams = new URLSearchParams();
    
    queryParams.set('q', params.q);
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category_id) queryParams.set('category_id', params.category_id);
    if (params.event_type) queryParams.set('event_type', params.event_type);
    if (params.publishing_status) queryParams.set('publishing_status', params.publishing_status);
    if (params.event_date_from) queryParams.set('event_date_from', params.event_date_from);
    if (params.event_date_to) queryParams.set('event_date_to', params.event_date_to);
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const endpoint = `/api/v1/events/search?${queryParams}`;
    
    return this.request<EventsResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get recent events
   * Uses getEvents with appropriate date sorting
   */
  async getRecentEvents(limit: number = 5): Promise<EventsResponse> {
    return this.getEvents({
      pageSize: limit,
      sortBy: 'event_date_desc',
      publishing_status: 'published',
    });
  }

  /**
   * Get events by category ID
   * Uses getEvents with category_id filter
   */
  async getEventsByCategory(category_id: string, params?: Partial<GetEventsParams>): Promise<EventsResponse> {
    if (!category_id) {
      throw new Error('Category ID is required');
    }

    return this.getEvents({
      ...params,
      category_id,
    });
  }

  /**
   * Get event categories
   * Maps to GET /api/v1/events/categories endpoint through Public Gateway
   * Heavily cached since categories change infrequently
   */
  async getEventCategories(): Promise<EventCategoriesResponse> {
    const endpoint = '/api/v1/events/categories';
    const cacheKey = 'events:categories';
    
    return this.requestWithCache<EventCategoriesResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, EventsRestClient.CACHE_TTL.CATEGORIES);
  }

  /**
   * Create new event
   * Maps to POST /api/v1/events endpoint through Admin Gateway
   * Clears relevant cache entries after creation
   */
  async createEvent(params: CreateEventParams): Promise<EventResponse> {
    const endpoint = '/api/v1/events';
    
    try {
      const result = await this.request<EventResponse>(endpoint, {
        method: 'POST',
        body: JSON.stringify(params),
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      // Invalidate relevant caches
      this.invalidateCachePattern('events:');
      
      return result;
    } catch (error) {
      this.metrics.errorCount++;
      throw error;
    }
  }

  /**
   * Update existing event
   * Maps to PUT /api/v1/events/{event_id} endpoint through Admin Gateway
   * Clears relevant cache entries after update
   */
  async updateEvent(params: UpdateEventParams): Promise<EventResponse> {
    const { event_id, ...updateData } = params;
    const endpoint = `/api/v1/events/${encodeURIComponent(event_id)}`;
    
    try {
      const result = await this.request<EventResponse>(endpoint, {
        method: 'PUT',
        body: JSON.stringify(updateData),
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      // Invalidate specific event caches and event lists
      this.invalidateCacheKey(`event:id:${event_id}`);
      this.invalidateCachePattern('events:');
      
      return result;
    } catch (error) {
      this.metrics.errorCount++;
      throw error;
    }
  }

  /**
   * Delete event (soft delete)
   * Maps to DELETE /api/v1/events/{event_id} endpoint through Admin Gateway
   * Clears relevant cache entries after deletion
   */
  async deleteEvent(event_id: string): Promise<void> {
    if (!event_id) {
      throw new Error('Event ID is required');
    }

    const endpoint = `/api/v1/events/${encodeURIComponent(event_id)}`;
    
    try {
      await this.request<void>(endpoint, {
        method: 'DELETE',
      });
      
      // Invalidate specific event caches and event lists
      this.invalidateCacheKey(`event:id:${event_id}`);
      this.invalidateCachePattern('events:');
      this.invalidateCachePattern('event:');
    } catch (error) {
      this.metrics.errorCount++;
      throw error;
    }
  }

  // Performance optimization methods
  
  /**
   * Request with caching and deduplication
   */
  private async requestWithCache<T>(
    endpoint: string,
    options: RequestInit,
    cacheKey: string,
    ttl: number
  ): Promise<T> {
    const startTime = performance.now();
    this.metrics.totalRequests++;
    
    // Check cache first
    const cached = this.getFromCache<T>(cacheKey);
    if (cached) {
      this.metrics.cacheHits++;
      return cached;
    }
    
    this.metrics.cacheMisses++;
    
    // Check for pending request to prevent duplicate requests
    const requestKey = `${endpoint}:${JSON.stringify(options)}`;
    if (this.pendingRequests.has(requestKey)) {
      return this.pendingRequests.get(requestKey)!;
    }
    
    // Make the request
    const requestPromise = this.request<T>(endpoint, options);
    this.pendingRequests.set(requestKey, requestPromise);
    
    try {
      const result = await requestPromise;
      
      // Cache the result
      this.setCache(cacheKey, result, ttl);
      
      // Update performance metrics
      const endTime = performance.now();
      const requestTime = endTime - startTime;
      this.updateAverageResponseTime(requestTime);
      
      return result;
    } catch (error) {
      this.metrics.errorCount++;
      throw error;
    } finally {
      this.pendingRequests.delete(requestKey);
    }
  }
  
  /**
   * Get data from cache if not expired
   */
  private getFromCache<T>(key: string): T | null {
    const entry = this.cache.get(key);
    if (!entry) return null;
    
    const now = Date.now();
    if (now > entry.timestamp + entry.ttl) {
      this.cache.delete(key);
      return null;
    }
    
    return entry.data;
  }
  
  /**
   * Set data in cache with TTL
   */
  private setCache<T>(key: string, data: T, ttl: number): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
  }
  
  /**
   * Invalidate specific cache key
   */
  private invalidateCacheKey(key: string): void {
    this.cache.delete(key);
  }
  
  /**
   * Invalidate cache keys matching pattern
   */
  private invalidateCachePattern(pattern: string): void {
    const keysToDelete = [];
    for (const key of this.cache.keys()) {
      if (key.startsWith(pattern)) {
        keysToDelete.push(key);
      }
    }
    keysToDelete.forEach(key => this.cache.delete(key));
  }
  
  /**
   * Clear expired cache entries
   */
  private clearExpiredCache(): void {
    const now = Date.now();
    const keysToDelete = [];
    
    for (const [key, entry] of this.cache.entries()) {
      if (now > entry.timestamp + entry.ttl) {
        keysToDelete.push(key);
      }
    }
    
    keysToDelete.forEach(key => this.cache.delete(key));
  }
  
  /**
   * Update average response time metric
   */
  private updateAverageResponseTime(responseTime: number): void {
    const totalResponseTime = this.metrics.averageResponseTime * (this.metrics.totalRequests - 1);
    this.metrics.averageResponseTime = (totalResponseTime + responseTime) / this.metrics.totalRequests;
  }
  
  /**
   * Get performance metrics
   */
  public getMetrics(): RequestMetrics {
    return { ...this.metrics };
  }
  
  /**
   * Clear all cache entries
   */
  public clearCache(): void {
    this.cache.clear();
  }
  
  /**
   * Get cache statistics
   */
  public getCacheStats(): { size: number; hitRate: number } {
    const hitRate = this.metrics.totalRequests > 0 
      ? (this.metrics.cacheHits / this.metrics.totalRequests) * 100 
      : 0;
    
    return {
      size: this.cache.size,
      hitRate: Math.round(hitRate * 100) / 100,
    };
  }
}