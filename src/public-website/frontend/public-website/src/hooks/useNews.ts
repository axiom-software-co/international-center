// News Composable - Vue composable for news data using contract-generated clients
// Provides clean interface for components to interact with news domain

import { ref, computed, watch, type Ref } from 'vue';
import { useContractNews } from '../composables/useContractApi';
import { ContractErrorHandler } from '../lib/error-handling';
import type { NewsArticle, GetNews200Response, GetNewsRequest } from '@international-center/public-api-client';

export interface UseNewsResult {
  articles: Ref<NewsArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  hasNext: Ref<boolean>;
  hasPrevious: Ref<boolean>;
  refetch: () => Promise<void>;
}

export interface UseNewsOptions {
  page?: number;
  limit?: number;
  search?: string;
  categoryId?: string;
  enabled?: boolean;
}

export function useNews(options: UseNewsOptions = {}): UseNewsResult {
  const { enabled = true, ...params } = options;

  // Use contract-based composable for type-safe API operations
  const contractNews = useContractNews();
  
  const articles = ref<NewsArticle[]>([]);
  const loading = ref(enabled);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(params.page || 1);
  const pageSize = ref(params.limit || 10);
  const totalPages = ref(0);
  const hasNext = ref(false);
  const hasPrevious = ref(false);

  const fetchNews = async () => {
    if (!enabled) return;

    try {
      // Use contract composable for type-safe API calls with error handling
      const newsData = await contractNews.fetchNews({
        page: page.value,
        limit: pageSize.value,
        search: params.search,
        categoryId: params.categoryId,
      });

      // Extract data with contract type safety
      articles.value = newsData || [];
      total.value = newsData?.length || 0;
      // Simplified pagination for now - would be enhanced with actual pagination response
      totalPages.value = Math.ceil(total.value / pageSize.value) || 1;
      hasNext.value = page.value < totalPages.value;
      hasPrevious.value = page.value > 1;

    } catch (err) {
      const errorMessage = ContractErrorHandler.getUserFriendlyMessage(
        ContractErrorHandler.parseContractError(err), 
        'news'
      );
      error.value = errorMessage;
    }
    
    // Sync loading state with contract composable
    loading.value = contractNews.loading.value;
  };

  // Watch for parameter changes and refetch
  watch(() => [enabled, params.page, params.limit, params.search, params.categoryId], fetchNews, { immediate: true });

  return {
    articles,
    loading,
    error: computed(() => error.value || contractNews.error.value),
    total,
    page,
    pageSize,
    totalPages,
    hasNext,
    hasPrevious,
    refetch: fetchNews,
  };
}

export interface UseNewsArticleResult {
  article: Ref<NewsArticle | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useNewsArticle(slug: string | null): UseNewsArticleResult {
  const article = ref<NewsArticle | null>(null);
  const loading = ref(!!slug);
  const error = ref<string | null>(null);

  const fetchArticle = async () => {
    if (!slug) {
      article.value = null;
      loading.value = false;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Use contract-generated client - note: using ID instead of slug for now
      // In a full implementation, you'd add slug-based lookup to the API
      const response = await apiClient.getNewsById(slug);
      article.value = response.data || null;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch news article';
      error.value = errorMessage;
      console.error('Error fetching news article:', err);
    } finally {
      loading.value = false;
    }
  };

  // Watch slug changes and refetch
  watch(() => slug, fetchArticle, { immediate: true });

  return {
    article,
    loading,
    error,
    refetch: fetchArticle,
  };
}

export interface UseNewsSearchResult {
  articles: Ref<NewsArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  total: Ref<number>;
  page: Ref<number>;
  pageSize: Ref<number>;
  totalPages: Ref<number>;
  search: (query: string, params?: { page?: number; limit?: number; categoryId?: string }) => Promise<void>;
  clearResults: () => void;
}

export function useNewsSearch(): UseNewsSearchResult {
  const articles = ref<NewsArticle[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);
  const totalPages = ref(0);

  const search = async (query: string, params: { page?: number; limit?: number; categoryId?: string } = {}) => {
    if (!query.trim()) {
      clearResults();
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Use contract-generated client for search
      const response = await apiClient.getNews({
        search: query,
        page: params.page || 1,
        limit: params.limit || 10,
        categoryId: params.categoryId,
      });

      articles.value = response.data || [];
      total.value = response.pagination?.total_items || 0;
      page.value = response.pagination?.current_page || 1;
      pageSize.value = response.pagination?.items_per_page || 10;
      totalPages.value = response.pagination?.total_pages || 0;

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to search news';
      error.value = errorMessage;
      console.error('Error searching news:', err);
    } finally {
      loading.value = false;
    }
  };

  const clearResults = () => {
    articles.value = [];
    total.value = 0;
    page.value = 1;
    pageSize.value = 10;
    totalPages.value = 0;
    error.value = null;
  };

  return {
    articles,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    search,
    clearResults,
  };
}

export interface UseFeaturedNewsResult {
  articles: Ref<NewsArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useFeaturedNews(limit?: number): UseFeaturedNewsResult {
  const articles = ref<NewsArticle[]>([]);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchFeaturedNews = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Use contract-generated client for featured news
      const response = await apiClient.getFeaturedNews();
      articles.value = response.data?.slice(0, limit) || [];
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured news';
      error.value = errorMessage;
      console.error('Error fetching featured news:', err);
    } finally {
      loading.value = false;
    }
  };

  // Watch limit changes and refetch
  watch(() => limit, fetchFeaturedNews, { immediate: true });

  return {
    articles,
    loading,
    error,
    refetch: fetchFeaturedNews,
  };
}

export interface UseRecentNewsResult {
  articles: Ref<NewsArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useRecentNews(limit: number = 5): UseRecentNewsResult {
  const articles = ref<NewsArticle[]>([]);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchRecentNews = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Use contract-generated client - get latest news with limit
      const response = await apiClient.getNews({
        page: 1,
        limit: limit,
        // Could add sort parameter for recency if available in the contract
      });
      
      articles.value = response.data?.slice(0, limit) || [];
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch recent news';
      error.value = errorMessage;
      console.error('Error fetching recent news:', err);
    } finally {
      loading.value = false;
    }
  };

  // Watch limit changes and refetch
  watch(() => limit, fetchRecentNews, { immediate: true });

  return {
    articles,
    loading,
    error,
    refetch: fetchRecentNews,
  };
}
