// News Composables - Database Schema Compliant
// Updated to work with TABLES-NEWS.md aligned types

import { ref, computed, onMounted, watch, type Ref } from 'vue';
import { newsClient } from '../lib/clients';
import type { NewsArticle, NewsCategory, GetNewsParams, SearchNewsParams } from '../lib/clients/news/types';

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

export function useNews(options: UseNewsOptions = {}): UseNewsResult {
  const { enabled = true, immediate = true, ...params } = options;

  const news = ref<NewsArticle[]>([]);
  const loading = ref(enabled && immediate);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);

  const totalPages = computed(() => {
    return Math.ceil(total.value / pageSize.value) || 0;
  });

  const fetchNews = async () => {
    if (!enabled) return;

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { news: [...], count: number, correlation_id: string }
      const response = await newsClient.getNews(params);

      news.value = response.news;
      total.value = response.count;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch news';
      error.value = errorMessage;
      console.error('Error fetching news:', err);
      news.value = [];
      total.value = 0;
    } finally {
      loading.value = false;
    }
  };

  // Watch for parameter changes
  watch(() => params, fetchNews, { deep: true });
  
  if (immediate) {
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
}

export interface UseNewsArticleResult {
  news: Ref<NewsArticle | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useNewsArticle(slug: Ref<string | null> | string | null): UseNewsArticleResult {
  const slugRef = typeof slug === 'string' ? ref(slug) : slug || ref(null);
  
  const news = ref<NewsArticle | null>(null);
  const loading = ref(!!slugRef.value);
  const error = ref<string | null>(null);

  const fetchNews = async () => {
    if (!slugRef.value) {
      news.value = null;
      loading.value = false;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { news: {...}, correlation_id: string }
      const response = await newsClient.getNewsArticleBySlug(slugRef.value);
      
      news.value = response.news;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch news article';
      error.value = errorMessage;
      console.error('Error fetching news article:', err);
      news.value = null;
    } finally {
      loading.value = false;
    }
  };

  // Watch for slug changes
  watch(slugRef, fetchNews, { immediate: true });

  return {
    news,
    loading,
    error,
    refetch: fetchNews,
  };
}

export interface UseFeaturedNewsResult {
  news: Ref<NewsArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useFeaturedNews(limit?: Ref<number | undefined> | number | undefined): UseFeaturedNewsResult {
  const limitRef = typeof limit === 'number' ? ref(limit) : limit || ref(undefined);
  
  const news = ref<NewsArticle[]>([]);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchFeaturedNews = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { news: [...], count: number, correlation_id: string }
      const response = await newsClient.getFeaturedNews(limitRef.value);
      
      news.value = response.news;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured news';
      error.value = errorMessage;
      console.error('Error fetching featured news:', err);
      news.value = [];
    } finally {
      loading.value = false;
    }
  };

  // Watch for limit changes
  watch(limitRef, fetchFeaturedNews, { immediate: true });

  onMounted(fetchFeaturedNews);

  return {
    news,
    loading,
    error,
    refetch: fetchFeaturedNews,
  };
}

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

export function useSearchNews(): UseSearchNewsResult {
  const results = ref<NewsArticle[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);

  const totalPages = computed(() => {
    return Math.ceil(total.value / pageSize.value) || 0;
  });

  const search = async (query: string, options: Partial<SearchNewsParams> = {}) => {
    // Handle empty queries
    if (!query || query.trim() === '') {
      results.value = [];
      total.value = 0;
      page.value = 1;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      const searchParams: SearchNewsParams = {
        q: query.trim(),
        page: options.page || 1,
        pageSize: options.pageSize || 10,
        ...options,
      };

      // Backend returns: { news: [...], count: number, correlation_id: string }
      const response = await newsClient.searchNews(searchParams);

      results.value = response.news;
      total.value = response.count;
      page.value = searchParams.page!;
      pageSize.value = searchParams.pageSize!;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to search news';
      error.value = errorMessage;
      console.error('Error searching news:', err);
      results.value = [];
      total.value = 0;
    } finally {
      loading.value = false;
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
}

export interface UseNewsCategoriesResult {
  categories: Ref<NewsCategory[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useNewsCategories(): UseNewsCategoriesResult {
  const categories = ref<NewsCategory[]>([]);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchCategories = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { categories: [...], count: number, correlation_id: string }
      const response = await newsClient.getNewsCategories();
      
      categories.value = response.categories;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch news categories';
      error.value = errorMessage;
      console.error('Error fetching news categories:', err);
      categories.value = [];
    } finally {
      loading.value = false;
    }
  };

  onMounted(fetchCategories);

  return {
    categories,
    loading,
    error,
    refetch: fetchCategories,
  };
}