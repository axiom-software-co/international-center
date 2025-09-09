// Generic REST Client Factory
// Generates domain-specific REST client methods to eliminate code duplication

import { BaseRestClient } from './BaseRestClient';

// Generic interfaces for domain clients
export interface DomainRestClientConfig {
  domain: string; // 'events', 'news', 'services', 'research'
  baseEndpoint: string; // '/api/v1/events', '/api/v1/news', etc.
  itemsField: string; // 'events', 'news', 'services', 'research'
  cacheTTL: {
    categories: number;
    featured: number;
    detail: number;
    list: number;
  };
}

// Query parameter building factory
export const createQueryParams = (params: Record<string, any>): URLSearchParams => {
  const queryParams = new URLSearchParams();
  
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      if (typeof value === 'boolean') {
        queryParams.set(key, value.toString());
      } else {
        queryParams.set(key, String(value));
      }
    }
  });
  
  return queryParams;
};

// Generic GET list method factory
export const createGetListMethod = <TParams, TResponse>(
  client: BaseRestClient,
  config: DomainRestClientConfig
) => {
  return async function(params: TParams = {} as TParams): Promise<TResponse> {
    const queryParams = createQueryParams(params as Record<string, any>);
    const endpoint = `${config.baseEndpoint}${queryParams.toString() ? `?${queryParams}` : ''}`;
    const cacheKey = `${config.domain}:${endpoint}`;
    
    return (client as any).requestWithCache<TResponse>(
      endpoint, 
      { method: 'GET' }, 
      cacheKey, 
      config.cacheTTL.list
    );
  };
};

// Generic GET by slug method factory
export const createGetBySlugMethod = <TResponse>(
  client: BaseRestClient,
  config: DomainRestClientConfig,
  itemName: string
) => {
  return async function(slug: string): Promise<TResponse> {
    if (!slug) {
      throw new Error(`${itemName} slug is required`);
    }

    const endpoint = `${config.baseEndpoint}/slug/${encodeURIComponent(slug)}`;
    const cacheKey = `${config.domain}:slug:${slug}`;
    
    return (client as any).requestWithCache<TResponse>(
      endpoint,
      { method: 'GET' },
      cacheKey,
      config.cacheTTL.detail
    );
  };
};

// Generic GET by ID method factory
export const createGetByIdMethod = <TResponse>(
  client: BaseRestClient,
  config: DomainRestClientConfig,
  itemName: string,
  idField: string = 'id'
) => {
  return async function(id: string): Promise<TResponse> {
    if (!id) {
      throw new Error(`${itemName} ID is required`);
    }

    const endpoint = `${config.baseEndpoint}/${encodeURIComponent(id)}`;
    const cacheKey = `${config.domain}:id:${id}`;
    
    return (client as any).requestWithCache<TResponse>(
      endpoint,
      { method: 'GET' },
      cacheKey,
      config.cacheTTL.detail
    );
  };
};

// Generic GET featured method factory
export const createGetFeaturedMethod = <TResponse>(
  client: BaseRestClient,
  config: DomainRestClientConfig,
  featuredEndpoint: string = '/published'
) => {
  return async function(limit?: number): Promise<TResponse> {
    const queryParams = new URLSearchParams();
    if (limit !== undefined) queryParams.set('limit', limit.toString());
    
    const endpoint = `${config.baseEndpoint}${featuredEndpoint}${queryParams.toString() ? `?${queryParams}` : ''}`;
    const cacheKey = `${config.domain}:featured${limit ? `:${limit}` : ''}`;
    
    return (client as any).requestWithCache<TResponse>(
      endpoint,
      { method: 'GET' },
      cacheKey,
      config.cacheTTL.featured
    );
  };
};

// Generic search method factory
export const createSearchMethod = <TSearchParams, TResponse>(
  client: BaseRestClient,
  config: DomainRestClientConfig,
  useSearchEndpoint: boolean = false
) => {
  return async function(params: TSearchParams): Promise<TResponse> {
    const queryParams = createQueryParams(params as Record<string, any>);
    
    let endpoint: string;
    if (useSearchEndpoint) {
      endpoint = `${config.baseEndpoint}/search?${queryParams}`;
    } else {
      // Convert 'q' to 'search' parameter for main endpoint
      const searchParams = new URLSearchParams(queryParams);
      if (searchParams.has('q')) {
        const query = searchParams.get('q');
        searchParams.delete('q');
        searchParams.set('search', query!);
      }
      endpoint = `${config.baseEndpoint}?${searchParams}`;
    }
    
    return (client as any).request<TResponse>(endpoint, { method: 'GET' });
  };
};

// Generic GET categories method factory
export const createGetCategoriesMethod = <TResponse>(
  client: BaseRestClient,
  config: DomainRestClientConfig
) => {
  return async function(): Promise<TResponse> {
    const endpoint = `${config.baseEndpoint}/categories`;
    const cacheKey = `${config.domain}:categories`;
    
    return (client as any).requestWithCache<TResponse>(
      endpoint,
      { method: 'GET' },
      cacheKey,
      config.cacheTTL.categories
    );
  };
};

// Generic GET recent method factory
export const createGetRecentMethod = <TResponse>(
  client: BaseRestClient,
  config: DomainRestClientConfig
) => {
  return async function(limit: number = 5): Promise<TResponse> {
    const queryParams = new URLSearchParams();
    queryParams.set('limit', limit.toString());
    queryParams.set('sortBy', 'date-desc');
    
    const endpoint = `${config.baseEndpoint}?${queryParams}`;
    
    return (client as any).request<TResponse>(endpoint, { method: 'GET' });
  };
};

// Generic GET by category method factory  
export const createGetByCategoryMethod = <TResponse>(
  client: BaseRestClient,
  config: DomainRestClientConfig
) => {
  return async function(category_id: string, params?: Record<string, any>): Promise<TResponse> {
    if (!category_id) {
      throw new Error('Category is required');
    }

    const endpoint = `${config.baseEndpoint}/categories/${encodeURIComponent(category_id)}/${config.itemsField}`;
    
    return (client as any).request<TResponse>(endpoint, { method: 'GET' });
  };
};

// CRUD operations factory
export const createCRUDMethods = <TCreateParams, TUpdateParams, TResponse>(
  client: BaseRestClient,
  config: DomainRestClientConfig
) => ({
  create: async function(params: TCreateParams): Promise<TResponse> {
    const endpoint = config.baseEndpoint;
    
    try {
      const result = await (client as any).request<TResponse>(endpoint, {
        method: 'POST',
        body: JSON.stringify(params),
        headers: { 'Content-Type': 'application/json' },
      });
      
      // Invalidate relevant caches
      (client as any).invalidateCachePattern(`${config.domain}:`);
      
      return result;
    } catch (error) {
      throw error;
    }
  },

  update: async function(params: TUpdateParams & { [key: string]: any }): Promise<TResponse> {
    const { [Object.keys(params).find(k => k.includes('_id')) || 'id']: id, ...updateData } = params;
    const endpoint = `${config.baseEndpoint}/${encodeURIComponent(id)}`;
    
    try {
      const result = await (client as any).request<TResponse>(endpoint, {
        method: 'PUT',
        body: JSON.stringify(updateData),
        headers: { 'Content-Type': 'application/json' },
      });
      
      // Invalidate specific item and list caches
      (client as any).invalidateCacheKey(`${config.domain}:id:${id}`);
      (client as any).invalidateCachePattern(`${config.domain}:`);
      
      return result;
    } catch (error) {
      throw error;
    }
  },

  delete: async function(id: string): Promise<void> {
    if (!id) {
      throw new Error(`${config.domain} ID is required`);
    }

    const endpoint = `${config.baseEndpoint}/${encodeURIComponent(id)}`;
    
    try {
      await (client as any).request<void>(endpoint, { method: 'DELETE' });
      
      // Invalidate caches
      (client as any).invalidateCacheKey(`${config.domain}:id:${id}`);
      (client as any).invalidateCachePattern(`${config.domain}:`);
    } catch (error) {
      throw error;
    }
  }
});