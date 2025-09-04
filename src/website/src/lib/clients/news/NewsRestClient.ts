// News REST Client - Database Schema Compliant with Performance Optimizations
// Updated to work with TABLES-NEWS.md aligned types
// Features: Response caching, request deduplication, performance monitoring

import { BaseRestClient } from '../rest/BaseRestClient';
import { config } from '../../environments';
import type {
  NewsArticle,
  NewsResponse,
  NewsArticleResponse,
  NewsCategoriesResponse,
  GetNewsParams,
  SearchNewsParams,
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

export class NewsRestClient extends BaseRestClient {
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
    FEATURED: 5 * 60 * 1000, // 5 minutes - featured news changes occasionally
    ARTICLE_DETAIL: 2 * 60 * 1000, // 2 minutes - news articles may be updated frequently
    ARTICLE_LIST: 30 * 1000, // 30 seconds - news lists change very frequently
  };

  constructor(baseUrl?: string) {
    // Handle test environment or missing configuration
    const newsConfig = config?.domains?.news || {
      baseUrl: baseUrl || 'http://localhost:7220', // Public Gateway URL fallback
      timeout: 5000,
      retryAttempts: 2,
    };
    
    super({
      baseUrl: newsConfig.baseUrl,
      timeout: newsConfig.timeout,
      retryAttempts: newsConfig.retryAttempts,
    });

    // Clear expired cache entries every 5 minutes
    setInterval(() => this.clearExpiredCache(), 5 * 60 * 1000);
  }

  /**
   * Get all news articles with optional filtering and pagination
   * Maps to GET /api/v1/news endpoint through Public Gateway
   * Uses database schema-compliant query parameters with caching
   */
  async getNews(params: GetNewsParams = {}): Promise<NewsResponse> {
    const queryParams = new URLSearchParams();
    
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);
    if (params.featured !== undefined) queryParams.set('featured', params.featured.toString());
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const cacheKey = `news:${queryParams.toString()}`;
    const url = `/api/v1/news${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    
    return this.cachedRequest(cacheKey, () => this.request<NewsResponse>(url, { method: 'GET' }), NewsRestClient.CACHE_TTL.ARTICLE_LIST);
  }

  /**
   * Get single news article by slug
   * Maps to GET /api/v1/news/slug/{slug} endpoint through Public Gateway
   * Uses caching for article detail with database schema compliance
   */
  async getNewsArticleBySlug(slug: string): Promise<NewsArticleResponse> {
    if (!slug || slug.trim() === '') {
      throw new Error('News article slug is required');
    }

    const cacheKey = `news:slug:${slug}`;
    const url = `/api/v1/news/slug/${encodeURIComponent(slug)}`;
    
    return this.cachedRequest(
      cacheKey,
      () => this.request<NewsArticleResponse>(url, { method: 'GET' }),
      NewsRestClient.CACHE_TTL.ARTICLE_DETAIL
    );
  }

  /**
   * Get featured news articles with optional limit
   * Maps to GET /api/v1/news/featured endpoint through Public Gateway
   * Uses caching for featured content with database schema compliance
   */
  async getFeaturedNews(limit?: number): Promise<NewsResponse> {
    const queryParams = new URLSearchParams();
    if (limit !== undefined) queryParams.set('limit', limit.toString());

    const cacheKey = `news:featured:${limit || 'all'}`;
    const url = `/api/v1/news/featured${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    
    return this.cachedRequest(
      cacheKey,
      () => this.request<NewsResponse>(url, { method: 'GET' }),
      NewsRestClient.CACHE_TTL.FEATURED
    );
  }

  /**
   * Search news articles with query parameters
   * Maps to GET /api/v1/news/search endpoint through Public Gateway
   * Database schema-compliant search with no caching for real-time results
   */
  async searchNews(params: SearchNewsParams): Promise<NewsResponse> {
    // Return empty results for empty queries without making HTTP request
    if (!params.q || params.q.trim() === '') {
      return {
        news: [],
        count: 0,
        correlation_id: `empty-search-${Date.now()}`
      };
    }

    const queryParams = new URLSearchParams();
    queryParams.set('q', params.q.trim());
    
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const url = `/api/v1/news/search?${queryParams.toString()}`;
    
    // Search results are not cached for real-time freshness
    return this.request<NewsResponse>(url, { method: 'GET' });
  }

  /**
   * Get news categories
   * Maps to GET /api/v1/news/categories endpoint through Public Gateway
   * Uses aggressive caching as categories change infrequently
   */
  async getNewsCategories(): Promise<NewsCategoriesResponse> {
    const cacheKey = 'news:categories';
    const url = '/api/v1/news/categories';
    
    return this.cachedRequest(
      cacheKey,
      () => this.request<NewsCategoriesResponse>(url, { method: 'GET' }),
      NewsRestClient.CACHE_TTL.CATEGORIES
    );
  }

  /**
   * Cached request wrapper with performance monitoring and deduplication
   */
  private async cachedRequest<T>(
    cacheKey: string,
    requestFn: () => Promise<T>,
    ttl: number
  ): Promise<T> {
    const startTime = Date.now();
    this.metrics.totalRequests++;

    // Check cache first
    const cachedEntry = this.cache.get(cacheKey);
    if (cachedEntry && (Date.now() - cachedEntry.timestamp) < cachedEntry.ttl) {
      this.metrics.cacheHits++;
      this.updateResponseTime(startTime);
      return cachedEntry.data;
    }

    this.metrics.cacheMisses++;

    // Check for pending request to avoid duplicate requests
    if (this.pendingRequests.has(cacheKey)) {
      return this.pendingRequests.get(cacheKey)!;
    }

    // Make the request
    const requestPromise = requestFn()
      .then((data) => {
        // Cache successful response
        this.cache.set(cacheKey, {
          data,
          timestamp: Date.now(),
          ttl,
        });

        this.updateResponseTime(startTime);
        return data;
      })
      .catch((error) => {
        this.metrics.errorCount++;
        throw error;
      })
      .finally(() => {
        // Remove from pending requests
        this.pendingRequests.delete(cacheKey);
      });

    // Store pending request
    this.pendingRequests.set(cacheKey, requestPromise);

    return requestPromise;
  }

  /**
   * Clear expired cache entries to prevent memory leaks
   */
  private clearExpiredCache(): void {
    const now = Date.now();
    for (const [key, entry] of this.cache.entries()) {
      if ((now - entry.timestamp) >= entry.ttl) {
        this.cache.delete(key);
      }
    }
  }

  /**
   * Update average response time metrics
   */
  private updateResponseTime(startTime: number): void {
    const responseTime = Date.now() - startTime;
    this.metrics.averageResponseTime = 
      (this.metrics.averageResponseTime * (this.metrics.totalRequests - 1) + responseTime) / 
      this.metrics.totalRequests;
  }

  /**
   * Get performance metrics for monitoring
   */
  public getMetrics(): RequestMetrics {
    return { ...this.metrics };
  }

  /**
   * Clear all cached data and metrics
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
}