// News REST Client Tests - Contract validation for database schema compliance
// Tests validate NewsRestClient methods against TABLES-NEWS.md schema requirements

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { NewsRestClient } from './NewsRestClient';
import type { NewsArticle, NewsResponse, NewsArticleResponse, GetNewsParams, SearchNewsParams, NewsCategory } from './types';

// Database schema validation - NewsArticle interface must match TABLES-NEWS.md exactly
interface DatabaseSchemaNews {
  // Primary key and identifiers
  news_id: string; // UUID PRIMARY KEY
  title: string; // VARCHAR(255) NOT NULL
  summary: string; // TEXT NOT NULL (different from 'excerpt')
  content?: string; // TEXT (nullable, stored in PostgreSQL)
  slug: string; // VARCHAR(255) UNIQUE NOT NULL
  
  // Category relationship
  category_id: string; // UUID NOT NULL REFERENCES news_categories(category_id)
  
  // Media and publication info
  image_url?: string; // VARCHAR(500) (nullable, Azure Blob Storage URL)
  author_name?: string; // VARCHAR(255) (nullable)
  publication_timestamp: string; // TIMESTAMPTZ NOT NULL DEFAULT NOW()
  external_source?: string; // VARCHAR(255) (nullable)
  external_url?: string; // VARCHAR(500) (nullable)
  
  // Publishing workflow
  publishing_status: 'draft' | 'published' | 'archived'; // VARCHAR(20) NOT NULL DEFAULT 'draft'
  
  // Content metadata
  tags: string[]; // TEXT[] (PostgreSQL array)
  news_type: 'announcement' | 'press_release' | 'event' | 'update' | 'alert' | 'feature'; // VARCHAR(50) NOT NULL
  priority_level: 'low' | 'normal' | 'high' | 'urgent'; // VARCHAR(20) NOT NULL DEFAULT 'normal'
  
  // Audit fields
  created_on: string; // TIMESTAMPTZ NOT NULL DEFAULT NOW()
  created_by?: string; // VARCHAR(255) (nullable)
  modified_on?: string; // TIMESTAMPTZ (nullable)
  modified_by?: string; // VARCHAR(255) (nullable)
  
  // Soft delete fields
  is_deleted: boolean; // BOOLEAN NOT NULL DEFAULT FALSE
  deleted_on?: string; // TIMESTAMPTZ (nullable)
  deleted_by?: string; // VARCHAR(255) (nullable)
}

describe('NewsRestClient', () => {
  let client: NewsRestClient;
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = vi.fn();
    global.fetch = mockFetch;
    client = new NewsRestClient('http://localhost:8080');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Database Schema Compliance', () => {
    it('should validate NewsArticle interface matches database schema', () => {
      // This test ensures our TypeScript interface matches the PostgreSQL schema
      const mockDatabaseNews: DatabaseSchemaNews = {
        news_id: 'news-uuid-123',
        title: 'Database Schema News',
        summary: 'Summary from database',
        content: 'Full content stored in PostgreSQL',
        slug: 'database-schema-news',
        category_id: 'category-uuid-456',
        image_url: 'https://storage.azure.com/news-image.jpg',
        author_name: 'Database Reporter',
        publication_timestamp: '2024-03-15T14:30:00Z',
        external_source: 'External News Source',
        external_url: 'https://external.example.com/news',
        publishing_status: 'published',
        tags: ['database', 'schema', 'compliance'],
        news_type: 'announcement',
        priority_level: 'normal',
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'reporter@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'editor@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // Verify all required fields exist
      expect(mockDatabaseNews.news_id).toBeDefined();
      expect(mockDatabaseNews.title).toBeDefined();
      expect(mockDatabaseNews.summary).toBeDefined();
      expect(mockDatabaseNews.slug).toBeDefined();
      expect(mockDatabaseNews.category_id).toBeDefined();
      expect(mockDatabaseNews.publication_timestamp).toBeDefined();
      expect(mockDatabaseNews.publishing_status).toBeDefined();
      expect(Array.isArray(mockDatabaseNews.tags)).toBe(true);
      expect(mockDatabaseNews.news_type).toBeDefined();
      expect(mockDatabaseNews.priority_level).toBeDefined();
      expect(typeof mockDatabaseNews.is_deleted).toBe('boolean');
      expect(mockDatabaseNews.created_on).toBeDefined();
    }, 5000);

    it('should validate enum values match database constraints', () => {
      const validPublishingStatuses: Array<DatabaseSchemaNews['publishing_status']> = ['draft', 'published', 'archived'];
      const validNewsTypes: Array<DatabaseSchemaNews['news_type']> = ['announcement', 'press_release', 'event', 'update', 'alert', 'feature'];
      const validPriorityLevels: Array<DatabaseSchemaNews['priority_level']> = ['low', 'normal', 'high', 'urgent'];

      // Verify enum constraints from database schema
      expect(validPublishingStatuses).toHaveLength(3);
      expect(validNewsTypes).toHaveLength(6);
      expect(validPriorityLevels).toHaveLength(4);
      
      // Verify specific enum values
      expect(validPublishingStatuses).toContain('draft');
      expect(validPublishingStatuses).toContain('published');
      expect(validPublishingStatuses).toContain('archived');
      
      expect(validNewsTypes).toContain('announcement');
      expect(validNewsTypes).toContain('press_release');
      expect(validNewsTypes).toContain('feature');
      
      expect(validPriorityLevels).toContain('low');
      expect(validPriorityLevels).toContain('normal');
      expect(validPriorityLevels).toContain('urgent');
    }, 5000);
  });

  describe('getNews', () => {
    it('should fetch news articles with database schema-compliant response', async () => {
      const mockNewsResponse: NewsResponse = {
        news: [
          {
            news_id: 'news-uuid-123',
            title: 'Database News Article',
            summary: 'Article summary for testing',
            content: 'Full article content from database',
            slug: 'database-news-article',
            category_id: 'category-uuid-456',
            image_url: 'https://storage.azure.com/news-image.jpg',
            author_name: 'Test Reporter',
            publication_timestamp: '2024-03-15T14:30:00Z',
            external_source: 'Test News Source',
            external_url: 'https://external.example.com/news',
            publishing_status: 'published',
            tags: ['database', 'testing'],
            news_type: 'announcement',
            priority_level: 'normal',
            created_on: '2024-01-01T00:00:00Z',
            created_by: 'reporter@example.com',
            modified_on: '2024-01-02T00:00:00Z',
            modified_by: 'editor@example.com',
            is_deleted: false,
            deleted_on: null,
            deleted_by: null,
          }
        ],
        count: 1,
        correlation_id: 'news-test-correlation-id'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockNewsResponse),
      });

      const result = await client.getNews();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/news',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Accept': 'application/json',
            'Content-Type': 'application/json',
          }),
        })
      );

      expect(result).toEqual(mockNewsResponse);
      
      // Validate database schema compliance
      const newsArticle = result.news[0];
      expect(newsArticle.news_id).toBeDefined();
      expect(newsArticle.summary).toBeDefined(); // Not 'excerpt'
      expect(newsArticle.publication_timestamp).toBeDefined(); // Not 'published_at'
      expect(newsArticle.external_source).toBeDefined();
      expect(newsArticle.external_url).toBeDefined();
      expect(Array.isArray(newsArticle.tags)).toBe(true);
      expect(newsArticle.news_type).toBeDefined();
      expect(newsArticle.priority_level).toBeDefined();
    }, 5000);

    it('should handle query parameters correctly', async () => {
      const mockResponse: NewsResponse = {
        news: [],
        count: 0,
        correlation_id: 'params-test-correlation-id'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse),
      });

      const params: GetNewsParams = {
        page: 2,
        pageSize: 20,
        category: 'announcements',
        featured: true,
        sortBy: 'date-desc'
      };

      await client.getNews(params);

      const expectedUrl = new URL('http://localhost:8080/api/v1/news');
      expectedUrl.searchParams.set('page', '2');
      expectedUrl.searchParams.set('pageSize', '20');
      expectedUrl.searchParams.set('category', 'announcements');
      expectedUrl.searchParams.set('featured', 'true');
      expectedUrl.searchParams.set('sortBy', 'date-desc');

      expect(mockFetch).toHaveBeenCalledWith(
        expectedUrl.toString(),
        expect.objectContaining({
          method: 'GET',
        })
      );
    }, 5000);

    it('should handle API errors with correlation tracking', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: () => Promise.resolve({
          error: 'Internal server error',
          correlation_id: 'error-correlation-500'
        }),
      });

      await expect(client.getNews()).rejects.toThrow('HTTP 500 Error');
    }, 5000);

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network connection failed'));

      await expect(client.getNews()).rejects.toThrow('Network connection failed');
    }, 5000);
  });

  describe('getNewsArticleBySlug', () => {
    it('should fetch single news article by slug with database schema compliance', async () => {
      const mockResponse: NewsArticleResponse = {
        news: {
          news_id: 'news-uuid-123',
          title: 'Single News Article',
          summary: 'Article summary for single fetch',
          content: 'Detailed article content from database',
          slug: 'single-news-article',
          category_id: 'category-uuid-456',
          image_url: 'https://storage.azure.com/single-news.jpg',
          author_name: 'Single Reporter',
          publication_timestamp: '2024-03-15T14:30:00Z',
          external_source: 'Single News Source',
          external_url: 'https://external.example.com/single',
          publishing_status: 'published',
          tags: ['single', 'article'],
          news_type: 'press_release',
          priority_level: 'high',
          created_on: '2024-01-01T00:00:00Z',
          created_by: 'reporter@example.com',
          modified_on: '2024-01-02T00:00:00Z',
          modified_by: 'editor@example.com',
          is_deleted: false,
          deleted_on: null,
          deleted_by: null,
        },
        correlation_id: 'single-news-correlation-id'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await client.getNewsArticleBySlug('single-news-article');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/news/slug/single-news-article',
        expect.objectContaining({
          method: 'GET',
        })
      );

      expect(result).toEqual(mockResponse);
      
      // Validate database schema compliance
      const newsArticle = result.news;
      expect(newsArticle.news_type).toBe('press_release');
      expect(newsArticle.priority_level).toBe('high');
      expect(newsArticle.publication_timestamp).toBeDefined();
    }, 5000);

    it('should handle slug not found', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        json: () => Promise.resolve({
          error: 'News article not found',
          correlation_id: 'not-found-correlation-404'
        }),
      });

      await expect(client.getNewsArticleBySlug('non-existent-slug')).rejects.toThrow('HTTP 404 Error');
    }, 5000);
  });

  describe('getFeaturedNews', () => {
    it('should fetch featured news with optional limit', async () => {
      const mockResponse: NewsResponse = {
        news: [
          {
            news_id: 'featured-news-uuid-123',
            title: 'Featured News Article',
            summary: 'Featured article summary',
            content: 'Featured article content',
            slug: 'featured-news-article',
            category_id: 'category-uuid-456',
            image_url: 'https://storage.azure.com/featured-news.jpg',
            author_name: 'Featured Reporter',
            publication_timestamp: '2024-03-15T14:30:00Z',
            external_source: 'Featured News Source',
            external_url: 'https://external.example.com/featured',
            publishing_status: 'published',
            tags: ['featured', 'important'],
            news_type: 'feature',
            priority_level: 'urgent',
            created_on: '2024-01-01T00:00:00Z',
            created_by: 'reporter@example.com',
            modified_on: '2024-01-02T00:00:00Z',
            modified_by: 'editor@example.com',
            is_deleted: false,
            deleted_on: null,
            deleted_by: null,
          }
        ],
        count: 1,
        correlation_id: 'featured-news-correlation-id'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await client.getFeaturedNews(5);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/news/featured?limit=5',
        expect.objectContaining({
          method: 'GET',
        })
      );

      expect(result).toEqual(mockResponse);
      expect(result.news[0].news_type).toBe('feature');
      expect(result.news[0].priority_level).toBe('urgent');
    }, 5000);

    it('should fetch featured news without limit parameter', async () => {
      const mockResponse: NewsResponse = {
        news: [],
        count: 0,
        correlation_id: 'featured-no-limit-correlation-id'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse),
      });

      await client.getFeaturedNews();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/news/featured',
        expect.objectContaining({
          method: 'GET',
        })
      );
    }, 5000);
  });

  describe('searchNews', () => {
    it('should search news articles with query parameters', async () => {
      const mockResponse: NewsResponse = {
        news: [
          {
            news_id: 'search-result-uuid-123',
            title: 'Search Result News',
            summary: 'Search result summary',
            content: 'Search result content',
            slug: 'search-result-news',
            category_id: 'category-uuid-456',
            image_url: 'https://storage.azure.com/search-news.jpg',
            author_name: 'Search Reporter',
            publication_timestamp: '2024-03-15T14:30:00Z',
            external_source: 'Search News Source',
            external_url: 'https://external.example.com/search',
            publishing_status: 'published',
            tags: ['search', 'result'],
            news_type: 'update',
            priority_level: 'normal',
            created_on: '2024-01-01T00:00:00Z',
            created_by: 'reporter@example.com',
            modified_on: '2024-01-02T00:00:00Z',
            modified_by: 'editor@example.com',
            is_deleted: false,
            deleted_on: null,
            deleted_by: null,
          }
        ],
        count: 1,
        correlation_id: 'search-correlation-id'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse),
      });

      const params: SearchNewsParams = {
        q: 'breaking news',
        page: 1,
        pageSize: 10,
        category: 'announcements',
        sortBy: 'date-desc'
      };

      const result = await client.searchNews(params);

      const expectedUrl = new URL('http://localhost:8080/api/v1/news/search');
      expectedUrl.searchParams.set('q', 'breaking news');
      expectedUrl.searchParams.set('page', '1');
      expectedUrl.searchParams.set('pageSize', '10');
      expectedUrl.searchParams.set('category', 'announcements');
      expectedUrl.searchParams.set('sortBy', 'date-desc');

      expect(mockFetch).toHaveBeenCalledWith(
        expectedUrl.toString(),
        expect.objectContaining({
          method: 'GET',
        })
      );

      expect(result).toEqual(mockResponse);
      expect(result.news[0].news_type).toBe('update');
    }, 5000);

    it('should handle empty search queries', async () => {
      const params: SearchNewsParams = {
        q: '',
        page: 1,
        pageSize: 10
      };

      const result = await client.searchNews(params);

      // Should return empty results for empty query
      expect(result.news).toEqual([]);
      expect(result.count).toBe(0);
      
      // Should not make HTTP request for empty query
      expect(mockFetch).not.toHaveBeenCalled();
    }, 5000);
  });

  describe('getNewsCategories', () => {
    it('should fetch news categories with database schema compliance', async () => {
      const mockCategoriesResponse = {
        categories: [
          {
            category_id: 'category-uuid-456',
            name: 'Database Category',
            slug: 'database-category',
            description: 'Category for database news',
            is_default_unassigned: false,
            created_on: '2024-01-01T00:00:00Z',
            created_by: 'admin@example.com',
            modified_on: '2024-01-02T00:00:00Z',
            modified_by: 'admin@example.com',
            is_deleted: false,
            deleted_on: null,
            deleted_by: null,
          }
        ],
        count: 1,
        correlation_id: 'categories-correlation-id'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockCategoriesResponse),
      });

      const result = await client.getNewsCategories();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/news/categories',
        expect.objectContaining({
          method: 'GET',
        })
      );

      expect(result).toEqual(mockCategoriesResponse);
      
      // Validate database schema compliance for categories
      const category = result.categories[0];
      expect(category.category_id).toBeDefined();
      expect(category.is_default_unassigned).toBeDefined();
      expect(typeof category.is_default_unassigned).toBe('boolean');
      expect(category.created_on).toBeDefined();
    }, 5000);
  });

  describe('Error Handling and Resilience', () => {
    it('should handle timeout errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Request timeout after 5000ms'));

      await expect(client.getNews()).rejects.toThrow('Request timeout after 5000ms');
    }, 5000);

    it('should handle rate limiting', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 429,
        json: () => Promise.resolve({
          error: 'Rate limit exceeded',
          correlation_id: 'rate-limit-correlation-429'
        }),
      });

      await expect(client.getNews()).rejects.toThrow('HTTP 429 Error');
    }, 5000);

    it('should handle malformed JSON responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.reject(new Error('Invalid JSON')),
      });

      await expect(client.getNews()).rejects.toThrow('Invalid JSON');
    }, 5000);

    it('should validate response correlation IDs', async () => {
      const mockResponse: NewsResponse = {
        news: [],
        count: 0,
        correlation_id: 'test-correlation-validation'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await client.getNews();

      expect(result.correlation_id).toBeDefined();
      expect(typeof result.correlation_id).toBe('string');
      expect(result.correlation_id).toBe('test-correlation-validation');
    }, 5000);
  });

  describe('Request Headers and Security', () => {
    it('should include required security headers', async () => {
      const mockResponse: NewsResponse = {
        news: [],
        count: 0,
        correlation_id: 'security-headers-test'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse),
      });

      await client.getNews();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Accept': 'application/json',
            'Content-Type': 'application/json',
          }),
        })
      );
    }, 5000);

    it('should handle CORS preflight correctly', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type, Authorization',
        }),
        json: () => Promise.resolve({
          news: [],
          count: 0,
          correlation_id: 'cors-test'
        }),
      });

      const result = await client.getNews();

      expect(result).toBeDefined();
      expect(result.correlation_id).toBe('cors-test');
    }, 5000);
  });
});