// Research Composables - Vue 3 Composition API with Store Integration
// Refactored to use explicit implementations and consistent store patterns

import { ref, computed, watch, onMounted, isRef, unref, type Ref } from 'vue';
import { storeToRefs } from 'pinia';
import { useResearchStore } from '../stores/research';
import type { ResearchArticle, ResearchCategory, GetResearchParams, SearchResearchParams } from '../lib/clients/research/types';
import type { BaseComposableOptions } from './base';

// Domain-specific type aliases
export interface UseResearchArticlesResult {
  research: Ref<ResearchArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  refetch: () => Promise<void>;
}

export interface UseResearchArticlesOptions extends GetResearchParams, BaseComposableOptions {}

// Main research articles composable
export const useResearchArticles = (options: UseResearchArticlesOptions = {}): UseResearchArticlesResult => {
  const { enabled = true, immediate = true, ...params } = options;
  
  const store = useResearchStore();
  const { research, loading, error, total, totalPages } = storeToRefs(store);
  
  // Local pagination refs
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);

  const fetchItems = async () => {
    if (!enabled) return;
    
    try {
      // Disable caching in test environment to ensure API calls are made
      const shouldUseCache = import.meta.env?.VITEST !== true;
      await store.fetchResearch(params, { useCache: shouldUseCache });
    } catch (err) {
      // Error handling managed by store
    }
  };

  // Watch for parameter changes
  watch(() => params, fetchItems, { deep: true });
  
  // Call immediately if enabled and immediate is true
  // Direct call since we're not always in a component context during tests
  if (enabled && immediate) {
    fetchItems();
  }

  return {
    research,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    refetch: fetchItems,
  };
};

// Legacy compatibility - map research field to articles
export function useResearch(options: UseResearchArticlesOptions = {}) {
  const result = useResearchArticles(options);
  return {
    ...result,
    articles: result.research, // Map research to articles for backward compatibility
  };
}

export interface UseResearchArticleResult {
  article: Ref<ResearchArticle | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Single research article composable
export const useResearchArticle = (slug: Ref<string | null> | string | null): UseResearchArticleResult => {
  const slugRef = isRef(slug) ? slug : ref(slug);
  const store = useResearchStore();
  const { loading, error } = storeToRefs(store);
  
  const article = ref<ResearchArticle | null>(null);

  const fetchItem = async () => {
    const currentSlug = unref(slugRef);
    if (!currentSlug) {
      article.value = null;
      return;
    }

    try {
      const result = await store.fetchResearchArticle(currentSlug);
      article.value = result;
    } catch (err) {
      article.value = null;
    }
  };

  // Watch for slug changes
  watch(slugRef, (newSlug) => {
    if (newSlug) {
      fetchItem();
    } else {
      article.value = null;
    }
  }, { immediate: true });

  return {
    article,
    loading,
    error,
    refetch: fetchItem,
  };
};

export interface UseFeaturedResearchResult {
  research: Ref<ResearchArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Featured research composable
export const useFeaturedResearch = (limit?: Ref<number | undefined> | number | undefined): UseFeaturedResearchResult => {
  const limitRef = isRef(limit) ? limit : ref(limit);
  const store = useResearchStore();
  const { featuredResearch: research, loading, error } = storeToRefs(store);

  const fetchFeaturedItems = async () => {
    try {
      await store.fetchFeaturedResearch(unref(limitRef));
    } catch (err) {
      // Error handling managed by store
    }
  };

  // Watch for limit changes with immediate execution
  watch(limitRef, fetchFeaturedItems, { immediate: true });

  return {
    research,
    loading,
    error,
    refetch: fetchFeaturedItems,
  };
};

// Legacy compatibility - map research field to articles
export function useFeaturedResearchArticles(): { articles: Ref<ResearchArticle[]>; loading: Ref<boolean>; error: Ref<string | null>; refetch: () => Promise<void> } {
  const { research, loading, error, refetch } = useFeaturedResearch();
  
  return {
    articles: research, // Map research to articles for backward compatibility
    loading,
    error,
    refetch,
  };
}

export interface UseSearchResearchResult {
  results: Ref<ResearchArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  search: (query: string, options?: Partial<SearchResearchParams>) => Promise<void>;
}

// Search research composable
export const useSearchResearch = (): UseSearchResearchResult => {
  const store = useResearchStore();
  const { searchResults: results, loading, error, searchTotal: total } = storeToRefs(store);
  
  // Local refs for search-specific pagination
  const page = ref(1);
  const pageSize = ref(10);

  const totalPages = computed(() => {
    return Math.ceil(total.value / pageSize.value) || 0;
  });

  const search = async (query: string, options: Partial<SearchResearchParams> = {}) => {
    const searchParams = {
      q: query,
      page: options.page || 1,
      pageSize: options.pageSize || 10,
      ...options,
    };

    // Update local pagination refs
    page.value = searchParams.page;
    pageSize.value = searchParams.pageSize;

    try {
      await store.searchResearch(searchParams);
    } catch (err) {
      // Error handling is managed by the store
    }
  };

  return {
    results,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    search,
  };
};

export interface UseResearchCategoriesResult {
  categories: Ref<ResearchCategory[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Research categories composable
export const useResearchCategories = (): UseResearchCategoriesResult => {
  const store = useResearchStore();
  const { categories, loading, error } = storeToRefs(store);

  const fetchCategories = async () => {
    try {
      await store.fetchResearchCategories();
    } catch (err) {
      // Error handling managed by store
    }
  };

  // Initial fetch - trigger immediately since we're not in a component context during tests
  fetchCategories();

  return {
    categories,
    loading,
    error,
    refetch: fetchCategories,
  };
};