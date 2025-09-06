// Research REST Client - Database Schema Compliant with Performance Optimizations
// Updated to work with TABLES-RESEARCH.md aligned types
// Features: Response caching, request deduplication, performance monitoring

import { BaseRestClient } from '../rest/BaseRestClient';
import { RestClientCache, STANDARD_CACHE_TTL } from '../rest/RestClientCache';
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

export class ResearchRestClient extends BaseRestClient {
  private cache = new RestClientCache();

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
    
    return this.cache.requestWithCache<ResearchResponse>(
      this,
      endpoint,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.LIST
    );
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
    
    return this.cache.requestWithCache<ResearchArticleResponse>(
      this,
      endpoint,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.DETAIL
    );
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
    
    return this.cache.requestWithCache<ResearchArticleResponse>(
      this,
      endpoint,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.DETAIL
    );
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
    
    return this.cache.requestWithCache<FeaturedResearchResponse>(
      this,
      endpoint,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.FEATURED
    );
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
    
    return this.cache.requestWithCache<ResearchCategoriesResponse>(
      this,
      endpoint,
      { method: 'GET' },
      cacheKey,
      STANDARD_CACHE_TTL.CATEGORIES
    );
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