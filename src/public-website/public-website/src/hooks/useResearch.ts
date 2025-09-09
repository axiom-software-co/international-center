// Research Hook - React hook for research data
// Provides clean interface for components to interact with research domain

import { useState, useEffect } from 'react';
import { researchClient } from '../lib/clients';
import type { ResearchArticle, GetResearchParams } from '../lib/clients';

export interface UseResearchResult {
  articles: ResearchArticle[];
  loading: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  refetch: () => Promise<void>;
}

export interface UseResearchOptions extends GetResearchParams {
  enabled?: boolean;
}

export function useResearch(options: UseResearchOptions = {}): UseResearchResult {
  const { enabled = true, ...params } = options;

  const [articles, setArticles] = useState<ResearchArticle[]>([]);
  const [loading, setLoading] = useState(enabled);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    total: 0,
    page: 1,
    pageSize: 10,
    totalPages: 0,
  });

  const fetchResearch = async () => {
    if (!enabled) return;

    try {
      setLoading(true);
      setError(null);

      const response = await researchClient.getResearchArticles(params);

      setArticles(response.articles || []);
      setPagination({
        total: response.pagination?.total || 0,
        page: response.pagination?.page || 1,
        pageSize: response.pagination?.pageSize || 10,
        totalPages: response.pagination?.totalPages || 0,
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch research';
      setError(errorMessage);
      console.error('Error fetching research:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchResearch();
  }, [enabled, JSON.stringify(params)]);

  return {
    articles,
    loading,
    error,
    ...pagination,
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

