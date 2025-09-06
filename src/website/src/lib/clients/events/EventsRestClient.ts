// Events REST Client - Database Schema Compliant with Performance Optimizations
// Updated to work with TABLES-EVENTS.md aligned types
// Features: Response caching, request deduplication, performance monitoring

import { BaseRestClient } from '../rest/BaseRestClient';
import { RestClientCache, STANDARD_CACHE_TTL } from '../rest/RestClientCache';
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

export class EventsRestClient extends BaseRestClient {
  private cache = new RestClientCache();
  
  // Domain configuration for factory methods - updated to use shared cache TTL
  private static readonly DOMAIN_CONFIG: DomainRestClientConfig = {
    domain: 'events',
    baseEndpoint: '/api/v1/events',
    itemsField: 'events',
    cacheTTL: {
      categories: STANDARD_CACHE_TTL.CATEGORIES,
      featured: STANDARD_CACHE_TTL.FEATURED,
      detail: STANDARD_CACHE_TTL.DETAIL,
      list: STANDARD_CACHE_TTL.LIST,
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