import { BaseRestClient, RestClientConfig } from './BaseRestClient';
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

export class ServicesRestClient extends BaseRestClient {
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
    FEATURED: 5 * 60 * 1000, // 5 minutes - featured services change occasionally
    SERVICE_DETAIL: 2 * 60 * 1000, // 2 minutes - services may be updated frequently
    SERVICE_LIST: 30 * 1000, // 30 seconds - service lists change more frequently
  };

  constructor() {
    super({
      baseUrl: config.domains.services.baseUrl,
      timeout: config.domains.services.timeout,
      retryAttempts: config.domains.services.retryAttempts,
    });

    // Clear expired cache entries every 5 minutes
    setInterval(() => this.clearExpiredCache(), 5 * 60 * 1000);
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
      return this.requestWithCache<BackendServicesResponse>(
        endpoint,
        { method: 'GET' },
        cacheKey,
        ServicesRestClient.CACHE_TTL.SERVICE_LIST
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
    
    return this.requestWithCache<BackendServiceResponse>(
      endpoint,
      { method: 'GET' },
      cacheKey,
      ServicesRestClient.CACHE_TTL.SERVICE_DETAIL
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
    
    return this.requestWithCache<BackendServiceCategoriesResponse>(
      endpoint,
      { method: 'GET' },
      cacheKey,
      ServicesRestClient.CACHE_TTL.CATEGORIES
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
    
    return this.requestWithCache<BackendServicesResponse>(
      endpoint,
      { method: 'GET' },
      cacheKey,
      ServicesRestClient.CACHE_TTL.FEATURED
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
    const startTime = Date.now();
    this.metrics.totalRequests++;
    
    // Check cache first
    const cached = this.getFromCache<T>(cacheKey);
    if (cached) {
      this.metrics.cacheHits++;
      this.updateResponseTime(startTime);
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
      this.updateResponseTime(startTime);
      
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
  private updateResponseTime(startTime: number): void {
    const responseTime = Date.now() - startTime;
    this.metrics.averageResponseTime = 
      (this.metrics.averageResponseTime * (this.metrics.totalRequests - 1) + responseTime) / 
      this.metrics.totalRequests;
  }
  
  /**
   * Get performance metrics
   */
  public getMetrics(): RequestMetrics {
    return { ...this.metrics };
  }
  
  /**
   * Clear all cache entries and reset metrics
   */
  public clearCache(): void {
    this.cache.clear();
    this.pendingRequests.clear();
    this.metrics = {
      totalRequests: 0,
      cacheHits: 0,
      cacheMisses: 0,
      averageResponseTime: 0,
      errorCount: 0,
    };
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