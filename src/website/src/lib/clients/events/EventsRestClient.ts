// Events REST Client - Database Schema Compliant with Performance Optimizations
// Updated to work with TABLES-EVENTS.md aligned types
// Features: Response caching, request deduplication, performance monitoring

import { BaseRestClient } from '../rest/BaseRestClient';
import { config } from '../../environments';
import {
  DomainRestClientConfig,
  createGetListMethod,
  createGetBySlugMethod,
  createGetByIdMethod,
  createGetFeaturedMethod,
  createSearchMethod,
  createGetCategoriesMethod,
  createGetRecentMethod,
  createGetByCategoryMethod,
  createCRUDMethods
} from '../rest/DomainRestClientFactory';
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
  
  // Domain configuration for factory methods
  private static readonly DOMAIN_CONFIG: DomainRestClientConfig = {
    domain: 'events',
    baseEndpoint: '/api/v1/events',
    itemsField: 'events',
    cacheTTL: {
      categories: 15 * 60 * 1000, // 15 minutes - categories change infrequently
      featured: 5 * 60 * 1000, // 5 minutes - featured events change occasionally  
      detail: 2 * 60 * 1000, // 2 minutes - individual events may be updated
      list: 30 * 1000, // 30 seconds - lists change more frequently
    }
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

  // Factory-generated methods
  getEvents = createGetListMethod<GetEventsParams, EventsResponse>(this, EventsRestClient.DOMAIN_CONFIG);

  getEventBySlug = createGetBySlugMethod<EventResponse>(this, EventsRestClient.DOMAIN_CONFIG, 'Event');

  getEventById = createGetByIdMethod<EventResponse>(this, EventsRestClient.DOMAIN_CONFIG, 'Event');

  getFeaturedEvents = createGetFeaturedMethod<FeaturedEventResponse>(this, EventsRestClient.DOMAIN_CONFIG, '/published');

  // Convenience methods
  async getPublishedEvents(params: Partial<GetEventsParams> = {}): Promise<EventsResponse> {
    return this.getEvents({
      ...params,
      publishing_status: 'published',
    });
  }

  searchEvents = createSearchMethod<SearchEventsParams, EventsResponse>(this, EventsRestClient.DOMAIN_CONFIG, false);

  getRecentEvents = createGetRecentMethod<EventsResponse>(this, EventsRestClient.DOMAIN_CONFIG);

  getEventsByCategory = createGetByCategoryMethod<EventsResponse>(this, EventsRestClient.DOMAIN_CONFIG);

  getEventCategories = createGetCategoriesMethod<EventCategoriesResponse>(this, EventsRestClient.DOMAIN_CONFIG);

  // CRUD operations
  private crudMethods = createCRUDMethods<CreateEventParams, UpdateEventParams, EventResponse>(this, EventsRestClient.DOMAIN_CONFIG);
  createEvent = this.crudMethods.create;
  updateEvent = this.crudMethods.update;
  deleteEvent = this.crudMethods.delete;



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