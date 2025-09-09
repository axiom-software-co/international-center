// Unified Search Composable - Cross-domain search functionality
// Provides Vue 3 composition API for unified search across all content types

import { ref, computed, watch, type Ref } from 'vue';
import { searchService } from '../../lib/services/SearchService';
import type { 
  SearchResult, 
  SearchOptions, 
  UnifiedSearchResponse 
} from '../../lib/services/SearchService';

export interface UseUnifiedSearchResult {
  // Search results
  results: Ref<SearchResult[]>;
  total: Ref<number>;
  query: Ref<string>;
  suggestions: Ref<string[]>;
  
  // Faceted search
  facets: Ref<{
    types: Array<{ type: string; count: number }>;
    categories: Array<{ category: string; count: number }>;
  }>;
  
  // Pagination
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  
  // State
  loading: Ref<boolean>;
  error: Ref<string | null>;
  
  // Search history
  searchHistory: Ref<string[]>;
  
  // Actions
  search: (query: string, options?: SearchOptions) => Promise<void>;
  quickSearch: (query: string, limit?: number) => Promise<SearchResult[]>;
  searchByType: (query: string, type: 'service' | 'news' | 'research', options?: { limit?: number; category?: string }) => Promise<SearchResult[]>;
  getSuggestions: (query: string, limit?: number) => Promise<string[]>;
  clearResults: () => void;
  clearHistory: () => void;
}

export interface UseUnifiedSearchOptions {
  immediate?: boolean;
  initialQuery?: string;
  initialOptions?: SearchOptions;
}

/**
 * Unified search composable - search across all content types
 */
export function useUnifiedSearch(options: UseUnifiedSearchOptions = {}): UseUnifiedSearchResult {
  const { immediate = false, initialQuery = '', initialOptions = {} } = options;

  // State
  const results = ref<SearchResult[]>([]);
  const total = ref(0);
  const query = ref(initialQuery);
  const suggestions = ref<string[]>([]);
  const facets = ref<{
    types: Array<{ type: string; count: number }>;
    categories: Array<{ category: string; count: number }>;
  }>({ types: [], categories: [] });
  
  // Pagination
  const page = ref(initialOptions.page || 1);
  const pageSize = ref(initialOptions.pageSize || 20);
  const totalPages = ref(0);
  
  // Loading and error states
  const loading = ref(false);
  const error = ref<string | null>(null);
  
  // Search history
  const searchHistory = ref<string[]>([]);

  // Computed
  const computedTotalPages = computed(() => {
    return Math.ceil(total.value / pageSize.value) || 0;
  });

  // Watch for totalPages changes
  watch(computedTotalPages, (newTotalPages) => {
    totalPages.value = newTotalPages;
  });

  // Actions
  const search = async (searchQuery: string, searchOptions: SearchOptions = {}): Promise<void> => {
    if (!searchQuery.trim()) {
      clearResults();
      return;
    }

    try {
      loading.value = true;
      error.value = null;
      query.value = searchQuery;

      // Merge with current pagination state
      const options: SearchOptions = {
        page: page.value,
        pageSize: pageSize.value,
        ...searchOptions,
      };

      // Update pagination state from options
      if (options.page) page.value = options.page;
      if (options.pageSize) pageSize.value = options.pageSize;

      const response = await searchService.search(searchQuery, options);

      // Update state
      results.value = response.results;
      total.value = response.total;
      suggestions.value = response.suggestions || [];
      facets.value = response.facets;
      
      // Update search history
      updateSearchHistory();

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Search failed';
      error.value = errorMessage;
      console.error('Search error:', err);
      
      // Clear results on error
      clearResults();
    } finally {
      loading.value = false;
    }
  };

  const quickSearch = async (searchQuery: string, limit: number = 10): Promise<SearchResult[]> => {
    if (!searchQuery.trim()) {
      return [];
    }

    try {
      const quickResults = await searchService.quickSearch(searchQuery, limit);
      return quickResults;
    } catch (err) {
      console.error('Quick search error:', err);
      return [];
    }
  };

  const searchByType = async (
    searchQuery: string, 
    type: 'service' | 'news' | 'research',
    searchOptions: { limit?: number; category?: string } = {}
  ): Promise<SearchResult[]> => {
    if (!searchQuery.trim()) {
      return [];
    }

    try {
      const typeResults = await searchService.searchByType(searchQuery, type, searchOptions);
      return typeResults;
    } catch (err) {
      console.error('Type search error:', err);
      return [];
    }
  };

  const getSuggestionsAction = async (searchQuery: string, limit: number = 5): Promise<string[]> => {
    if (!searchQuery.trim()) {
      return [];
    }

    try {
      const searchSuggestions = await searchService.getSuggestions(searchQuery, limit);
      return searchSuggestions;
    } catch (err) {
      console.error('Suggestions error:', err);
      return [];
    }
  };

  const clearResults = (): void => {
    results.value = [];
    total.value = 0;
    query.value = '';
    suggestions.value = [];
    facets.value = { types: [], categories: [] };
    page.value = 1;
    totalPages.value = 0;
    error.value = null;
  };

  const clearHistory = (): void => {
    searchService.clearSearchHistory();
    updateSearchHistory();
  };

  // Helper functions
  const updateSearchHistory = (): void => {
    searchHistory.value = searchService.getSearchHistory();
  };

  // Initialize
  if (immediate && initialQuery) {
    search(initialQuery, initialOptions);
  }

  // Initialize search history
  updateSearchHistory();

  return {
    // Results
    results,
    total,
    query,
    suggestions,
    facets,
    
    // Pagination
    page,
    pageSize,
    totalPages,
    
    // State
    loading,
    error,
    searchHistory,
    
    // Actions
    search,
    quickSearch,
    searchByType,
    getSuggestions: getSuggestionsAction,
    clearResults,
    clearHistory,
  };
}

/**
 * Quick search composable - for instant search/autocomplete functionality
 */
export function useQuickSearch() {
  const loading = ref(false);
  const results = ref<SearchResult[]>([]);
  const suggestions = ref<string[]>([]);

  const quickSearch = async (query: string, limit: number = 5): Promise<void> => {
    if (!query.trim()) {
      results.value = [];
      return;
    }

    try {
      loading.value = true;
      const searchResults = await searchService.quickSearch(query, limit);
      results.value = searchResults;
    } catch (error) {
      console.error('Quick search error:', error);
      results.value = [];
    } finally {
      loading.value = false;
    }
  };

  const getSuggestions = async (query: string, limit: number = 5): Promise<void> => {
    if (!query.trim()) {
      suggestions.value = [];
      return;
    }

    try {
      const searchSuggestions = await searchService.getSuggestions(query, limit);
      suggestions.value = searchSuggestions;
    } catch (error) {
      console.error('Suggestions error:', error);
      suggestions.value = [];
    }
  };

  return {
    loading,
    results,
    suggestions,
    quickSearch,
    getSuggestions,
  };
}

/**
 * Search history composable
 */
export function useSearchHistory() {
  const history = ref<string[]>([]);

  const updateHistory = (): void => {
    history.value = searchService.getSearchHistory();
  };

  const clearHistory = (): void => {
    searchService.clearSearchHistory();
    updateHistory();
  };

  // Initialize
  updateHistory();

  return {
    history,
    updateHistory,
    clearHistory,
  };
}