// Research Composables - Vue 3 Composition API composables for research data
// Provides clean interface for Vue components to interact with research domain

import { ref, computed, onMounted, watch, type Ref } from 'vue';
import { researchClient } from '../lib/clients';
import type { ResearchArticle, GetResearchParams } from '../lib/clients/research/types';

export interface UseResearchArticlesResult {
  articles: Ref<ResearchArticle[]>;
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

export function useResearchArticles(options: UseResearchArticlesOptions = {}): UseResearchArticlesResult {
  const { enabled = true, immediate = true, ...params } = options;

  const articles = ref<ResearchArticle[]>([]);
  const loading = ref(enabled && immediate);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(params.page || 1);
  const pageSize = ref(params.pageSize || 10);
  const totalPages = ref(0);

  const fetchArticles = async () => {
    if (!enabled) return;

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { data: [...], pagination: {...}, success: boolean }
      const response = await researchClient.getResearchArticles(params);

      if (response.data) {
        articles.value = response.data;
        total.value = response.pagination?.total || response.data.length;
        // Calculate pagination from response
        page.value = response.pagination?.page || params.page || 1;
        pageSize.value = response.pagination?.pageSize || params.pageSize || 10;
        totalPages.value = response.pagination?.totalPages || Math.ceil(total.value / pageSize.value);
      } else {
        throw new Error('Failed to fetch research articles');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch research articles';
      error.value = errorMessage;
      console.error('Error fetching research articles:', err);
      articles.value = [];
      total.value = 0;
      totalPages.value = 0;
    } finally {
      loading.value = false;
    }
  };

  // Watch for parameter changes
  watch(() => params, fetchArticles, { deep: true });
  
  if (immediate) {
    onMounted(fetchArticles);
  }

  return {
    articles,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    refetch: fetchArticles,
  };
}

export interface UseResearchArticleResult {
  article: Ref<ResearchArticle | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useResearchArticle(slug: Ref<string | null> | string | null): UseResearchArticleResult {
  const slugRef = typeof slug === 'string' ? ref(slug) : slug || ref(null);
  
  const article = ref<ResearchArticle | null>(null);
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

      // Backend returns: { data: {...}, success: boolean }
      const response = await researchClient.getResearchArticleBySlug(slugRef.value);
      
      if (response.data) {
        article.value = response.data;
      } else {
        throw new Error('Failed to fetch research article');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch research article';
      error.value = errorMessage;
      console.error('Error fetching research article:', err);
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

export interface UseFeaturedResearchResult {
  articles: Ref<ResearchArticle[]>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useFeaturedResearch(limit?: Ref<number> | number): UseFeaturedResearchResult {
  const limitRef = typeof limit === 'number' ? ref(limit) : limit;
  
  const articles = ref<ResearchArticle[]>([]);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchFeaturedArticles = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { data: [...], pagination: {...}, success: boolean }
      const response = await researchClient.getFeaturedResearch(limitRef?.value);
      
      if (response.data) {
        articles.value = response.data;
      } else {
        throw new Error('Failed to fetch featured research articles');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured research articles';
      error.value = errorMessage;
      console.error('Error fetching featured research articles:', err);
      articles.value = [];
    } finally {
      loading.value = false;
    }
  };

  // Watch for limit changes
  if (limitRef) {
    watch(limitRef, fetchFeaturedArticles, { immediate: true });
  } else {
    onMounted(fetchFeaturedArticles);
  }

  return {
    articles,
    loading,
    error,
    refetch: fetchFeaturedArticles,
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
  search: (query: string, options?: Partial<GetResearchParams>) => Promise<void>;
}

export function useSearchResearch(): UseSearchResearchResult {
  const results = ref<ResearchArticle[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);
  const totalPages = ref(0);

  const search = async (query: string, options: Partial<GetResearchParams> = {}) => {
    if (!query.trim()) {
      results.value = [];
      total.value = 0;
      totalPages.value = 0;
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { data: [...], pagination: {...}, success: boolean }
      const response = await researchClient.searchResearch({
        q: query,
        page: options.page || 1,
        pageSize: options.pageSize || 10,
        category: options.category,
        ...options,
      });

      if (response.data) {
        results.value = response.data;
        total.value = response.pagination?.total || response.data.length;
        page.value = response.pagination?.page || options.page || 1;
        pageSize.value = response.pagination?.pageSize || options.pageSize || 10;
        totalPages.value = response.pagination?.totalPages || Math.ceil(total.value / pageSize.value);
      } else {
        throw new Error('Failed to search research articles');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to search research articles';
      error.value = errorMessage;
      console.error('Error searching research articles:', err);
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