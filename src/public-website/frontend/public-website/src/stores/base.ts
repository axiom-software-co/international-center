// Base Store Utilities
// Consolidated common patterns for domain stores

import type { CacheOptions } from './interfaces';

// Common cache timeout - 5 minutes
export const CACHE_TIMEOUT = 5 * 60 * 1000;

// Base state interface for all domain stores
export interface BaseStoreState {
  loading: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
  searchTotal: number;
  cacheKey: string | null;
  lastCacheTime: number;
}

// Base state factory
export const createBaseState = (): BaseStoreState => ({
  loading: false,
  error: null,
  total: 0,
  page: 1,
  pageSize: 10,
  searchTotal: 0,
  cacheKey: null,
  lastCacheTime: 0,
});

// Common getters that all stores use
export const createBaseGetters = () => ({
  totalPages(): number {
    return Math.ceil(this.total / this.pageSize) || 0;
  },
});

// Cache management utilities
export const createCacheActions = (domainPrefix: string) => ({
  generateCacheKey(params?: any): string {
    return `${domainPrefix}_${JSON.stringify(params || {})}`;
  },

  isCacheValid(cacheKey: string, cacheTimeout = CACHE_TIMEOUT): boolean {
    return this.cacheKey === cacheKey && 
           this.lastCacheTime > 0 && 
           (Date.now() - this.lastCacheTime) < cacheTimeout;
  },

  setCacheData(cacheKey: string): void {
    this.cacheKey = cacheKey;
    this.lastCacheTime = Date.now();
  },

  invalidateCache(): void {
    this.cacheKey = null;
    this.lastCacheTime = 0;
  },
});

// State management utilities
export const createStateActions = () => ({
  setLoading(loading: boolean): void {
    this.loading = loading;
  },

  setError(error: string | null): void {
    this.error = error;
  },

  clearError(): void {
    this.error = null;
  },
});

// API action wrapper for consistent error handling and loading states
export const withApiAction = async <T>(
  context: { setLoading: (loading: boolean) => void; setError: (error: string | null) => void; clearError: () => void },
  apiCall: () => Promise<T>,
  errorMessage: string = 'Operation failed'
): Promise<T | null> => {
  try {
    context.setLoading(true);
    context.clearError();
    return await apiCall();
  } catch (err) {
    const message = err instanceof Error ? err.message : errorMessage;
    context.setError(message);
    return null;
  } finally {
    context.setLoading(false);
  }
};

// Cached API action wrapper
export const withCachedApiAction = async <T>(
  context: any, // Store context with cache methods
  params: any,
  options: CacheOptions,
  apiCall: () => Promise<T>,
  onSuccess: (result: T) => void,
  onError: (items: any[], count: number) => void,
  errorMessage: string = 'Operation failed'
): Promise<void> => {
  const { useCache = false, cacheTimeout = CACHE_TIMEOUT } = options;
  const cacheKey = context.generateCacheKey(params);

  // Check cache if requested
  if (useCache && context.isCacheValid(cacheKey, cacheTimeout)) {
    return;
  }

  try {
    context.setLoading(true);
    context.clearError();

    const result = await apiCall();
    onSuccess(result);
    context.setCacheData(cacheKey);
  } catch (err) {
    const message = err instanceof Error ? err.message : errorMessage;
    context.setError(message);
    onError([], 0);
  } finally {
    context.setLoading(false);
  }
};

// Helper for creating domain-specific "by category" getters
export const createByCategoryGetter = <TItem extends { category_id: string }>(
  itemsField: string
) => {
  return function(this: any): Record<string, TItem[]> {
    return this[itemsField].reduce((acc: Record<string, TItem[]>, item: TItem) => {
      const categoryId = item.category_id;
      if (!acc[categoryId]) {
        acc[categoryId] = [];
      }
      acc[categoryId].push(item);
      return acc;
    }, {});
  };
};

// Search helper - clears results for empty queries
export const handleEmptySearch = <TItem>(
  query: string,
  setSearchResults: (results: TItem[], total: number) => void
): boolean => {
  if (!query || query.trim() === '') {
    setSearchResults([], 0);
    return true; // Indicates search was handled (empty query)
  }
  return false; // Indicates search should proceed
};

// Generic state setters factory
export const createDomainStateSetters = <TItem, TCategory>(itemsField: string, categoriesField: string, featuredField: string) => ({
  [`set${itemsField.charAt(0).toUpperCase() + itemsField.slice(1)}`]: function(items: TItem[], total: number, page: number, pageSize: number) {
    this[itemsField] = items;
    this.total = total;
    this.page = page;
    this.pageSize = pageSize;
  },

  setCategories(categories: TCategory[]): void {
    this[categoriesField] = categories;
  },

  [`setFeatured${itemsField.charAt(0).toUpperCase() + itemsField.slice(1)}`]: function(items: TItem[]) {
    this[featuredField] = items;
  },

  setSearchResults(results: TItem[], total: number): void {
    this.searchResults = results;
    this.searchTotal = total;
  },
});

// Generic getters factory
export const createDomainGetters = <TItem>(itemsField: string) => ({
  [`has${itemsField.charAt(0).toUpperCase() + itemsField.slice(1)}`]: function(): boolean {
    return this[itemsField].length > 0;
  },

  [`${itemsField}ByCategory`]: createByCategoryGetter<TItem>(itemsField),
});

// Generic grouping getter factory
export const createGroupingGetter = <TItem>(itemsField: string, groupByField: keyof TItem) => {
  return function(this: any): Record<string, TItem[]> {
    return this[itemsField].reduce((acc: Record<string, TItem[]>, item: TItem) => {
      const groupValue = String(item[groupByField]);
      if (!acc[groupValue]) {
        acc[groupValue] = [];
      }
      acc[groupValue].push(item);
      return acc;
    }, {});
  };
};