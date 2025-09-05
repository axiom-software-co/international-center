// News Composables Tests - State management and API integration validation
// Tests validate useNews composables with database schema-compliant reactive state

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ref, nextTick, defineComponent } from 'vue';
import { mount } from '@vue/test-utils';
import { useNews, useNewsArticle, useFeaturedNews, useSearchNews, useNewsCategories } from './useNews';
import type { NewsArticle, NewsResponse, NewsArticleResponse, GetNewsParams, SearchNewsParams, NewsCategory } from '../lib/clients/news/types';
import { useNewsStore } from '../stores/news';

// Mock the newsClient singleton with hoisted functions
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
  },
  // Pass through any types that might be imported
  ...vi.importActual('../lib/clients')
}));

// Mock the stores/news module so the store uses our mocked client
vi.mock('../stores/news', () => {
  return {
    useNewsStore: vi.fn()
  };
});

// Database schema-compliant mock news article for testing
const createMockDatabaseNews = (overrides: Partial<any> = {}): any => ({
  news_id: 'news-uuid-123',
  title: 'Mock Database News Article',
  summary: 'News article summary from database schema',
  content: 'Full news article content with journalism standards',
  slug: 'mock-database-news-article',
  category_id: 'news-category-uuid-456',
  image_url: 'https://example.com/news-image.jpg',
  author_name: 'Database Reporter',
  publication_timestamp: '2024-03-15T14:30:00Z',
  external_source: 'Database News Source',
  external_url: 'https://external.example.com/news',
  publishing_status: 'published' as const,
  tags: ['database', 'schema', 'news'],
  news_type: 'announcement' as const,
  priority_level: 'normal' as const,
  created_on: '2024-01-01T00:00:00Z',
  created_by: 'reporter@example.com',
  modified_on: '2024-01-02T00:00:00Z',
  modified_by: 'editor@example.com',
  is_deleted: false,
  deleted_on: null,
  deleted_by: null,
  ...overrides,
});

// Database schema-compliant mock news category for testing
const createMockDatabaseNewsCategory = (overrides: Partial<any> = {}): any => ({
  category_id: 'news-category-uuid-456',
  name: 'Database News Category',
  slug: 'database-news-category',
  description: 'News category for database-related articles',
  is_default_unassigned: false,
  created_on: '2024-01-01T00:00:00Z',
  created_by: 'admin@example.com',
  modified_on: '2024-01-02T00:00:00Z',
  modified_by: 'admin@example.com',
  is_deleted: false,
  deleted_on: null,
  deleted_by: null,
  ...overrides,
});

describe('useNews Composables', () => {
  beforeEach(() => {
    mockGetNews.mockClear();
    mockGetNewsArticleBySlug.mockClear();
    mockGetFeaturedNews.mockClear();
    mockSearchNews.mockClear();
    mockGetNewsCategories.mockClear();
    
    // Import Vue reactivity functions
    const { computed } = require('vue');
    
    // Set up the mock store with reactive state and actions that call our mocked client
    const mockStore = {
      // Reactive state
      news: ref([]),
      categories: ref([]),
      featuredNews: ref([]),
      searchResults: ref([]),
      loading: ref(false),
      error: ref(null),
      total: ref(0),
      page: ref(1),
      pageSize: ref(10),
      searchTotal: ref(0),
      
      // Actions that call the mocked client and update state
      async fetchNews(params?: any, options?: any) {
        mockStore.loading.value = true;
        mockStore.error.value = null;
        try {
          const result = await mockGetNews(params);
          mockStore.news.value = result.news;
          mockStore.total.value = result.count;
          mockStore.page.value = params?.page || 1;
          mockStore.pageSize.value = params?.pageSize || 10;
        } catch (err) {
          mockStore.error.value = err instanceof Error ? err.message : 'Unknown error';
          mockStore.news.value = [];
          mockStore.total.value = 0;
        } finally {
          mockStore.loading.value = false;
        }
      },
      
      async fetchNewsArticle(slug: string) {
        mockStore.loading.value = true;
        mockStore.error.value = null;
        try {
          const result = await mockGetNewsArticleBySlug(slug);
          return result?.news || null;
        } catch (err) {
          mockStore.error.value = err instanceof Error ? err.message : 'Unknown error';
          return null;
        } finally {
          mockStore.loading.value = false;
        }
      },
      
      async fetchFeaturedNews(limit?: number) {
        mockStore.loading.value = true;
        mockStore.error.value = null;
        try {
          const result = await mockGetFeaturedNews(limit);
          mockStore.featuredNews.value = result.news;
        } catch (err) {
          mockStore.error.value = err instanceof Error ? err.message : 'Unknown error';
          mockStore.featuredNews.value = [];
        } finally {
          mockStore.loading.value = false;
        }
      },
      
      async searchNews(params: any) {
        mockStore.loading.value = true;
        mockStore.error.value = null;
        try {
          const result = await mockSearchNews(params);
          mockStore.searchResults.value = result.news;
          mockStore.searchTotal.value = result.count;
        } catch (err) {
          mockStore.error.value = err instanceof Error ? err.message : 'Unknown error';
          mockStore.searchResults.value = [];
          mockStore.searchTotal.value = 0;
        } finally {
          mockStore.loading.value = false;
        }
      },
      
      async fetchNewsCategories() {
        mockStore.loading.value = true;
        mockStore.error.value = null;
        try {
          const result = await mockGetNewsCategories();
          mockStore.categories.value = result.categories;
        } catch (err) {
          mockStore.error.value = err instanceof Error ? err.message : 'Unknown error';
          mockStore.categories.value = [];
        } finally {
          mockStore.loading.value = false;
        }
      }
    };
    
    // Add computed property after mockStore is defined
    mockStore.totalPages = computed(() => Math.ceil(mockStore.total.value / mockStore.pageSize.value) || 0);
    
    // Set up the useNewsStore mock to return our mock store
    vi.mocked(useNewsStore).mockReturnValue(mockStore);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('useNews', () => {
    it('should initialize with proper default state', () => {
      mockGetNews.mockResolvedValue({
        news: [],
        count: 0,
        correlation_id: 'test-correlation-id'
      });

      const { news, loading, error, total, page, pageSize, totalPages } = useNews({
        immediate: false
      });

      expect(news.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      expect(totalPages.value).toBe(0);
    }, 5000);

    it('should fetch news with database schema-compliant data', async () => {
      const mockDatabaseNews = [
        createMockDatabaseNews(),
        createMockDatabaseNews({
          news_id: 'news-uuid-124',
          title: 'Second Database News',
          slug: 'second-database-news',
          news_type: 'press_release' as const,
          priority_level: 'high' as const,
          author_name: 'Second Reporter',
        })
      ];

      const mockResponse: NewsResponse = {
        news: mockDatabaseNews,
        count: 2,
        correlation_id: 'news-correlation-id'
      };

      mockGetNews.mockResolvedValue(mockResponse);

      const { news, loading, error, total, totalPages, refetch } = useNews({
        immediate: false
      });

      expect(loading.value).toBe(false);

      await refetch();

      await nextTick();

      expect(mockGetNews).toHaveBeenCalledTimes(1);
      expect(news.value).toHaveLength(2);
      expect(total.value).toBe(2);
      expect(totalPages.value).toBe(1);
      expect(error.value).toBe(null);
      expect(loading.value).toBe(false);

      // Validate database schema fields are present
      const firstNews = news.value[0];
      expect(firstNews.news_id).toBeDefined();
      expect(firstNews.summary).toBeDefined(); // Not 'excerpt'
      expect(firstNews.author_name).toBeDefined(); // Not 'author'
      expect(firstNews.publishing_status).toBeDefined();
      expect(firstNews.news_type).toBeDefined();
      expect(firstNews.priority_level).toBeDefined();
      expect(firstNews.publication_timestamp).toBeDefined(); // Not 'published_at'
      expect(firstNews.external_source).toBeDefined();
      expect(firstNews.external_url).toBeDefined();
      expect(firstNews.tags).toBeDefined(); // Array field
      expect(firstNews.is_deleted).toBeDefined();
      expect(firstNews.created_on).toBeDefined();
    }, 5000);

    it('should handle API errors gracefully', async () => {
      mockGetNews.mockRejectedValue(new Error('API Error'));

      const { news, loading, error, refetch } = useNews({
        immediate: false
      });

      await refetch();
      await nextTick();

      expect(error.value).toBe('API Error');
      expect(news.value).toEqual([]);
      expect(loading.value).toBe(false);
    }, 5000);

    it('should handle query parameters correctly', async () => {
      mockGetNews.mockResolvedValue({
        news: [],
        count: 0,
        correlation_id: 'params-correlation-id'
      });

      const params: GetNewsParams = {
        page: 2,
        pageSize: 20,
        category: 'announcements',
        featured: true,
        sortBy: 'date-desc'
      };

      const { refetch } = useNews({
        immediate: false,
        ...params
      });

      await refetch();

      expect(mockGetNews).toHaveBeenCalledWith(
        expect.objectContaining(params)
      );
    }, 5000);

    it('should handle pagination calculations correctly', async () => {
      mockGetNews.mockResolvedValue({
        news: Array(15).fill(null).map((_, i) => createMockDatabaseNews({
          news_id: `news-${i}`,
          title: `News Article ${i}`,
          slug: `news-article-${i}`
        })),
        count: 150,
        correlation_id: 'pagination-correlation-id'
      });

      const { total, pageSize, totalPages, refetch } = useNews({
        immediate: false,
        pageSize: 15
      });

      await refetch();
      await nextTick();

      expect(total.value).toBe(150);
      expect(pageSize.value).toBe(15);
      expect(totalPages.value).toBe(10); // 150 / 15 = 10
    }, 5000);
  });

  describe('useNewsArticle', () => {
    it('should fetch single news article by slug', async () => {
      const mockNews = createMockDatabaseNews({
        slug: 'single-news-test'
      });

      const mockResponse: NewsArticleResponse = {
        news: mockNews,
        correlation_id: 'single-news-correlation-id'
      };

      mockGetNewsArticleBySlug.mockResolvedValue(mockResponse);

      const slugRef = ref('single-news-test');
      const { news, loading, error } = useNewsArticle(slugRef);

      await nextTick();

      expect(mockGetNewsArticleBySlug).toHaveBeenCalledWith('single-news-test');
      expect(news.value).toEqual(mockNews);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);

      // Validate database schema compliance
      expect(news.value?.news_id).toBeDefined();
      expect(news.value?.summary).toBeDefined(); // Not 'excerpt'
      expect(news.value?.author_name).toBeDefined(); // Not 'author'
      expect(news.value?.publication_timestamp).toBeDefined(); // Not 'published_at'
      expect(news.value?.news_type).toBeDefined();
    }, 5000);

    it('should handle slug changes reactively', async () => {
      mockGetNewsArticleBySlug.mockResolvedValue({
        news: createMockDatabaseNews(),
        correlation_id: 'reactive-correlation-id'
      });

      const slugRef = ref('initial-slug');
      const { refetch } = useNewsArticle(slugRef);

      await nextTick();

      expect(mockGetNewsArticleBySlug).toHaveBeenCalledWith('initial-slug');

      // Change slug
      slugRef.value = 'updated-slug';
      await nextTick();

      expect(mockGetNewsArticleBySlug).toHaveBeenCalledWith('updated-slug');
      expect(mockGetNewsArticleBySlug).toHaveBeenCalledTimes(2);
    }, 5000);

    it('should handle empty slug gracefully', async () => {
      const { news, loading } = useNewsArticle(ref(null));

      await nextTick();

      expect(mockGetNewsArticleBySlug).not.toHaveBeenCalled();
      expect(news.value).toBe(null);
      expect(loading.value).toBe(false);
    }, 5000);

    it('should handle API errors', async () => {
      mockGetNewsArticleBySlug.mockRejectedValue(new Error('News article not found'));

      const { news, error, loading } = useNewsArticle('non-existent-slug');

      await nextTick();

      expect(error.value).toBe('News article not found');
      expect(news.value).toBe(null);
      expect(loading.value).toBe(false);
    }, 5000);
  });

  describe('useFeaturedNews', () => {
    it('should fetch featured news articles', async () => {
      const mockFeaturedNews = [
        createMockDatabaseNews({ title: 'Featured News 1' }),
        createMockDatabaseNews({ 
          news_id: 'news-uuid-125',
          title: 'Featured News 2',
          slug: 'featured-news-2',
          news_type: 'feature' as const,
          priority_level: 'high' as const,
          author_name: 'Featured Reporter'
        })
      ];

      const mockResponse: NewsResponse = {
        news: mockFeaturedNews,
        count: 2,
        correlation_id: 'featured-correlation-id'
      };

      mockGetFeaturedNews.mockResolvedValue(mockResponse);

      const { news, loading, error } = useFeaturedNews();

      await nextTick();

      expect(mockGetFeaturedNews).toHaveBeenCalledWith(undefined);
      expect(news.value).toHaveLength(2);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);

      // Validate database schema compliance
      expect(news.value[0].news_id).toBeDefined();
      expect(news.value[0].publishing_status).toBeDefined();
      expect(news.value[0].news_type).toBeDefined();
    }, 5000);

    it('should handle limit parameter', async () => {
      mockGetFeaturedNews.mockResolvedValue({
        news: [],
        count: 0,
        correlation_id: 'featured-limit-correlation-id'
      });

      const limitRef = ref(5);
      useFeaturedNews(limitRef);

      await nextTick();

      expect(mockGetFeaturedNews).toHaveBeenCalledWith(5);
    }, 5000);

    it('should handle limit changes reactively', async () => {
      mockGetFeaturedNews.mockResolvedValue({
        news: [],
        count: 0,
        correlation_id: 'featured-reactive-correlation-id'
      });

      const limitRef = ref(3);
      useFeaturedNews(limitRef);

      await nextTick();
      expect(mockGetFeaturedNews).toHaveBeenCalledWith(3);

      limitRef.value = 7;
      await nextTick();
      expect(mockGetFeaturedNews).toHaveBeenCalledWith(7);
      expect(mockGetFeaturedNews).toHaveBeenCalledTimes(2);
    }, 5000);
  });

  describe('useSearchNews', () => {
    it('should search news articles with query', async () => {
      const mockSearchResults = [
        createMockDatabaseNews({ title: 'Search Result 1' }),
        createMockDatabaseNews({
          news_id: 'news-uuid-126',
          title: 'Search Result 2',
          slug: 'search-result-2',
          news_type: 'update' as const,
          author_name: 'Search Reporter'
        })
      ];

      const mockResponse: NewsResponse = {
        news: mockSearchResults,
        count: 2,
        correlation_id: 'search-correlation-id'
      };

      mockSearchNews.mockResolvedValue(mockResponse);

      const { results, loading, error, total, search } = useSearchNews();

      await search('breaking news update', {
        page: 1,
        pageSize: 10,
        category: 'announcements'
      });

      expect(mockSearchNews).toHaveBeenCalledWith({
        q: 'breaking news update',
        page: 1,
        pageSize: 10,
        category: 'announcements'
      });
      expect(results.value).toHaveLength(2);
      expect(total.value).toBe(2);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);

      // Validate database schema compliance
      expect(results.value[0].news_id).toBeDefined();
      expect(results.value[0].news_type).toBeDefined();
      expect(results.value[0].publication_timestamp).toBeDefined();
    }, 5000);

    it('should handle empty search queries', async () => {
      const { results, total, totalPages, search } = useSearchNews();

      await search('');

      expect(results.value).toEqual([]);
      expect(total.value).toBe(0);
      expect(totalPages.value).toBe(0);
    }, 5000);

    it('should handle search errors', async () => {
      mockSearchNews.mockRejectedValue(new Error('Search failed'));

      const { results, error, loading, search } = useSearchNews();

      await search('test query');

      expect(error.value).toBe('Search failed');
      expect(results.value).toEqual([]);
      expect(loading.value).toBe(false);
    }, 5000);

    it('should calculate pagination correctly for search results', async () => {
      mockSearchNews.mockResolvedValue({
        news: Array(5).fill(null).map((_, i) => createMockDatabaseNews({
          news_id: `search-result-${i}`,
          title: `Search Result ${i}`,
          slug: `search-result-${i}`
        })),
        count: 50,
        correlation_id: 'search-pagination-correlation-id'
      });

      const { total, page, pageSize, totalPages, search } = useSearchNews();

      await search('test query', {
        page: 2,
        pageSize: 5
      });

      expect(total.value).toBe(50);
      expect(page.value).toBe(2);
      expect(pageSize.value).toBe(5);
      expect(totalPages.value).toBe(10); // 50 / 5 = 10
    }, 5000);

    it('should handle search options correctly', async () => {
      mockSearchNews.mockResolvedValue({
        news: [],
        count: 0,
        correlation_id: 'search-options-correlation-id'
      });

      const { search } = useSearchNews();

      const searchOptions: Partial<SearchNewsParams> = {
        page: 3,
        pageSize: 25,
        category: 'press-releases',
        sortBy: 'date-desc'
      };

      await search('corporate announcement', searchOptions);

      expect(mockSearchNews).toHaveBeenCalledWith({
        q: 'corporate announcement',
        page: 3,
        pageSize: 25,
        category: 'press-releases',
        sortBy: 'date-desc'
      });
    }, 5000);
  });

  describe('useNewsCategories', () => {
    it('should fetch news categories with database schema-compliant data', async () => {
      const mockCategories = [
        createMockDatabaseNewsCategory(),
        createMockDatabaseNewsCategory({
          category_id: 'news-category-uuid-457',
          name: 'Press Releases',
          slug: 'press-releases',
          description: 'Official press releases and corporate news',
        })
      ];

      const mockResponse = {
        categories: mockCategories,
        count: 2,
        correlation_id: 'categories-correlation-id'
      };

      mockGetNewsCategories.mockResolvedValue(mockResponse);

      const TestComponent = defineComponent({
        setup() {
          return useNewsCategories();
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      const { categories, loading, error } = (wrapper.vm as any);

      expect(mockGetNewsCategories).toHaveBeenCalled();
      expect(categories).toHaveLength(2);
      expect(loading).toBe(false);
      expect(error).toBe(null);

      // Validate database schema compliance
      expect(categories[0].category_id).toBeDefined();
      expect(categories[0].is_default_unassigned).toBeDefined();
      expect(categories[0].created_on).toBeDefined();
    }, 5000);

    it('should handle empty categories response', async () => {
      mockGetNewsCategories.mockResolvedValue({
        categories: [],
        count: 0,
        correlation_id: 'empty-categories-correlation'
      });

      const TestComponent = defineComponent({
        setup() {
          return useNewsCategories();
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      const { categories } = (wrapper.vm as any);

      expect(categories).toEqual([]);
    }, 5000);

    it('should handle categories API errors', async () => {
      mockGetNewsCategories.mockRejectedValue(new Error('Categories API Error'));

      const TestComponent = defineComponent({
        setup() {
          return useNewsCategories();
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      const { categories, error, loading } = (wrapper.vm as any);

      expect(error).toBe('Categories API Error');
      expect(categories).toEqual([]);
      expect(loading).toBe(false);
    }, 5000);
  });

  describe('Database Schema Field Validation', () => {
    it('should validate news_type enum values in responses', async () => {
      const validNewsTypes = ['announcement', 'press_release', 'event', 'update', 'alert', 'feature'] as const;
      
      for (const newsType of validNewsTypes) {
        const mockNews = createMockDatabaseNews({ news_type: newsType });
        
        mockGetNews.mockResolvedValue({
          news: [mockNews],
          count: 1,
          correlation_id: `news-type-${newsType}-correlation-id`
        });

        const { news, refetch } = useNews({
          immediate: false
        });

        await refetch();
        await nextTick();

        expect(news.value[0].news_type).toBe(newsType);
      }
    }, 5000);

    it('should validate priority_level enum values in responses', async () => {
      const validPriorityLevels = ['low', 'normal', 'high', 'urgent'] as const;
      
      for (const priorityLevel of validPriorityLevels) {
        const mockNews = createMockDatabaseNews({ priority_level: priorityLevel });
        
        mockGetNews.mockResolvedValue({
          news: [mockNews],
          count: 1,
          correlation_id: `priority-${priorityLevel}-correlation-id`
        });

        const { news, refetch } = useNews({
          immediate: false
        });

        await refetch();
        await nextTick();

        expect(news.value[0].priority_level).toBe(priorityLevel);
      }
    }, 5000);

    it('should validate publishing_status enum values in responses', async () => {
      const validStatuses = ['draft', 'published', 'archived'] as const;
      
      for (const status of validStatuses) {
        const mockNews = createMockDatabaseNews({ publishing_status: status });
        
        mockGetNews.mockResolvedValue({
          news: [mockNews],
          count: 1,
          correlation_id: `status-${status}-correlation-id`
        });

        const { news, refetch } = useNews({
          immediate: false
        });

        await refetch();
        await nextTick();

        expect(news.value[0].publishing_status).toBe(status);
      }
    }, 5000);

    it('should handle database schema field types correctly', async () => {
      const mockNews = createMockDatabaseNews({
        tags: ['tag1', 'tag2', 'tag3'], // Array field
        publication_timestamp: '2024-03-15T14:30:00Z', // Timestamp field as ISO string
        is_deleted: false, // Boolean field
        created_on: '2024-01-01T00:00:00Z', // Timestamp field as ISO string
      });

      mockGetNews.mockResolvedValue({
        news: [mockNews],
        count: 1,
        correlation_id: 'field-types-correlation-id'
      });

      const { news, refetch } = useNews({
        immediate: false
      });

      await refetch();
      await nextTick();

      const article = news.value[0];
      expect(Array.isArray(article.tags)).toBe(true);
      expect(article.tags).toHaveLength(3);
      expect(typeof article.publication_timestamp).toBe('string');
      expect(typeof article.is_deleted).toBe('boolean');
      expect(typeof article.created_on).toBe('string');
    }, 5000);
  });

  describe('Reactive State Management', () => {
    it('should maintain proper loading states during transitions', async () => {
      // Simulate slow API call
      let resolvePromise: (value: NewsResponse) => void;
      const slowPromise = new Promise<NewsResponse>((resolve) => {
        resolvePromise = resolve;
      });
      mockGetNews.mockReturnValue(slowPromise);

      const { loading, refetch } = useNews({
        immediate: false
      });

      expect(loading.value).toBe(false);

      const fetchPromise = refetch();
      
      // Should be loading during fetch
      expect(loading.value).toBe(true);

      // Resolve the promise
      resolvePromise!({
        news: [],
        count: 0,
        correlation_id: 'loading-test-correlation-id'
      });

      await fetchPromise;
      await nextTick();

      // Should not be loading after fetch completes
      expect(loading.value).toBe(false);
    }, 5000);

    it('should properly clear errors when making new requests', async () => {
      // First call fails
      mockGetNews.mockRejectedValueOnce(new Error('First error'));
      
      const { error, refetch } = useNews({
        immediate: false
      });

      await refetch();
      await nextTick();

      expect(error.value).toBe('First error');

      // Second call succeeds
      mockGetNews.mockResolvedValueOnce({
        news: [],
        count: 0,
        correlation_id: 'error-clear-correlation-id'
      });

      await refetch();
      await nextTick();

      expect(error.value).toBe(null);
    }, 5000);
  });
});