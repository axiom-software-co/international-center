import { BaseRestClient, RestClientConfig } from './BaseRestClient';
import { config } from '../../../environments';
import type { 
  NewsArticle, 
  GetNewsParams, 
  SearchNewsParams 
} from '../news/types';
import type {
  BackendNewsResponse,
  BackendSingleNewsResponse,
  BackendNewsCategoriesResponse,
  BackendFeaturedNewsResponse,
  BackendNewsParams,
  BackendSearchNewsParams
} from './BackendNewsTypes';

export class NewsRestClient extends BaseRestClient {
  constructor() {
    super({
      baseUrl: config.domains.news.baseUrl,
      timeout: config.domains.news.timeout,
      retryAttempts: config.domains.news.retryAttempts,
    });
  }

  /**
   * Get news articles list
   * Maps to GET /api/v1/news endpoint through Public Gateway
   * Returns backend format: { news: [...], count: number, correlation_id: string }
   */
  async getNews(params: GetNewsParams = {}): Promise<BackendNewsResponse> {
    // Handle category filtering by using different endpoint
    if (params.category) {
      const endpoint = `/api/v1/news/categories/${encodeURIComponent(params.category)}/news`;
      return this.request<BackendNewsResponse>(endpoint, { method: 'GET' });
    }

    const queryParams = new URLSearchParams();
    
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.search) queryParams.set('search', params.search);
    if (params.featured !== undefined) queryParams.set('featured', params.featured.toString());

    const endpoint = `/api/v1/news${queryParams.toString() ? `?${queryParams}` : ''}`;
    
    return this.request<BackendNewsResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get news article by slug
   * Maps to GET /api/v1/news/slug/{slug} endpoint through Public Gateway
   * Returns backend format: { news: {...}, correlation_id: string }
   */
  async getNewsBySlug(slug: string): Promise<BackendSingleNewsResponse> {
    if (!slug) {
      throw new Error('News article slug is required');
    }

    const endpoint = `/api/v1/news/slug/${encodeURIComponent(slug)}`;
    
    return this.request<BackendSingleNewsResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get news categories
   * Maps to GET /api/v1/news/categories endpoint through Public Gateway
   * Returns backend format: { categories: [...], count: number, correlation_id: string }
   */
  async getNewsCategories(): Promise<BackendNewsCategoriesResponse> {
    const endpoint = '/api/v1/news/categories';
    
    return this.request<BackendNewsCategoriesResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get featured news articles
   * Maps to GET /api/v1/news/featured endpoint through Public Gateway
   * Returns backend format: { news: [...], count: number, correlation_id: string }
   */
  async getFeaturedNews(limit?: number): Promise<BackendNewsResponse> {
    const queryParams = new URLSearchParams();
    if (limit !== undefined) queryParams.set('limit', limit.toString());

    const endpoint = `/api/v1/news/featured${queryParams.toString() ? `?${queryParams}` : ''}`;
    
    return this.request<BackendNewsResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Search news articles
   * Uses GET /api/v1/news/search with search parameter
   * Returns backend format: { news: [...], count: number, correlation_id: string }
   */
  async searchNews(params: SearchNewsParams): Promise<BackendNewsResponse> {
    const queryParams = new URLSearchParams();
    
    queryParams.set('q', params.q);
    if (params.page !== undefined) queryParams.set('page', params.page.toString());
    if (params.pageSize !== undefined) queryParams.set('pageSize', params.pageSize.toString());
    if (params.category) queryParams.set('category', params.category);

    const endpoint = `/api/v1/news/search?${queryParams}`;
    
    return this.request<BackendNewsResponse>(endpoint, {
      method: 'GET',
    });
  }
}