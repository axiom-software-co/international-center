// News Composables - Vue 3 Composition API with Store Delegation

import { ref, computed, watch, onMounted, isRef, unref, type Ref } from 'vue';
import { storeToRefs } from 'pinia';
import { useNewsStore } from '../stores/news';
import type { NewsArticle, NewsCategory, GetNewsParams, SearchNewsParams } from '../lib/clients/news/types';

// Domain-specific type aliases
export interface UseNewsResult {
  news: Ref<NewsArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  refetch: () => Promise<void>;
}

export interface UseNewsOptions extends GetNewsParams {
  enabled?: boolean;
  immediate?: boolean;
}

// Main news list composable - delegates to store
export const useNews = (options: UseNewsOptions = {}): UseNewsResult => {
  const { enabled = true, immediate = true, ...params } = options;
  const store = useNewsStore();
  const { news, loading, error, total, totalPages } = storeToRefs(store);
  
  // Local refs for pagination
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);

  const fetchNews = async () => {
    if (!enabled) return;
    await store.fetchNews(params);
  };

  // Watch for parameter changes
  watch(() => params, fetchNews, { deep: true });
  
  // Call immediately if enabled and immediate is true
  if (enabled && immediate) {
    onMounted(fetchNews);
  }

  return {
    news,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    refetch: fetchNews,
  };
};

export interface UseNewsArticleResult {
  news: Ref<NewsArticle | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Single news article composable - delegates to store
export const useNewsArticle = (slug: Ref<string | null> | string | null): UseNewsArticleResult => {
  const slugRef = isRef(slug) ? slug : ref(slug);
  const store = useNewsStore();
  const { article: news, loading, error } = storeToRefs(store);

  const fetchNewsArticle = async () => {
    const currentSlug = unref(slugRef);
    if (!currentSlug) {
      store.article = null;
      return;
    }

    await store.fetchNewsArticle(currentSlug);
  };

  // Watch for slug changes and call immediately
  watch(slugRef, (newSlug) => {
    if (newSlug) {
      fetchNewsArticle();
    } else {
      store.article = null;
    }
  }, { immediate: true });

  return {
    news,
    loading,
    error,
    refetch: fetchNewsArticle,
  };
};

export interface UseFeaturedNewsResult {
  news: Ref<NewsArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// Featured news composable - delegates to store
export const useFeaturedNews = (limit?: Ref<number | undefined> | number | undefined): UseFeaturedNewsResult => {
  const limitRef = isRef(limit) ? limit : ref(limit);
  const store = useNewsStore();
  const { featuredNews: news, loading, error } = storeToRefs(store);

  const fetchFeaturedNews = async () => {
    await store.fetchFeaturedNews(unref(limitRef));
  };

  // Trigger initial fetch immediately
  fetchFeaturedNews();
  
  // Watch for limit changes
  watch(limitRef, fetchFeaturedNews);

  return {
    news,
    loading,
    error,
    refetch: fetchFeaturedNews,
  };
};

export interface UseSearchNewsResult {
  results: Ref<NewsArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  search: (query: string, options?: Partial<SearchNewsParams>) => Promise<void>;
}

// Search news composable - delegates to store
export const useSearchNews = (): UseSearchNewsResult => {
  const store = useNewsStore();
  const { searchResults: results, loading, error, searchTotal: total } = storeToRefs(store);
  
  // Local refs for search-specific pagination
  const page = ref(1);
  const pageSize = ref(10);

  const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 0);

  const search = async (query: string, options: Partial<SearchNewsParams> = {}) => {
    const searchParams = {
      q: query,
      page: options.page || 1,
      pageSize: options.pageSize || 10,
      ...options,
    };

    // Update local pagination refs
    page.value = searchParams.page;
    pageSize.value = searchParams.pageSize;

    await store.searchNews(searchParams);
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

export interface UseNewsCategoriesResult {
  categories: Ref<NewsCategory[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

// News categories composable - delegates to store
export const useNewsCategories = (): UseNewsCategoriesResult => {
  const store = useNewsStore();
  const { categories, loading, error } = storeToRefs(store);

  const fetchCategories = async () => {
    await store.fetchNewsCategories();
  };

  // Trigger initial fetch immediately
  fetchCategories();

  return {
    categories,
    loading,
    error,
    refetch: fetchCategories,
  };
};