import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useNewsStore } from './news';
import type { NewsArticle, NewsCategory, GetNewsParams, SearchNewsParams } from '../lib/clients/news/types';

// Mock the news client with hoisted functions
const {
  mockGetNews,
  mockGetNewsArticleBySlug,
  mockGetFeaturedNews,
  mockSearchNews,
  mockGetNewsCategories
} = vi.hoisted(() => {
  const mockGetNewsFunc = vi.fn();
  const mockGetNewsArticleBySlugFunc = vi.fn();
  const mockGetFeaturedNewsFunc = vi.fn();
  const mockSearchNewsFunc = vi.fn();
  const mockGetNewsCategoriesFunc = vi.fn();
  
  return {
    mockGetNews: mockGetNewsFunc,
    mockGetNewsArticleBySlug: mockGetNewsArticleBySlugFunc,
    mockGetFeaturedNews: mockGetFeaturedNewsFunc,
    mockSearchNews: mockSearchNewsFunc,
    mockGetNewsCategories: mockGetNewsCategoriesFunc,
  };
});

vi.mock('../lib/clients', () => ({
  newsClient: {
    getNews: mockGetNews,
    getNewsArticleBySlug: mockGetNewsArticleBySlug,
    getFeaturedNews: mockGetFeaturedNews,
    searchNews: mockSearchNews,
    getNewsCategories: mockGetNewsCategories,
  }
}));

describe('NewsStore', () => {
  let pinia: ReturnType<typeof createPinia>;

  beforeEach(() => {
    // Clear all mock function calls and reset mock implementations
    mockGetNews.mockClear();
    mockGetNewsArticleBySlug.mockClear();
    mockGetFeaturedNews.mockClear();
    mockSearchNews.mockClear();
    mockGetNewsCategories.mockClear();
    vi.clearAllMocks();
    
    // Create fresh pinia instance for each test
    pinia = createPinia();
    setActivePinia(pinia);
  });

  afterEach(() => {
    // Manually reset the store to initial state
    try {
      const store = useNewsStore();
      store.$reset();
    } catch (e) {
      // Ignore if store doesn't exist
    }
  });

  describe('Initial State', () => {
    it('should initialize with empty state and default values', () => {
      const store = useNewsStore();
      
      expect(store.news).toEqual([]);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(store.total).toBe(0);
      expect(store.page).toBe(1);
      expect(store.pageSize).toBe(10);
      expect(store.categories).toEqual([]);
      expect(store.featuredNews).toEqual([]);
      expect(store.searchResults).toEqual([]);
    });

    it('should provide totalPages getter based on total and pageSize', () => {
      const store = useNewsStore();
      
      // Initially should be 0
      expect(store.totalPages).toBe(0);
      
      // Set some data to test calculation
      store.$patch({
        total: 25,
        pageSize: 10
      });
      
      expect(store.totalPages).toBe(3); // Math.ceil(25/10) = 3
    });
  });

  describe('State Management', () => {
    it('should manage loading state during operations', () => {
      const store = useNewsStore();
      
      expect(store.loading).toBe(false);
      
      // Should be able to set loading state
      store.setLoading(true);
      expect(store.loading).toBe(true);
      
      store.setLoading(false);
      expect(store.loading).toBe(false);
    });

    it('should manage error state with proper clearing', () => {
      const store = useNewsStore();
      
      expect(store.error).toBeNull();
      
      // Should set error
      store.setError('Network error occurred');
      expect(store.error).toBe('Network error occurred');
      
      // Should clear error
      store.clearError();
      expect(store.error).toBeNull();
    });

    it('should update news data and pagination state', () => {
      const store = useNewsStore();
      const mockNews: NewsArticle[] = [
        {
          news_id: 'news-1',
          title: 'Test News 1',
          summary: 'Summary 1',
          slug: 'test-news-1',
          category_id: 'cat-1',
          publishing_status: 'published',
          publication_timestamp: '2024-01-01T00:00:00Z',
          tags: ['test'],
          news_type: 'announcement',
          priority_level: 'normal',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.setNews(mockNews, 25, 2, 10);
      
      expect(store.news).toEqual(mockNews);
      expect(store.total).toBe(25);
      expect(store.page).toBe(2);
      expect(store.pageSize).toBe(10);
    });
  });

  describe('Actions - News Operations', () => {
    it('should fetch news with parameters and update state accordingly', async () => {
      const store = useNewsStore();
      
      // Verify initial state
      expect(store.news).toEqual([]);
      expect(store.total).toBe(0);
      
      const mockResponse = {
        news: [
          {
            news_id: 'news-1',
            title: 'Test News',
            summary: 'Test Summary',
            slug: 'test-news',
            category_id: 'cat-1',
            publishing_status: 'published' as const,
            publication_timestamp: '2024-01-01T00:00:00Z',
            tags: ['test'],
            news_type: 'announcement' as const,
            priority_level: 'normal' as const,
            created_on: '2024-01-01T00:00:00Z',
            is_deleted: false
          }
        ],
        count: 1,
        correlation_id: 'test-correlation-id'
      };
      
      mockGetNews.mockResolvedValueOnce(mockResponse);
      
      const params: GetNewsParams = { page: 1, pageSize: 10 };
      
      await store.fetchNews(params);
      
      expect(mockGetNews).toHaveBeenCalledWith(params);
      expect(store.news).toEqual(mockResponse.news);
      expect(store.total).toBe(mockResponse.count);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    it('should handle news fetch errors properly', async () => {
      const store = useNewsStore();
      
      const errorMessage = 'Failed to fetch news';
      mockGetNews.mockRejectedValueOnce(new Error(errorMessage));
      
      await store.fetchNews();
      
      expect(store.error).toBe(errorMessage);
      expect(store.loading).toBe(false);
      expect(store.news).toEqual([]);
    });

    it('should fetch single news article by slug', async () => {
      const store = useNewsStore();
      
      const mockArticle: NewsArticle = {
        news_id: 'news-1',
        title: 'Single Article',
        summary: 'Article Summary',
        slug: 'single-article',
        category_id: 'cat-1',
        publishing_status: 'published',
        publication_timestamp: '2024-01-01T00:00:00Z',
        tags: ['article'],
        news_type: 'feature',
        priority_level: 'high',
        created_on: '2024-01-01T00:00:00Z',
        is_deleted: false,
        content: 'Full article content'
      };
      
      const mockResponse = {
        news: mockArticle,
        correlation_id: 'article-correlation-id'
      };
      
      mockGetNewsArticleBySlug.mockResolvedValueOnce(mockResponse);
      
      const result = await store.fetchNewsArticle('single-article');
      
      expect(mockGetNewsArticleBySlug).toHaveBeenCalledWith('single-article');
      expect(result).toEqual(mockArticle);
    });
  });

  describe('Actions - Featured News', () => {
    it('should fetch featured news and cache in store', async () => {
      const store = useNewsStore();
      
      const mockFeaturedNews = [
        {
          news_id: 'featured-1',
          title: 'Featured News',
          summary: 'Featured Summary',
          slug: 'featured-news',
          category_id: 'cat-1',
          publishing_status: 'published' as const,
          publication_timestamp: '2024-01-01T00:00:00Z',
          tags: ['featured'],
          news_type: 'feature' as const,
          priority_level: 'urgent' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      const mockResponse = {
        news: mockFeaturedNews,
        count: 1,
        correlation_id: 'featured-correlation-id'
      };
      
      mockGetFeaturedNews.mockResolvedValueOnce(mockResponse);
      
      await store.fetchFeaturedNews(5);
      
      expect(mockGetFeaturedNews).toHaveBeenCalledWith(5);
      expect(store.featuredNews).toEqual(mockFeaturedNews);
    });
  });

  describe('Actions - Search Operations', () => {
    it('should perform news search and store results separately', async () => {
      const store = useNewsStore();
      
      const mockSearchResults = [
        {
          news_id: 'search-1',
          title: 'Search Result',
          summary: 'Search Summary',
          slug: 'search-result',
          category_id: 'cat-1',
          publishing_status: 'published' as const,
          publication_timestamp: '2024-01-01T00:00:00Z',
          tags: ['search'],
          news_type: 'update' as const,
          priority_level: 'normal' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      const mockResponse = {
        news: mockSearchResults,
        count: 1,
        correlation_id: 'search-correlation-id'
      };
      
      mockSearchNews.mockResolvedValueOnce(mockResponse);
      
      const searchParams: SearchNewsParams = {
        q: 'test query',
        page: 1,
        pageSize: 10
      };
      
      await store.searchNews(searchParams);
      
      expect(mockSearchNews).toHaveBeenCalledWith(searchParams);
      expect(store.searchResults).toEqual(mockSearchResults);
      expect(store.searchTotal).toBe(1);
    });

    it('should clear search results when query is empty', async () => {
      const store = useNewsStore();
      
      // Set some initial search results
      store.$patch({
        searchResults: [{ 
          news_id: 'test',
          title: 'Test',
          summary: 'Test',
          slug: 'test',
          category_id: 'cat-1',
          publishing_status: 'published' as const,
          publication_timestamp: '2024-01-01T00:00:00Z',
          tags: [],
          news_type: 'announcement' as const,
          priority_level: 'normal' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }],
        searchTotal: 1
      });
      
      await store.searchNews({ q: '', page: 1, pageSize: 10 });
      
      expect(store.searchResults).toEqual([]);
      expect(store.searchTotal).toBe(0);
    });
  });

  describe('Actions - Categories', () => {
    it('should fetch and cache news categories', async () => {
      const store = useNewsStore();
      
      const mockCategories: NewsCategory[] = [
        {
          category_id: 'cat-1',
          name: 'Company News',
          slug: 'company-news',
          order_number: 1
        }
      ];
      
      const mockResponse = {
        categories: mockCategories,
        count: 1,
        correlation_id: 'categories-correlation-id'
      };
      
      mockGetNewsCategories.mockResolvedValueOnce(mockResponse);
      
      await store.fetchNewsCategories();
      
      expect(mockGetNewsCategories).toHaveBeenCalled();
      expect(store.categories).toEqual(mockCategories);
    });
  });

  describe('Getters', () => {
    it('should provide computed values for pagination', () => {
      const store = useNewsStore();
      
      store.$patch({
        total: 0,
        pageSize: 10
      });
      expect(store.totalPages).toBe(0);
      
      store.$patch({
        total: 15,
        pageSize: 5
      });
      expect(store.totalPages).toBe(3);
    });

    it('should provide hasNews getter for conditional rendering', () => {
      const store = useNewsStore();
      
      expect(store.hasNews).toBe(false);
      
      store.$patch({
        news: [{
          news_id: 'test',
          title: 'Test',
          summary: 'Test',
          slug: 'test',
          category_id: 'cat-1',
          publishing_status: 'published',
          publication_timestamp: '2024-01-01T00:00:00Z',
          tags: [],
          news_type: 'announcement',
          priority_level: 'normal',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }]
      });
      
      expect(store.hasNews).toBe(true);
    });
  });

  describe('Cache Management', () => {
    it('should cache news data and avoid duplicate fetches', async () => {
      const store = useNewsStore();
      
      const mockResponse = {
        news: [{
          news_id: 'cached-1',
          title: 'Cached News',
          summary: 'Cached Summary',
          slug: 'cached-news',
          category_id: 'cat-1',
          publishing_status: 'published' as const,
          publication_timestamp: '2024-01-01T00:00:00Z',
          tags: [],
          news_type: 'announcement' as const,
          priority_level: 'normal' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }],
        count: 1,
        correlation_id: 'cache-test-id'
      };
      
      mockGetNews.mockResolvedValueOnce(mockResponse);
      
      // First fetch should call API
      await store.fetchNews({ page: 1, pageSize: 10 });
      expect(mockGetNews).toHaveBeenCalledTimes(1);
      
      // Second fetch with same params should use cache
      await store.fetchNews({ page: 1, pageSize: 10 }, { useCache: true });
      expect(mockGetNews).toHaveBeenCalledTimes(1); // Still only called once
    });

    it('should invalidate cache and refetch when requested', async () => {
      const store = useNewsStore();
      
      const mockResponse = {
        news: [],
        count: 0,
        correlation_id: 'invalidate-test-id'
      };
      
      mockGetNews.mockResolvedValue(mockResponse);
      
      // First fetch
      await store.fetchNews({ page: 1, pageSize: 10 });
      expect(mockGetNews).toHaveBeenCalledTimes(1);
      
      // Invalidate cache and fetch again
      store.invalidateCache();
      await store.fetchNews({ page: 1, pageSize: 10 });
      expect(mockGetNews).toHaveBeenCalledTimes(2);
    });
  });
});