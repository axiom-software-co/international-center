import { BaseRestClient, RestClientConfig } from './BaseRestClient';
import { RestClientCache, STANDARD_CACHE_TTL } from './RestClientCache';
import { config } from '../../../environments';
import type { 
  Service, 
  GetServicesParams, 
  SearchServicesParams 
} from '../services/types';
import type {
  BackendServicesResponse,
  BackendServiceResponse,
  BackendServiceCategoriesResponse,
  BackendFeaturedCategoriesResponse,
  BackendServicesParams,
  BackendSearchServicesParams
} from './BackendTypes';

export class ServicesRestClient extends BaseRestClient {
  private cache = new RestClientCache();

  constructor() {
    super({
      baseUrl: config.domains.services.baseUrl,
      timeout: config.domains.services.timeout,
      retryAttempts: config.domains.services.retryAttempts,
    });
  }

  /**
   * Get services list
   * Maps to GET /api/v1/services endpoint through Public Gateway
   * Returns backend format: { services: [...], count: number, correlation_id: string }
   * Uses caching for non-search queries
   */
  async getServices(params: GetServicesParams = {}): Promise<BackendServicesResponse> {
    // Handle category filtering by using different endpoint
    if (params.category) {
      const endpoint = `/api/v1/services/categories/${encodeURIComponent(params.category)}/services`;
      return this.request<BackendServicesResponse>(endpoint, { method: 'GET' });
    }

    const queryParams = new URLSearchParams();
    
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.search) queryParams.set('search', params.search);

    const endpoint = `/api/v1/services${queryParams.toString() ? `?${queryParams}` : ''}`;
    
    // Use caching for non-search queries
    if (!params.search) {
      const cacheKey = `services:${queryParams.toString()}`;
      return this.cache.requestWithCache<BackendServicesResponse>(
        this,
        endpoint,
        { method: 'GET' },
        cacheKey,
        STANDARD_CACHE_TTL.LIST
      );
    }
    
    // No caching for search queries to ensure real-time results
    return this.request<BackendServicesResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get service by slug
   * Maps to GET /api/v1/services/slug/{slug} endpoint through Public Gateway
   * Returns backend format: { service: {...}, correlation_id: string }
   * Uses caching for performance
   */
  async getServiceBySlug(slug: string): Promise<BackendServiceResponse> {
    if (!slug) {
      throw new Error('Service slug is required');
    }

    const endpoint = `/api/v1/services/slug/${encodeURIComponent(slug)}`;
    const cacheKey = `services:slug:${slug}`;
    
    return this.cache.requestWithCache<BackendServiceResponse>(
      this,
      endpoint,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.DETAIL
    );
  }

  /**
   * Get service categories
   * Maps to GET /api/v1/services/categories endpoint through Public Gateway
   * Returns backend format: { categories: [...], count: number, correlation_id: string }
   * Uses aggressive caching as categories change infrequently
   */
  async getServiceCategories(): Promise<BackendServiceCategoriesResponse> {
    const endpoint = '/api/v1/services/categories';
    const cacheKey = 'services:categories';
    
    return this.cache.requestWithCache<BackendServiceCategoriesResponse>(
      this,
      endpoint,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.CATEGORIES
    );
  }

  /**
   * Get featured services (published services)
   * Maps to GET /api/v1/services/published endpoint through Public Gateway
   * Returns backend format: { services: [...], count: number, correlation_id: string }
   * Uses caching for featured content
   */
  async getFeaturedServices(limit?: number): Promise<BackendServicesResponse> {
    const endpoint = '/api/v1/services/published';
    const cacheKey = `services:featured${limit ? `:${limit}` : ''}`;
    
    return this.cache.requestWithCache<BackendServicesResponse>(
      this,
      endpoint,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.FEATURED
    );
  }

  /**
   * Get featured categories
   * Maps to GET /api/v1/services/featured endpoint through Public Gateway
   * Returns backend format: { featured_categories: [...], count: number, correlation_id: string }
   */
  async getFeaturedCategories(): Promise<BackendFeaturedCategoriesResponse> {
    const endpoint = '/api/v1/services/featured';
    
    return this.request<BackendFeaturedCategoriesResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Search services
   * Uses GET /api/v1/services with search parameter
   * Returns backend format: { services: [...], count: number, correlation_id: string }
   */
  async searchServices(params: SearchServicesParams): Promise<BackendServicesResponse> {
    const queryParams = new URLSearchParams();
    
    queryParams.set('search', params.q);
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);

    const endpoint = `/api/v1/services?${queryParams}`;
    
    return this.request<BackendServicesResponse>(endpoint, {
      method: 'GET',
    });
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
   * Clear all cache entries and reset metrics
   */
  public clearCache(): void {
    this.cache.clearCache();
  }
}