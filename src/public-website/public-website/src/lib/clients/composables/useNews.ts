// News Composables - Vue 3 Composition API composables for news data
// Provides clean interface for Vue components to interact with news domain
// Following the same pattern as useServices

import { ref, computed, onMounted, watch, type Ref } from 'vue';
import { NewsRestClient } from '../rest/NewsRestClient';
import type { NewsArticle, GetNewsParams, NewsCategory } from '../news/types';

// Create singleton instance
const newsClient = new NewsRestClient();

export interface UseNewsResult {
  articles: Ref<NewsArticle[]>;
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

  const articles = ref<NewsArticle[]>([]);
  const loading = ref(enabled && immediate);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);
  const totalPages = ref(0);

  const fetchNews = async () => {
    if (!enabled) return;

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { news: [...], count: number, correlation_id: string }
      const response = await newsClient.getNews(params);

      if (response.news) {
        articles.value = response.news;
        total.value = response.count || response.news.length;
        // Calculate pagination from current params since backend doesn't return pagination info
        page.value = params.page || 1;
        pageSize.value = params.pageSize || 10;
        totalPages.value = Math.ceil(total.value / pageSize.value);
      } else {
        throw new Error('Failed to fetch news articles');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch news articles';
      error.value = errorMessage;
      console.error('Error fetching news articles:', err);
      articles.value = [];
      total.value = 0;
      totalPages.value = 0;
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
    articles,
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
  article: Ref<NewsArticle | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useNewsArticle(slug: Ref<string | null> | string | null): UseNewsArticleResult {
  const slugRef = typeof slug === 'string' ? ref(slug) : slug || ref(null);
  
  const article = ref<NewsArticle | null>(null);
  const loading = ref(!!slugRef.value);
  const error = ref<string | null>(null);

  const fetchArticle = async () => {
    if (!slugRef.value) {
      article.value = null;
      loading.value = false;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { news: {...}, correlation_id: string }
      const response = await newsClient.getNewsBySlug(slugRef.value);
      
      if (response.news) {
        article.value = response.news;
      } else {
        throw new Error('Failed to fetch news article');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch news article';
      error.value = errorMessage;
      console.error('Error fetching news article:', err);
      article.value = null;
    } finally {
      loading.value = false;
    }
  };

  // Watch for slug changes
  watch(slugRef, fetchArticle, { immediate: true });

  return {
    article,
    loading,
    error,
    refetch: fetchArticle,
  };
}

export interface UseFeaturedNewsResult {
  articles: Ref<NewsArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useFeaturedNews(limit?: Ref<number> | number): UseFeaturedNewsResult {
  const limitRef = typeof limit === 'number' ? ref(limit) : limit;
  
  const articles = ref<NewsArticle[]>([]);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchFeaturedNews = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { news: [...], count: number, correlation_id: string }
      const response = await newsClient.getFeaturedNews(limitRef?.value);
      
      if (response.news) {
        articles.value = response.news;
      } else {
        throw new Error('Failed to fetch featured news articles');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured news articles';
      error.value = errorMessage;
      console.error('Error fetching featured news articles:', err);
      articles.value = [];
    } finally {
      loading.value = false;
    }
  };

  // Watch for limit changes
  if (limitRef) {
    watch(limitRef, fetchFeaturedNews, { immediate: true });
  } else {
    onMounted(fetchFeaturedNews);
  }

  return {
    articles,
    loading,
    error,
    refetch: fetchFeaturedNews,
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

  const fetchNewsCategories = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { categories: [...], count: number, correlation_id: string }
      const response = await newsClient.getNewsCategories();
      
      if (response.categories) {
        categories.value = response.categories;
      } else {
        throw new Error('Failed to fetch news categories');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch news categories';
      error.value = errorMessage;
      console.error('Error fetching news categories:', err);
      categories.value = [];
    } finally {
      loading.value = false;
    }
  };

  onMounted(fetchNewsCategories);

  return {
    categories,
    loading,
    error,
    refetch: fetchNewsCategories,
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
  search: (query: string, options?: Partial<GetNewsParams>) => Promise<void>;
}

export function useSearchNews(): UseSearchNewsResult {
  const results = ref<NewsArticle[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);
  const totalPages = ref(0);

  const search = async (query: string, options: Partial<GetNewsParams> = {}) => {
    if (!query.trim()) {
      results.value = [];
      total.value = 0;
      totalPages.value = 0;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { news: [...], count: number, correlation_id: string }
      const response = await newsClient.searchNews({
        q: query,
        page: options.page || 1,
        pageSize: options.pageSize || 10,
        category: options.category,
        ...options,
      });

      if (response.news) {
        results.value = response.news;
        total.value = response.count || response.news.length;
        page.value = options.page || 1;
        pageSize.value = options.pageSize || 10;
        totalPages.value = Math.ceil(total.value / pageSize.value);
      } else {
        throw new Error('Failed to search news articles');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to search news articles';
      error.value = errorMessage;
      console.error('Error searching news articles:', err);
      results.value = [];
      total.value = 0;
      totalPages.value = 0;
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