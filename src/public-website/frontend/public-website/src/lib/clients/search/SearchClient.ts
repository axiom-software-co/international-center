// Unified Search REST Client - Cross-domain search functionality
// Provides unified search across Services, News, Events, and Research domains

import { BaseRestClient } from '../rest/BaseRestClient';
import { config } from '../../environments';
import type {
  SearchResult,
  SearchResponse,
  SearchParams,
  UnifiedSearchResponse,
  SearchIndexRefreshResponse,
} from './types';

export class SearchClient extends BaseRestClient {
  constructor() {
    // Handle test environment or missing configuration
    const searchConfig = config.domains?.search || {
      baseUrl: 'http://localhost:7230', // Search Service Gateway URL fallback
      timeout: 5000,
      retryAttempts: 2,
    };
    
    super({
      baseUrl: searchConfig.baseUrl,
      timeout: searchConfig.timeout,
      retryAttempts: searchConfig.retryAttempts,
    });
  }

  /**
   * Unified search across all content types
   */
  async search(params: SearchParams): Promise<UnifiedSearchResponse> {
    const queryParams = new URLSearchParams();
    
    // Add required query parameter
    queryParams.set('q', params.q);
    
    // Add optional parameters
    if (params.page) queryParams.set('page', params.page.toString());
    if (params.pageSize) queryParams.set('pageSize', params.pageSize.toString());
    if (params.content_type) queryParams.set('content_type', params.content_type);
    if (params.category) queryParams.set('category', params.category);
    if (params.sortBy) queryParams.set('sortBy', params.sortBy);

    const endpoint = `/api/v1/search?${queryParams}`;
    
    return this.request<UnifiedSearchResponse>(endpoint, {
      method: 'GET',
    });
  }

  /**
   * Search within specific content type
   */
  async searchByType(
    query: string, 
    contentType: 'service' | 'news' | 'research' | 'events',
    options: {
      pageSize?: number;
      category?: string;
      sortBy?: string;
    } = {}
  ): Promise<UnifiedSearchResponse> {
    const params: SearchParams = {
      q: query,
      content_type: contentType,
      pageSize: options.pageSize || 10,
      category: options.category,
      sortBy: options.sortBy as any,
    };

    return this.search(params);
  }

  /**
   * Get search suggestions/autocomplete
   */
  async getSuggestions(query: string, limit: number = 5): Promise<string[]> {
    const queryParams = new URLSearchParams();
    queryParams.set('q', query);
    queryParams.set('limit', limit.toString());

    const endpoint = `/api/v1/search/suggestions?${queryParams}`;
    
    try {
      const response = await this.request<{ suggestions: string[] }>(endpoint, {
        method: 'GET',
      });
      
      return response.suggestions || [];
    } catch (error) {
      console.error('Error getting search suggestions:', error);
      return [];
    }
  }

  /**
   * Refresh search index (admin function)
   */
  async refreshIndex(): Promise<SearchIndexRefreshResponse> {
    const endpoint = '/api/v1/search/refresh-index';
    
    return this.request<SearchIndexRefreshResponse>(endpoint, {
      method: 'POST',
    });
  }

  /**
   * Get search analytics/stats
   */
  async getSearchStats(): Promise<{
    totalDocuments: number;
    lastIndexUpdate: string;
    popularQueries: Array<{ query: string; count: number }>;
  }> {
    const endpoint = '/api/v1/search/stats';
    
    return this.request<{
      totalDocuments: number;
      lastIndexUpdate: string;
      popularQueries: Array<{ query: string; count: number }>;
    }>(endpoint, {
      method: 'GET',
    });
  }
}

// Export singleton instance
export const searchClient = new SearchClient();