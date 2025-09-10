// Research Composable - Vue composable for research data using contract-generated clients
// Provides clean interface for components to interact with research domain

import { ref, watch, type Ref } from 'vue';
import { apiClient } from '../lib/api-client';
import type { ResearchPublication } from '@international-center/public-api-client';

export interface UseResearchResult {
  articles: Ref<ResearchPublication[]>;
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

export interface UseResearchOptions {
  page?: number;
  limit?: number;
  search?: string;
  categoryId?: string;
  enabled?: boolean;
}

export function useResearch(options: UseResearchOptions = {}): UseResearchResult {
  const { enabled = true, ...params } = options;

  const articles = ref<ResearchPublication[]>([]);
  const loading = ref(enabled);
  const error = ref<string | null>(null);
  const total = ref(0);
  const page = ref(params.page || 1);
  const pageSize = ref(params.limit || 10);
  const totalPages = ref(0);
  const hasNext = ref(false);
  const hasPrevious = ref(false);

  const fetchResearch = async () => {
    if (!enabled) return;

    try {
      loading.value = true;
      error.value = null;

      // Use contract-generated client for type-safe API calls
      const response = await apiClient.getResearch({
        page: page.value,
        limit: pageSize.value,
        search: params.search,
        categoryId: params.categoryId,
      });

      // Extract data with type safety from generated response types
      articles.value = response.data || [];
      total.value = response.pagination?.total_items || 0;
      page.value = response.pagination?.current_page || 1;
      pageSize.value = response.pagination?.items_per_page || 10;
      totalPages.value = response.pagination?.total_pages || 0;
      hasNext.value = response.pagination?.has_next || false;
      hasPrevious.value = response.pagination?.has_previous || false;

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch research';
      error.value = errorMessage;
      console.error('Error fetching research:', err);
    } finally {
      loading.value = false;
    }
  };

  // Watch for parameter changes and refetch
  watch(() => [enabled, params.page, params.limit, params.search, params.categoryId], fetchResearch, { immediate: true });

  return {
    articles,
    loading,
    error,
    total,
    page,
    pageSize,
    totalPages,
    hasNext,
    hasPrevious,
    refetch: fetchResearch,
  };
}

export interface UseResearchArticleResult {
  article: ResearchArticle | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useResearchArticle(slug: string | null): UseResearchArticleResult {
  const [article, setArticle] = useState<ResearchArticle | null>(null);
  const [loading, setLoading] = useState(!!slug);
  const [error, setError] = useState<string | null>(null);

  const fetchArticle = async () => {
    if (!slug) {
      setArticle(null);
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await researchClient.getResearchArticleBySlug(slug);
      setArticle(response);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch research article';
      setError(errorMessage);
      console.error('Error fetching research article:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchArticle();
  }, [slug]);

  return {
    article,
    loading,
    error,
    refetch: fetchArticle,
  };
}

export interface UseFeaturedResearchResult {
  articles: ResearchArticle[];
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useFeaturedResearch(limit?: number): UseFeaturedResearchResult {
  const [articles, setArticles] = useState<ResearchArticle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFeaturedResearch = async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await researchClient.getFeaturedResearch(limit);
      setArticles(response);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured research';
      setError(errorMessage);
      console.error('Error fetching featured research:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFeaturedResearch();
  }, [limit]);

  return {
    articles,
    loading,
    error,
    refetch: fetchFeaturedResearch,
  };
}

export interface UseRecentResearchResult {
  articles: ResearchArticle[];
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useRecentResearch(limit: number = 5): UseRecentResearchResult {
  const [articles, setArticles] = useState<ResearchArticle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchRecentResearch = async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await researchClient.getRecentResearch(limit);
      setArticles(response);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch recent research';
      setError(errorMessage);
      console.error('Error fetching recent research:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRecentResearch();
  }, [limit]);

  return {
    articles,
    loading,
    error,
    refetch: fetchRecentResearch,
  };
}

