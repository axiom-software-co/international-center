// News REST Client - Database Schema Compliant with Performance Optimizations
// Updated to work with TABLES-NEWS.md aligned types
// Features: Response caching, request deduplication, performance monitoring

import { BaseRestClient } from '../rest/BaseRestClient';
import { RestClientCache, STANDARD_CACHE_TTL } from '../rest/RestClientCache';
import { config } from '../../environments';
import type {
  NewsArticle,
  NewsResponse,
  NewsArticleResponse,
  NewsCategoriesResponse,
  GetNewsParams,
  SearchNewsParams,
} from './types';

export class NewsRestClient extends BaseRestClient {
  private cache = new RestClientCache();

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
    
    return this.cache.requestWithCache<NewsResponse>(
      this,
      url,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.LIST
    );
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
    
    return this.cache.requestWithCache<NewsArticleResponse>(
      this,
      url,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.DETAIL
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
    
    return this.cache.requestWithCache<NewsResponse>(
      this,
      url,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.FEATURED
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
    
    return this.cache.requestWithCache<NewsCategoriesResponse>(
      this,
      url,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.CATEGORIES
    );
  }

  /**
   * Get performance metrics for monitoring
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
   * Clear all cached data and metrics
   */
  public clearCache(): void {
    this.cache.clearCache();
  }
}