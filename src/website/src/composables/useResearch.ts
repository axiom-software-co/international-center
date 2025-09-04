// Research Composables - Database Schema Compliant
// Updated to work with TABLES-RESEARCH.md aligned types

import { ref, computed, onMounted, watch, type Ref } from 'vue';
import { researchClient } from '../lib/clients';
import type { ResearchArticle, GetResearchParams, FeaturedResearch } from '../lib/clients/research/types';

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

      // Backend returns: { data: [...], pagination: {...}, success: boolean, message?, errors? }
      const response = await researchClient.getResearchArticles(params);

      if (response.success && response.data) {
        articles.value = response.data;
        total.value = response.pagination.total;
        page.value = response.pagination.page;
        pageSize.value = response.pagination.pageSize;
        totalPages.value = response.pagination.totalPages;
      } else {
        throw new Error(response.message || 'Failed to fetch research articles');
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

      // Backend returns: { data: {...}, success: boolean, message?, errors? }
      const response = await researchClient.getResearchArticleBySlug(slugRef.value);
      
      if (response.success && response.data) {
        article.value = response.data;
      } else {
        throw new Error(response.message || 'Failed to fetch research article');
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
  featuredResearch: Ref<FeaturedResearch | null>;
  article: Ref<ResearchArticle | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  refetch: () => Promise<void>;
}

export function useFeaturedResearch(): UseFeaturedResearchResult {
  const featuredResearch = ref<FeaturedResearch | null>(null);
  const article = ref<ResearchArticle | null>(null);
  const loading = ref(true);
  const error = ref<string | null>(null);

  const fetchFeaturedResearch = async () => {
    try {
      loading.value = true;
      error.value = null;

      // Backend returns: { data: {...}, success: boolean, message?, errors? }
      const response = await researchClient.getFeaturedResearch();
      
      if (response.success && response.data) {
        featuredResearch.value = response.data;
        // Optionally fetch the full research article details if needed
        if (response.data.research_id) {
          const articleResponse = await researchClient.getResearchArticleById(response.data.research_id);
          if (articleResponse.success && articleResponse.data) {
            article.value = articleResponse.data;
          }
        }
      } else {
        // No featured research is not an error, just no data
        featuredResearch.value = null;
        article.value = null;
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured research';
      error.value = errorMessage;
      console.error('Error fetching featured research:', err);
      featuredResearch.value = null;
      article.value = null;
    } finally {
      loading.value = false;
    }
  };

  onMounted(fetchFeaturedResearch);

  return {
    featuredResearch,
    article,
    loading,
    error,
    refetch: fetchFeaturedResearch,
  };
}

// Legacy compatibility function
export function useFeaturedResearchArticles(): { articles: Ref<ResearchArticle[]>; loading: Ref<boolean>; error: Ref<string | null>; refetch: () => Promise<void> } {
  const { article, loading, error, refetch } = useFeaturedResearch();
  const articles = computed(() => article.value ? [article.value] : []);
  
  return {
    articles,
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

      // Backend returns: { data: [...], pagination: {...}, success: boolean, message?, errors? }
      const response = await researchClient.searchResearch({
        q: query,
        page: options.page || 1,
        pageSize: options.pageSize || 10,
        category_id: options.category_id,
        research_type: options.research_type,
        publishing_status: options.publishing_status,
        publication_date_from: options.publication_date_from,
        publication_date_to: options.publication_date_to,
        ...options,
      });

      if (response.success && response.data) {
        results.value = response.data;
        total.value = response.pagination.total;
        page.value = response.pagination.page;
        pageSize.value = response.pagination.pageSize;
        totalPages.value = response.pagination.totalPages;
      } else {
        throw new Error(response.message || 'Failed to search research articles');
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