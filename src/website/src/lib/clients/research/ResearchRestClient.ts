// Research REST Client - REST API client for research domain
// Follows standardized REST patterns with proper error handling and response transformation

import { BaseRestClient } from '../rest/BaseRestClient';
import { config } from '../../environments';
import type {
  ResearchArticle,
  ResearchResponse,
  ResearchArticleResponse,
  GetResearchParams,
  SearchResearchParams,
} from './types';

export class ResearchRestClient extends BaseRestClient {
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
   * Returns backend format: { data: [...], pagination: {...}, success: boolean }
   */
  async getResearchArticles(params: GetResearchParams = {}): Promise<ResearchResponse> {
    const queryParams = new URLSearchParams();
    
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);
    if (params.featured !== undefined) queryParams.set('featured', params.featured.toString());
    if (params.industry) queryParams.set('industry', params.industry);
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const endpoint = `/api/v1/research${queryParams.toString() ? `?${queryParams}` : ''}`;
    
    return this.request<ResearchResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get research article by slug
   * Maps to GET /api/v1/research/slug/{slug} endpoint through Public Gateway
   * Returns backend format: { data: {...}, success: boolean }
   */
  async getResearchArticleBySlug(slug: string): Promise<ResearchArticleResponse> {
    if (!slug) {
      throw new Error('Research article slug is required');
    }

    const endpoint = `/api/v1/research/slug/${encodeURIComponent(slug)}`;
    
    return this.request<ResearchArticleResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get research article by ID
   * Maps to GET /api/v1/research/{id} endpoint through Public Gateway
   * Returns backend format: { data: {...}, success: boolean }
   */
  async getResearchArticleById(id: string): Promise<ResearchArticleResponse> {
    if (!id) {
      throw new Error('Research article ID is required');
    }

    const endpoint = `/api/v1/research/${encodeURIComponent(id)}`;
    
    return this.request<ResearchArticleResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get featured research articles (published articles)
   * Maps to GET /api/v1/research/featured endpoint through Public Gateway
   * Returns backend format: { data: [...], pagination: {...}, success: boolean }
   */
  async getFeaturedResearch(limit?: number): Promise<ResearchResponse> {
    const queryParams = new URLSearchParams();
    if (limit !== undefined) queryParams.set('limit', limit.toString());
    
    const endpoint = `/api/v1/research/featured${queryParams.toString() ? `?${queryParams}` : ''}`;
    
    return this.request<ResearchResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Search research articles
   * Uses GET /api/v1/research/search with search parameter
   * Returns backend format: { data: [...], pagination: {...}, success: boolean }
   */
  async searchResearch(params: SearchResearchParams): Promise<ResearchResponse> {
    const queryParams = new URLSearchParams();
    
    queryParams.set('q', params.q);
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const endpoint = `/api/v1/research/search?${queryParams}`;
    
    return this.request<ResearchResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get recent research articles
   * Uses GET /api/v1/research with sortBy parameter
   * Returns backend format: { data: [...], pagination: {...}, success: boolean }
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
   * Maps to GET /api/v1/research/categories/{category}/articles endpoint through Public Gateway
   * Returns backend format: { data: [...], pagination: {...}, success: boolean }
   */
  async getResearchByCategory(category: string, params?: Partial<GetResearchParams>): Promise<ResearchResponse> {
    if (!category) {
      throw new Error('Category is required');
    }

    // Handle category filtering by using different endpoint
    const endpoint = `/api/v1/research/categories/${encodeURIComponent(category)}/articles`;
    return this.request<ResearchResponse>(endpoint, { method: 'GET' });
  }

  /**
   * Get research articles by industry
   * Uses GET /api/v1/research with industry parameter
   * Returns backend format: { data: [...], pagination: {...}, success: boolean }
   */
  async getResearchByIndustry(industry: string, params?: Partial<GetResearchParams>): Promise<ResearchResponse> {
    if (!industry) {
      throw new Error('Industry is required');
    }

    const queryParams = new URLSearchParams();
    queryParams.set('industry', industry);
    if (params?.page !== undefined) queryParams.set('page', params.page.toString());
    if (params?.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params?.sortBy) queryParams.set('sortBy', params.sortBy);

    const endpoint = `/api/v1/research?${queryParams}`;
    
    return this.request<ResearchResponse>(endpoint, {
      method: 'GET',
    });
  }
}