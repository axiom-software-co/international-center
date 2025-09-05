// Research REST Client - Database Schema Compliant with Performance Optimizations
// Updated to work with TABLES-RESEARCH.md aligned types
// Features: Response caching, request deduplication, performance monitoring

import { BaseRestClient } from '../rest/BaseRestClient';
import { config } from '../../environments';
import type {
  ResearchArticle,
  ResearchResponse,
  ResearchArticleResponse,
  ResearchCategoriesResponse,
  FeaturedResearchResponse,
  GetResearchParams,
  SearchResearchParams,
  CreateResearchParams,
  UpdateResearchParams,
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

export class ResearchRestClient extends BaseRestClient {
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
    FEATURED: 5 * 60 * 1000, // 5 minutes - featured research changes occasionally
    ARTICLE_DETAIL: 3 * 60 * 1000, // 3 minutes - research articles may be updated  
    ARTICLE_LIST: 45 * 1000, // 45 seconds - lists change more frequently
  };

  constructor() {
    // Handle test environment or missing configuration
    const researchConfig = config.domains?.research || {
      baseUrl: 'http://localhost:7220', // Public Gateway URL fallback
      timeout: 5000,
      retryAttempts: 2,
    };
    
    super({
      baseUrl: researchConfig.baseUrl,
      timeout: researchConfig.timeout,
      retryAttempts: researchConfig.retryAttempts,
    });

    // Clear expired cache entries every 5 minutes
    setInterval(() => this.clearExpiredCache(), 5 * 60 * 1000);
  }

  /**
   * Get all research articles with optional filtering and pagination
   * Maps to GET /api/v1/research endpoint through Public Gateway
   * Uses database schema-compliant query parameters with caching
   */
  async getResearchArticles(params: GetResearchParams = {}): Promise<ResearchResponse> {
    const queryParams = new URLSearchParams();
    
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category_id) queryParams.set('category_id', params.category_id);
    if (params.research_type) queryParams.set('research_type', params.research_type);
    if (params.publishing_status) queryParams.set('publishing_status', params.publishing_status);
    if (params.publication_date_from) queryParams.set('publication_date_from', params.publication_date_from);
    if (params.publication_date_to) queryParams.set('publication_date_to', params.publication_date_to);
    if (params.author_names) queryParams.set('author_names', params.author_names);
    if (params.doi) queryParams.set('doi', params.doi);
    if (params.industry) queryParams.set('industry', params.industry);
    if (params.featured !== undefined) queryParams.set('featured', params.featured.toString());
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const endpoint = `/api/v1/research${queryParams.toString() ? `?${queryParams}` : ''}`;
    const cacheKey = `research:${endpoint}`;
    
    return this.requestWithCache<ResearchResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, ResearchRestClient.CACHE_TTL.ARTICLE_LIST);
  }

  /**
   * Get research article by slug
   * Maps to GET /api/v1/research/slug/{slug} endpoint through Public Gateway
   * Cached for performance
   */
  async getResearchArticleBySlug(slug: string): Promise<ResearchArticleResponse> {
    if (!slug) {
      throw new Error('Research article slug is required');
    }

    const endpoint = `/api/v1/research/slug/${encodeURIComponent(slug)}`;
    const cacheKey = `research:slug:${slug}`;
    
    return this.requestWithCache<ResearchArticleResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, ResearchRestClient.CACHE_TTL.ARTICLE_DETAIL);
  }

  /**
   * Get research article by research_id (database primary key)
   * Maps to GET /api/v1/research/{research_id} endpoint through Public Gateway
   * Cached for performance
   */
  async getResearchArticleById(research_id: string): Promise<ResearchArticleResponse> {
    if (!research_id) {
      throw new Error('Research article ID is required');
    }

    const endpoint = `/api/v1/research/${encodeURIComponent(research_id)}`;
    const cacheKey = `research:id:${research_id}`;
    
    return this.requestWithCache<ResearchArticleResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, ResearchRestClient.CACHE_TTL.ARTICLE_DETAIL);
  }

  /**
   * Get featured research
   * Maps to GET /api/v1/research/featured endpoint through Public Gateway
   * Contract aligned with test expectations - supports optional limit parameter
   */
  async getFeaturedResearch(limit?: number): Promise<FeaturedResearchResponse> {
    const queryParams = new URLSearchParams();
    if (limit !== undefined) queryParams.set('limit', limit.toString());
    
    const endpoint = `/api/v1/research/featured${queryParams.toString() ? `?${queryParams}` : ''}`;
    const cacheKey = `research:featured${limit ? `:${limit}` : ''}`;
    
    return this.requestWithCache<FeaturedResearchResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, ResearchRestClient.CACHE_TTL.FEATURED);
  }

  /**
   * Get published research articles
   * Uses getResearchArticles with publishing_status filter
   */
  async getPublishedResearch(params: Partial<GetResearchParams> = {}): Promise<ResearchResponse> {
    return this.getResearchArticles({
      ...params,
      publishing_status: 'published',
    });
  }

  /**
   * Search research articles
   * Uses GET /api/v1/research/search endpoint with database schema-compliant parameters
   */
  async searchResearch(params: SearchResearchParams): Promise<ResearchResponse> {
    const queryParams = new URLSearchParams();
    
    queryParams.set('q', params.q);
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);
    if (params.research_type) queryParams.set('research_type', params.research_type);
    if (params.publishing_status) queryParams.set('publishing_status', params.publishing_status);
    if (params.publication_date_from) queryParams.set('publication_date_from', params.publication_date_from);
    if (params.publication_date_to) queryParams.set('publication_date_to', params.publication_date_to);
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const endpoint = `/api/v1/research/search?${queryParams}`;
    
    return this.request<ResearchResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get recent research articles
   * Contract aligned with test expectations using limit and date-desc parameters
   */
  async getRecentResearch(limit: number = 5): Promise<ResearchResponse> {
    const queryParams = new URLSearchParams();
    queryParams.set('limit', limit.toString());
    queryParams.set('sortBy', 'date-desc');
    
    const endpoint = `/api/v1/research?${queryParams}`;
    
    return this.request<ResearchResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get research articles by category
   * Uses RESTful endpoint - contract aligned with test expectations
   */
  async getResearchByCategory(category_id: string, params?: Partial<GetResearchParams>): Promise<ResearchResponse> {
    if (!category_id) {
      throw new Error('Category is required');
    }

    const endpoint = `/api/v1/research/categories/${encodeURIComponent(category_id)}/articles`;
    
    return this.request<ResearchResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get research articles by industry
   * Contract aligned with test expectations using industry query parameter
   */
  async getResearchByIndustry(industry: string, params?: Partial<GetResearchParams>): Promise<ResearchResponse> {
    if (!industry) {
      throw new Error('Industry is required');
    }

    return this.getResearchArticles({
      ...params,
      industry,
    });
  }

  /**
   * Get research categories
   * Maps to GET /api/v1/research/categories endpoint through Public Gateway
   * Heavily cached since categories change infrequently
   */
  async getResearchCategories(): Promise<ResearchCategoriesResponse> {
    const endpoint = '/api/v1/research/categories';
    const cacheKey = 'research:categories';
    
    return this.requestWithCache<ResearchCategoriesResponse>(endpoint, {
      method: 'GET',
    }, cacheKey, ResearchRestClient.CACHE_TTL.CATEGORIES);
  }

  /**
   * Create new research article
   * Maps to POST /api/v1/research endpoint through Admin Gateway
   * Clears relevant cache entries after creation
   */
  async createResearchArticle(params: CreateResearchParams): Promise<ResearchArticleResponse> {
    const endpoint = '/api/v1/research';
    
    try {
      const result = await this.request<ResearchArticleResponse>(endpoint, {
        method: 'POST',
        body: JSON.stringify(params),
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      // Invalidate relevant caches
      this.invalidateCachePattern('research:');
      
      return result;
    } catch (error) {
      this.metrics.errorCount++;
      throw error;
    }
  }

  /**
   * Update existing research article
   * Maps to PUT /api/v1/research/{research_id} endpoint through Admin Gateway
   * Clears relevant cache entries after update
   */
  async updateResearchArticle(params: UpdateResearchParams): Promise<ResearchArticleResponse> {
    const { research_id, ...updateData } = params;
    const endpoint = `/api/v1/research/${encodeURIComponent(research_id)}`;
    
    try {
      const result = await this.request<ResearchArticleResponse>(endpoint, {
        method: 'PUT',
        body: JSON.stringify(updateData),
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      // Invalidate specific article caches and research lists
      this.invalidateCacheKey(`research:id:${research_id}`);
      this.invalidateCachePattern('research:');
      
      return result;
    } catch (error) {
      this.metrics.errorCount++;
      throw error;
    }
  }

  /**
   * Delete research article (soft delete)
   * Maps to DELETE /api/v1/research/{research_id} endpoint through Admin Gateway
   * Clears relevant cache entries after deletion
   */
  async deleteResearchArticle(research_id: string): Promise<void> {
    if (!research_id) {
      throw new Error('Research ID is required');
    }

    const endpoint = `/api/v1/research/${encodeURIComponent(research_id)}`;
    
    try {
      await this.request<void>(endpoint, {
        method: 'DELETE',
      });
      
      // Invalidate specific article caches and research lists
      this.invalidateCacheKey(`research:id:${research_id}`);
      this.invalidateCachePattern('research:');
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