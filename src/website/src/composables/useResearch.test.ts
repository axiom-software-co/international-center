// Research Composables Tests - State management and API integration validation
// Tests validate useResearch composables with database schema-compliant reactive state

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ref, nextTick } from 'vue';
import { useResearchArticles, useResearchArticle, useFeaturedResearch, useSearchResearch } from './useResearch';
import type { ResearchArticle, ResearchResponse, ResearchArticleResponse, GetResearchParams, SearchResearchParams } from '../lib/clients/research/types';

// Mock the research client
vi.mock('../lib/clients', () => ({
  researchClient: {
    getResearchArticles: vi.fn(),
    getResearchArticleBySlug: vi.fn(),
    getFeaturedResearch: vi.fn(),
    searchResearch: vi.fn(),
  }
}));

import { researchClient } from '../lib/clients';

// Database schema-compliant mock research article for testing
const createMockDatabaseResearch = (overrides: Partial<any> = {}): any => ({
  research_id: 'research-uuid-123',
  title: 'Mock Database Research Article',
  abstract: 'Research article abstract from database schema',
  content: 'Full research article content with methodology and findings',
  slug: 'mock-database-research-article',
  category_id: 'research-category-uuid-456',
  image_url: 'https://example.com/research-image.jpg',
  author_names: 'Dr. Database Schema, Dr. Validation Test',
  publication_date: '2024-03-15',
  doi: '10.1234/mock.research.2024',
  external_url: 'https://journal.example.com/mock-research',
  report_url: 'https://storage.example.com/reports/mock-research.pdf',
  publishing_status: 'published' as const,
  keywords: ['database', 'schema', 'research', 'validation'],
  research_type: 'clinical_study' as const,
  created_on: '2024-01-01T00:00:00Z',
  created_by: 'researcher@example.com',
  modified_on: '2024-01-02T00:00:00Z',
  modified_by: 'admin@example.com',
  is_deleted: false,
  deleted_on: null,
  deleted_by: null,
  ...overrides,
});

describe('useResearch Composables', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('useResearchArticles', () => {
    it('should initialize with proper default state', () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getResearchArticles.mockResolvedValue({
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      });

      const { articles, loading, error, total, page, pageSize, totalPages } = useResearchArticles({
        enabled: false,
        immediate: false
      });

      expect(articles.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      expect(totalPages.value).toBe(0);
    }, 5000);

    it('should fetch research articles with database schema-compliant data', async () => {
      const mockDatabaseResearch = [
        createMockDatabaseResearch(),
        createMockDatabaseResearch({
          research_id: 'research-uuid-124',
          title: 'Second Database Research',
          slug: 'second-database-research',
          research_type: 'systematic_review' as const,
          author_names: 'Dr. Second Author',
        })
      ];

      const mockResponse: ResearchResponse = {
        data: mockDatabaseResearch,
        pagination: {
          page: 1,
          pageSize: 10,
          total: 2,
          totalPages: 1
        },
        success: true
      };

      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getResearchArticles.mockResolvedValue(mockResponse);

      const { articles, loading, error, total, totalPages, refetch } = useResearchArticles({
        enabled: false,
        immediate: false
      });

      expect(loading.value).toBe(false);

      await refetch();

      await nextTick();

      expect(mockResearchClient.getResearchArticles).toHaveBeenCalledTimes(1);
      expect(articles.value).toHaveLength(2);
      expect(total.value).toBe(2);
      expect(totalPages.value).toBe(1);
      expect(error.value).toBe(null);
      expect(loading.value).toBe(false);

      // Validate database schema fields are present
      const firstArticle = articles.value[0];
      expect(firstArticle.research_id).toBeDefined();
      expect(firstArticle.abstract).toBeDefined(); // Not 'excerpt'
      expect(firstArticle.author_names).toBeDefined(); // Not 'author'
      expect(firstArticle.publishing_status).toBeDefined();
      expect(firstArticle.keywords).toBeDefined(); // Not 'tags'
      expect(firstArticle.research_type).toBeDefined();
      expect(firstArticle.image_url).toBeDefined(); // Not 'featured_image'
      expect(firstArticle.publication_date).toBeDefined();
      expect(firstArticle.doi).toBeDefined();
      expect(firstArticle.external_url).toBeDefined();
      expect(firstArticle.report_url).toBeDefined();
      expect(firstArticle.is_deleted).toBeDefined();
      expect(firstArticle.created_on).toBeDefined();
    }, 5000);

    it('should handle API errors gracefully', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getResearchArticles.mockRejectedValue(new Error('API Error'));

      const { articles, loading, error, refetch } = useResearchArticles({
        enabled: false,
        immediate: false
      });

      await refetch();
      await nextTick();

      expect(error.value).toBe('API Error');
      expect(articles.value).toEqual([]);
      expect(loading.value).toBe(false);
    }, 5000);

    it('should handle query parameters correctly', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getResearchArticles.mockResolvedValue({
        data: [],
        pagination: {
          page: 2,
          pageSize: 20,
          total: 0,
          totalPages: 0
        },
        success: true
      });

      const params: GetResearchParams = {
        page: 2,
        pageSize: 20,
        category: 'clinical-research',
        featured: true,
        industry: 'healthcare'
      };

      const { refetch } = useResearchArticles({
        enabled: false,
        immediate: false,
        ...params
      });

      await refetch();

      expect(mockResearchClient.getResearchArticles).toHaveBeenCalledWith(
        expect.objectContaining(params)
      );
    }, 5000);

    it('should handle pagination calculations correctly', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getResearchArticles.mockResolvedValue({
        data: Array(15).fill(null).map((_, i) => createMockDatabaseResearch({
          research_id: `research-${i}`,
          title: `Research Article ${i}`,
          slug: `research-article-${i}`
        })),
        pagination: {
          page: 1,
          pageSize: 15,
          total: 150,
          totalPages: 10
        },
        success: true
      });

      const { total, pageSize, totalPages, refetch } = useResearchArticles({
        enabled: false,
        immediate: false,
        pageSize: 15
      });

      await refetch();
      await nextTick();

      expect(total.value).toBe(150);
      expect(pageSize.value).toBe(15);
      expect(totalPages.value).toBe(10);
    }, 5000);
  });

  describe('useResearchArticle', () => {
    it('should fetch single research article by slug', async () => {
      const mockResearch = createMockDatabaseResearch({
        slug: 'single-research-test'
      });

      const mockResponse: ResearchArticleResponse = {
        data: mockResearch,
        success: true
      };

      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getResearchArticleBySlug.mockResolvedValue(mockResponse);

      const slugRef = ref('single-research-test');
      const { article, loading, error } = useResearchArticle(slugRef);

      await nextTick();

      expect(mockResearchClient.getResearchArticleBySlug).toHaveBeenCalledWith('single-research-test');
      expect(article.value).toEqual(mockResearch);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);

      // Validate database schema compliance
      expect(article.value?.research_id).toBeDefined();
      expect(article.value?.abstract).toBeDefined(); // Not 'excerpt'
      expect(article.value?.author_names).toBeDefined(); // Not 'author'
      expect(article.value?.keywords).toBeDefined(); // Not 'tags'
      expect(article.value?.research_type).toBeDefined();
    }, 5000);

    it('should handle slug changes reactively', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getResearchArticleBySlug.mockResolvedValue({
        data: createMockDatabaseResearch(),
        success: true
      });

      const slugRef = ref('initial-slug');
      const { refetch } = useResearchArticle(slugRef);

      await nextTick();

      expect(mockResearchClient.getResearchArticleBySlug).toHaveBeenCalledWith('initial-slug');

      // Change slug
      slugRef.value = 'updated-slug';
      await nextTick();

      expect(mockResearchClient.getResearchArticleBySlug).toHaveBeenCalledWith('updated-slug');
      expect(mockResearchClient.getResearchArticleBySlug).toHaveBeenCalledTimes(2);
    }, 5000);

    it('should handle empty slug gracefully', async () => {
      const mockResearchClient = vi.mocked(researchClient);

      const { article, loading } = useResearchArticle(ref(null));

      await nextTick();

      expect(mockResearchClient.getResearchArticleBySlug).not.toHaveBeenCalled();
      expect(article.value).toBe(null);
      expect(loading.value).toBe(false);
    }, 5000);

    it('should handle API errors', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getResearchArticleBySlug.mockRejectedValue(new Error('Research article not found'));

      const { article, error, loading } = useResearchArticle('non-existent-slug');

      await nextTick();

      expect(error.value).toBe('Research article not found');
      expect(article.value).toBe(null);
      expect(loading.value).toBe(false);
    }, 5000);
  });

  describe('useFeaturedResearch', () => {
    it('should fetch featured research articles', async () => {
      const mockFeaturedResearch = [
        createMockDatabaseResearch({ title: 'Featured Research 1' }),
        createMockDatabaseResearch({ 
          research_id: 'research-uuid-125',
          title: 'Featured Research 2',
          slug: 'featured-research-2',
          research_type: 'meta_analysis' as const,
          author_names: 'Dr. Featured Author'
        })
      ];

      const mockResponse: ResearchResponse = {
        data: mockFeaturedResearch,
        pagination: {
          page: 1,
          pageSize: 10,
          total: 2,
          totalPages: 1
        },
        success: true
      };

      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getFeaturedResearch.mockResolvedValue(mockResponse);

      const { articles, loading, error } = useFeaturedResearch();

      await nextTick();

      expect(mockResearchClient.getFeaturedResearch).toHaveBeenCalledWith(undefined);
      expect(articles.value).toHaveLength(2);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);

      // Validate database schema compliance
      expect(articles.value[0].research_id).toBeDefined();
      expect(articles.value[0].publishing_status).toBeDefined();
      expect(articles.value[0].research_type).toBeDefined();
    }, 5000);

    it('should handle limit parameter', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getFeaturedResearch.mockResolvedValue({
        data: [],
        pagination: {
          page: 1,
          pageSize: 5,
          total: 0,
          totalPages: 0
        },
        success: true
      });

      const limitRef = ref(5);
      useFeaturedResearch(limitRef);

      await nextTick();

      expect(mockResearchClient.getFeaturedResearch).toHaveBeenCalledWith(5);
    }, 5000);

    it('should handle limit changes reactively', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getFeaturedResearch.mockResolvedValue({
        data: [],
        pagination: {
          page: 1,
          pageSize: 3,
          total: 0,
          totalPages: 0
        },
        success: true
      });

      const limitRef = ref(3);
      useFeaturedResearch(limitRef);

      await nextTick();
      expect(mockResearchClient.getFeaturedResearch).toHaveBeenCalledWith(3);

      limitRef.value = 7;
      await nextTick();
      expect(mockResearchClient.getFeaturedResearch).toHaveBeenCalledWith(7);
      expect(mockResearchClient.getFeaturedResearch).toHaveBeenCalledTimes(2);
    }, 5000);
  });

  describe('useSearchResearch', () => {
    it('should search research articles with query', async () => {
      const mockSearchResults = [
        createMockDatabaseResearch({ title: 'Search Result 1' }),
        createMockDatabaseResearch({
          research_id: 'research-uuid-126',
          title: 'Search Result 2',
          slug: 'search-result-2',
          research_type: 'case_report' as const,
          author_names: 'Dr. Search Result'
        })
      ];

      const mockResponse: ResearchResponse = {
        data: mockSearchResults,
        pagination: {
          page: 1,
          pageSize: 10,
          total: 2,
          totalPages: 1
        },
        success: true
      };

      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.searchResearch.mockResolvedValue(mockResponse);

      const { results, loading, error, total, search } = useSearchResearch();

      await search('clinical study diabetes', {
        page: 1,
        pageSize: 10,
        category: 'clinical-research'
      });

      expect(mockResearchClient.searchResearch).toHaveBeenCalledWith({
        q: 'clinical study diabetes',
        page: 1,
        pageSize: 10,
        category: 'clinical-research'
      });
      expect(results.value).toHaveLength(2);
      expect(total.value).toBe(2);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);

      // Validate database schema compliance
      expect(results.value[0].research_id).toBeDefined();
      expect(results.value[0].research_type).toBeDefined();
      expect(results.value[0].abstract).toBeDefined(); // Not 'excerpt'
      expect(results.value[0].author_names).toBeDefined(); // Not 'author'
    }, 5000);

    it('should handle empty search queries', async () => {
      const { results, total, totalPages, search } = useSearchResearch();

      await search('');

      expect(results.value).toEqual([]);
      expect(total.value).toBe(0);
      expect(totalPages.value).toBe(0);
    }, 5000);

    it('should handle search errors', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.searchResearch.mockRejectedValue(new Error('Search failed'));

      const { results, error, loading, search } = useSearchResearch();

      await search('test query');

      expect(error.value).toBe('Search failed');
      expect(results.value).toEqual([]);
      expect(loading.value).toBe(false);
    }, 5000);

    it('should calculate pagination correctly for search results', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.searchResearch.mockResolvedValue({
        data: Array(5).fill(null).map((_, i) => createMockDatabaseResearch({
          research_id: `search-result-${i}`,
          title: `Search Result ${i}`,
          slug: `search-result-${i}`
        })),
        pagination: {
          page: 2,
          pageSize: 5,
          total: 50,
          totalPages: 10
        },
        success: true
      });

      const { total, page, pageSize, totalPages, search } = useSearchResearch();

      await search('test query', {
        page: 2,
        pageSize: 5
      });

      expect(total.value).toBe(50);
      expect(page.value).toBe(2);
      expect(pageSize.value).toBe(5);
      expect(totalPages.value).toBe(10);
    }, 5000);

    it('should handle search options correctly', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.searchResearch.mockResolvedValue({
        data: [],
        pagination: {
          page: 3,
          pageSize: 25,
          total: 0,
          totalPages: 0
        },
        success: true
      });

      const { search } = useSearchResearch();

      const searchOptions: Partial<SearchResearchParams> = {
        page: 3,
        pageSize: 25,
        category: 'clinical-research',
        sortBy: 'date-desc'
      };

      await search('diabetes research', searchOptions);

      expect(mockResearchClient.searchResearch).toHaveBeenCalledWith({
        q: 'diabetes research',
        page: 3,
        pageSize: 25,
        category: 'clinical-research',
        sortBy: 'date-desc'
      });
    }, 5000);
  });

  describe('Database Schema Field Validation', () => {
    it('should validate research_type enum values in responses', async () => {
      const validResearchTypes = ['clinical_study', 'case_report', 'systematic_review', 'meta_analysis', 'editorial', 'commentary'] as const;
      
      for (const researchType of validResearchTypes) {
        const mockResearch = createMockDatabaseResearch({ research_type: researchType });
        
        const mockResearchClient = vi.mocked(researchClient);
        mockResearchClient.getResearchArticles.mockResolvedValue({
          data: [mockResearch],
          pagination: {
            page: 1,
            pageSize: 1,
            total: 1,
            totalPages: 1
          },
          success: true
        });

        const { articles, refetch } = useResearchArticles({
          enabled: false,
          immediate: false
        });

        await refetch();
        await nextTick();

        expect(articles.value[0].research_type).toBe(researchType);
      }
    }, 5000);

    it('should validate publishing_status enum values in responses', async () => {
      const validStatuses = ['draft', 'published', 'archived'] as const;
      
      for (const status of validStatuses) {
        const mockResearch = createMockDatabaseResearch({ publishing_status: status });
        
        const mockResearchClient = vi.mocked(researchClient);
        mockResearchClient.getResearchArticles.mockResolvedValue({
          data: [mockResearch],
          pagination: {
            page: 1,
            pageSize: 1,
            total: 1,
            totalPages: 1
          },
          success: true
        });

        const { articles, refetch } = useResearchArticles({
          enabled: false,
          immediate: false
        });

        await refetch();
        await nextTick();

        expect(articles.value[0].publishing_status).toBe(status);
      }
    }, 5000);

    it('should handle database schema field types correctly', async () => {
      const mockResearch = createMockDatabaseResearch({
        keywords: ['keyword1', 'keyword2', 'keyword3'], // Array field
        publication_date: '2024-03-15', // Date field as ISO string
        is_deleted: false, // Boolean field
        created_on: '2024-01-01T00:00:00Z', // Timestamp field as ISO string
      });

      const mockResearchClient = vi.mocked(researchClient);
      mockResearchClient.getResearchArticles.mockResolvedValue({
        data: [mockResearch],
        pagination: {
          page: 1,
          pageSize: 1,
          total: 1,
          totalPages: 1
        },
        success: true
      });

      const { articles, refetch } = useResearchArticles({
        enabled: false,
        immediate: false
      });

      await refetch();
      await nextTick();

      const article = articles.value[0];
      expect(Array.isArray(article.keywords)).toBe(true);
      expect(article.keywords).toHaveLength(3);
      expect(typeof article.publication_date).toBe('string');
      expect(typeof article.is_deleted).toBe('boolean');
      expect(typeof article.created_on).toBe('string');
    }, 5000);
  });

  describe('Reactive State Management', () => {
    it('should maintain proper loading states during transitions', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      
      // Simulate slow API call
      let resolvePromise: (value: ResearchResponse) => void;
      const slowPromise = new Promise<ResearchResponse>((resolve) => {
        resolvePromise = resolve;
      });
      mockResearchClient.getResearchArticles.mockReturnValue(slowPromise);

      const { loading, refetch } = useResearchArticles({
        enabled: false,
        immediate: false
      });

      expect(loading.value).toBe(false);

      const fetchPromise = refetch();
      
      // Should be loading during fetch
      expect(loading.value).toBe(true);

      // Resolve the promise
      resolvePromise!({
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      });

      await fetchPromise;
      await nextTick();

      // Should not be loading after fetch completes
      expect(loading.value).toBe(false);
    }, 5000);

    it('should properly clear errors when making new requests', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      
      // First call fails
      mockResearchClient.getResearchArticles.mockRejectedValueOnce(new Error('First error'));
      
      const { error, refetch } = useResearchArticles({
        enabled: false,
        immediate: false
      });

      await refetch();
      await nextTick();

      expect(error.value).toBe('First error');

      // Second call succeeds
      mockResearchClient.getResearchArticles.mockResolvedValueOnce({
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      });

      await refetch();
      await nextTick();

      expect(error.value).toBe(null);
    }, 5000);

    it('should handle backend response variations correctly', async () => {
      const mockResearchClient = vi.mocked(researchClient);
      
      // Test missing pagination in response
      mockResearchClient.getResearchArticles.mockResolvedValueOnce({
        data: [createMockDatabaseResearch()],
        success: true
      } as any);

      const { articles, total, totalPages, refetch } = useResearchArticles({
        enabled: false,
        immediate: false
      });

      await refetch();
      await nextTick();

      expect(articles.value).toHaveLength(1);
      expect(total.value).toBe(1); // Should default to data.length
      expect(totalPages.value).toBeGreaterThan(0);
    }, 5000);
  });
});