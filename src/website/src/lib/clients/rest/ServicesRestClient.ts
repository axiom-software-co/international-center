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

export class ServicesRestClient extends BaseRestClient {
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
    
    return this.request<BackendServicesResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get service by slug
   * Maps to GET /api/v1/services/slug/{slug} endpoint through Public Gateway
   * Returns backend format: { service: {...}, correlation_id: string }
   */
  async getServiceBySlug(slug: string): Promise<BackendServiceResponse> {
    if (!slug) {
      throw new Error('Service slug is required');
    }

    const endpoint = `/api/v1/services/slug/${encodeURIComponent(slug)}`;
    
    return this.request<BackendServiceResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get service categories
   * Maps to GET /api/v1/services/categories endpoint through Public Gateway
   * Returns backend format: { categories: [...], count: number, correlation_id: string }
   */
  async getServiceCategories(): Promise<BackendServiceCategoriesResponse> {
    const endpoint = '/api/v1/services/categories';
    
    return this.request<BackendServiceCategoriesResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Get featured services (published services)
   * Maps to GET /api/v1/services/published endpoint through Public Gateway
   * Returns backend format: { services: [...], count: number, correlation_id: string }
   */
  async getFeaturedServices(limit?: number): Promise<BackendServicesResponse> {
    const endpoint = '/api/v1/services/published';
    
    return this.request<BackendServicesResponse>(endpoint, {
      method: 'GET',
    });
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
}