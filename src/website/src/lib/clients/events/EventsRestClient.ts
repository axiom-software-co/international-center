// Events REST Client - REST API client for events domain
// Follows standardized REST patterns with proper error handling and response transformation

import { BaseRestClient } from '../rest/BaseRestClient';
import { config } from '../../environments';
import type {
  Event,
  EventsResponse,
  EventResponse,
  GetEventsParams,
  SearchEventsParams,
} from './types';

export class EventsRestClient extends BaseRestClient {
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
   * Returns backend format: { events: [...], count: number, correlation_id: string }
   */
  async getEvents(params: GetEventsParams = {}): Promise<EventsResponse> {
    const queryParams = new URLSearchParams();
    
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);
    if (params.featured !== undefined) queryParams.set('featured', params.featured.toString());
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const endpoint = `/api/v1/events${queryParams.toString() ? `?${queryParams}` : ''}`;
    
    return this.request<EventsResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get event by slug
   * Maps to GET /api/v1/events/slug/{slug} endpoint through Public Gateway
   * Returns backend format: { event: {...}, correlation_id: string }
   */
  async getEventBySlug(slug: string): Promise<EventResponse> {
    if (!slug) {
      throw new Error('Event slug is required');
    }

    const endpoint = `/api/v1/events/slug/${encodeURIComponent(slug)}`;
    
    return this.request<EventResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get event by ID
   * Maps to GET /api/v1/events/{id} endpoint through Public Gateway
   * Returns backend format: { event: {...}, correlation_id: string }
   */
  async getEventById(id: string): Promise<EventResponse> {
    if (!id) {
      throw new Error('Event ID is required');
    }

    const endpoint = `/api/v1/events/${encodeURIComponent(id)}`;
    
    return this.request<EventResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get featured events (published events)
   * Maps to GET /api/v1/events/published endpoint through Public Gateway
   * Returns backend format: { events: [...], count: number, correlation_id: string }
   */
  async getFeaturedEvents(limit?: number): Promise<EventsResponse> {
    const endpoint = '/api/v1/events/published';
    
    return this.request<EventsResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Search events
   * Uses GET /api/v1/events with search parameter
   * Returns backend format: { events: [...], count: number, correlation_id: string }
   */
  async searchEvents(params: SearchEventsParams): Promise<EventsResponse> {
    const queryParams = new URLSearchParams();
    
    queryParams.set('search', params.q);
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);

    const endpoint = `/api/v1/events?${queryParams}`;
    
    return this.request<EventsResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get recent events
   * Uses GET /api/v1/events with sortBy parameter
   * Returns backend format: { events: [...], count: number, correlation_id: string }
   */
  async getRecentEvents(limit: number = 5): Promise<EventsResponse> {
    const queryParams = new URLSearchParams();
    queryParams.set('limit', limit.toString());
    queryParams.set('sortBy', 'date-desc');

    const endpoint = `/api/v1/events?${queryParams}`;
    
    return this.request<EventsResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get events by category
   * Maps to GET /api/v1/events/categories/{category}/events endpoint through Public Gateway
   * Returns backend format: { events: [...], count: number, correlation_id: string }
   */
  async getEventsByCategory(category: string, params?: Partial<GetEventsParams>): Promise<EventsResponse> {
    if (!category) {
      throw new Error('Category is required');
    }

    // Handle category filtering by using different endpoint
    const endpoint = `/api/v1/events/categories/${encodeURIComponent(category)}/events`;
    return this.request<EventsResponse>(endpoint, { method: 'GET' });
  }
}