// News Hook - React hook for news data
// Provides clean interface for components to interact with news domain

import { useState, useEffect } from 'react';
import { newsClient } from '../lib/clients';
import type { NewsArticle, GetNewsParams, SearchNewsParams } from '../lib/clients';

export interface UseNewsResult {
  articles: NewsArticle[];
  loading: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  refetch: () => Promise<void>;
}

export interface UseNewsOptions extends GetNewsParams {
  enabled?: boolean;
}

export function useNews(options: UseNewsOptions = {}): UseNewsResult {
  const { enabled = true, ...params } = options;

  const [articles, setArticles] = useState<NewsArticle[]>([]);
  const [loading, setLoading] = useState(enabled);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    total: 0,
    page: 1,
    pageSize: 10,
    totalPages: 0,
  });

  const fetchNews = async () => {
    if (!enabled) return;

    try {
      setLoading(true);
      setError(null);

      const response = await newsClient.getNewsArticles(params);

      setArticles(response.data);
      setPagination({
        total: response.total,
        page: response.page,
        pageSize: response.pageSize,
        totalPages: response.totalPages,
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch news';
      setError(errorMessage);
      console.error('Error fetching news:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNews();
  }, [enabled, JSON.stringify(params)]);

  return {
    articles,
    loading,
    error,
    ...pagination,
    refetch: fetchNews,
  };
}

export interface UseNewsArticleResult {
  article: NewsArticle | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useNewsArticle(slug: string | null): UseNewsArticleResult {
  const [article, setArticle] = useState<NewsArticle | null>(null);
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

      const response = await newsClient.getNewsArticleBySlug(slug);
      setArticle(response);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch news article';
      setError(errorMessage);
      console.error('Error fetching news article:', err);
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

export interface UseNewsSearchResult {
  articles: NewsArticle[];
  loading: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  search: (query: string, params?: Omit<SearchNewsParams, 'q'>) => Promise<void>;
  clearResults: () => void;
}

export function useNewsSearch(): UseNewsSearchResult {
  const [articles, setArticles] = useState<NewsArticle[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    total: 0,
    page: 1,
    pageSize: 10,
    totalPages: 0,
  });

  const search = async (query: string, params: Omit<SearchNewsParams, 'q'> = {}) => {
    if (!query.trim()) {
      clearResults();
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await newsClient.searchNewsArticles({ q: query, ...params });

      setArticles(response.data);
      setPagination({
        total: response.total,
        page: response.page,
        pageSize: response.pageSize,
        totalPages: response.totalPages,
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to search news';
      setError(errorMessage);
      console.error('Error searching news:', err);
    } finally {
      setLoading(false);
    }
  };

  const clearResults = () => {
    setArticles([]);
    setPagination({ total: 0, page: 1, pageSize: 10, totalPages: 0 });
    setError(null);
  };

  return {
    articles,
    loading,
    error,
    ...pagination,
    search,
    clearResults,
  };
}

export interface UseFeaturedNewsResult {
  articles: NewsArticle[];
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useFeaturedNews(limit?: number): UseFeaturedNewsResult {
  const [articles, setArticles] = useState<NewsArticle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFeaturedNews = async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await newsClient.getFeaturedNews(limit);
      setArticles(response);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch featured news';
      setError(errorMessage);
      console.error('Error fetching featured news:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFeaturedNews();
  }, [limit]);

  return {
    articles,
    loading,
    error,
    refetch: fetchFeaturedNews,
  };
}

export interface UseRecentNewsResult {
  articles: NewsArticle[];
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useRecentNews(limit: number = 5): UseRecentNewsResult {
  const [articles, setArticles] = useState<NewsArticle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchRecentNews = async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await newsClient.getRecentNews(limit);
      setArticles(response);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch recent news';
      setError(errorMessage);
      console.error('Error fetching recent news:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRecentNews();
  }, [limit]);

  return {
    articles,
    loading,
    error,
    refetch: fetchRecentNews,
  };
}
