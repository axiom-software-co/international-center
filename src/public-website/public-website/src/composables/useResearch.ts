// Research Composables - Vue 3 Composition API with Store Delegation

import { ref, computed, watch, onMounted, isRef, unref, nextTick, type Ref } from 'vue';
import { storeToRefs } from 'pinia';
import { useResearchStore } from '../stores/research';
import type { ResearchArticle, ResearchCategory, GetResearchParams, SearchResearchParams } from '../lib/clients/research/types';

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

export interface UseResearchArticlesOptions extends GetResearchParams {
  enabled?: boolean;
  immediate?: boolean;
}

// Main research articles composable - delegates to store
export const useResearchArticles = (options: UseResearchArticlesOptions = {}): UseResearchArticlesResult => {
  const { enabled = true, immediate = true, ...params } = options;
  const store = useResearchStore();
  const { research, loading, error, total, totalPages } = storeToRefs(store);
  
  // Local refs for pagination
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);

  const fetchResearch = async () => {
    if (!enabled) return;
    await store.fetchResearch(params);
  };

  // Watch for parameter changes
  watch(() => params, fetchResearch, { deep: true });
  
  // Call immediately if enabled and immediate is true
  if (enabled && immediate) {
    onMounted(fetchResearch);
  }

  return {
    research,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    refetch: fetchResearch,
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

// Single research article composable - delegates to store
export const useResearchArticle = (slug: Ref<string | null> | string | null): UseResearchArticleResult => {
  const slugRef = isRef(slug) ? slug : ref(slug);
  const store = useResearchStore();
  const { article, loading, error } = storeToRefs(store);

  const fetchResearchArticle = async () => {
    const currentSlug = unref(slugRef);
    if (!currentSlug) {
      store.article = null;
      return;
    }

    await store.fetchResearchArticle(currentSlug);
  };

  // Watch for slug changes and call immediately
  watch(slugRef, (newSlug) => {
    if (newSlug) {
      fetchResearchArticle();
    } else {
      store.article = null;
    }
  }, { immediate: true });

  return {
    article,
    loading,
    error,
    refetch: fetchResearchArticle,
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