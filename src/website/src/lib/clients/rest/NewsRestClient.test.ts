import { describe, it, expect, vi, beforeEach } from 'vitest';
import { NewsRestClient } from './NewsRestClient';
import { RestError } from './BaseRestClient';
import { createMockFetchResponse } from '../../../test/setup';

describe('NewsRestClient', () => {
  let client: NewsRestClient;
  
  beforeEach(() => {
    client = new NewsRestClient();
    vi.clearAllMocks();
  });

  describe('getNews', () => {
    it('should call correct backend endpoint with v1 API path', async () => {
      const mockResponse = {
        news: [
          {
            news_id: '123',
            title: 'Test News Article',
            summary: 'Test Summary',
            slug: 'test-news',
            publishing_status: 'published',
            category_id: '456',
            author_name: 'John Doe',
            featured: false,
            order_number: 1
          }
        ],
        count: 1,
        correlation_id: 'test-correlation-123'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

      await client.getNews({ page: 1, pageSize: 10 });

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/news?page=1&pageSize=10',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          })
        })
      );
    });

    it('should return backend response format with news array and correlation_id', async () => {
      const expectedResponse = {
        news: [
          {
            news_id: '123',
            title: 'Breaking News',
            summary: 'Important news update',
            slug: 'breaking-news',
            publishing_status: 'published',
            category_id: '456',
            author_name: 'Jane Smith',
            content: '<h2>Important Update</h2><p>This is breaking news content with <strong>important information</strong>.</p>',
            image_url: 'https://storage.azure.com/images/breaking-news.jpg',
            featured: true,
            order_number: 1
          }
        ],
        count: 1,
        correlation_id: 'test-correlation-123'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(expectedResponse)
      );

      const result = await client.getNews();
      
      expect(result).toEqual(expectedResponse);
    });

    it('should handle search parameter correctly', async () => {
      const mockResponse = {
        news: [],
        count: 0,
        correlation_id: 'search-correlation-456'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

      await client.getNews({ search: 'healthcare' });

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/news?search=healthcare',
        expect.any(Object)
      );
    });

    it('should handle category filter correctly', async () => {
      const mockResponse = {
        news: [],
        count: 0,
        correlation_id: 'category-correlation-789'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

      await client.getNews({ category: 'health-updates' });

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/news/categories/health-updates/news',
        expect.any(Object)
      );
    });

    it('should handle featured parameter correctly', async () => {
      const mockResponse = {
        news: [],
        count: 0,
        correlation_id: 'featured-correlation-abc'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

      await client.getNews({ featured: true });

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/news?featured=true',
        expect.any(Object)
      );
    });
  });

  describe('getNewsBySlug', () => {
    it('should call correct v1 API endpoint for news by slug', async () => {
      const mockResponse = {
        news: {
          news_id: '123',
          title: 'Health News Update',
          slug: 'health-news-update',
          summary: 'Important health information',
          publishing_status: 'published',
          category_id: '456',
          author_name: 'Dr. Smith',
          content: '<h2>Health Update</h2><p>This article contains important health information including:</p><ul><li>Latest research findings</li><li>Treatment recommendations</li><li>Prevention guidelines</li></ul>',
          image_url: 'https://storage.azure.com/images/health-news.jpg',
          featured: false,
          order_number: 1
        },
        correlation_id: 'slug-correlation-def'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

      await client.getNewsBySlug('health-news-update');

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/news/slug/health-news-update',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });

    it('should return news article with content field from PostgreSQL storage', async () => {
      const expectedResponse = {
        news: {
          news_id: '123',
          title: 'Research Breakthrough',
          slug: 'research-breakthrough',
          summary: 'New medical research findings',
          publishing_status: 'published',
          category_id: '456',
          author_name: 'Research Team',
          content: '<h2>Breakthrough Research</h2>\n<p>Our research team has made significant discoveries in:</p>\n<ul>\n<li>Treatment effectiveness analysis</li>\n<li>Patient outcome improvements</li>\n<li>Clinical trial results</li>\n</ul>\n<p>This research will impact future healthcare practices.</p>',
          image_url: 'https://storage.azure.com/images/research.jpg',
          featured: true,
          order_number: 1
        },
        correlation_id: 'slug-correlation-def'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(expectedResponse)
      );

      const result = await client.getNewsBySlug('research-breakthrough');
      
      expect(result).toEqual(expectedResponse);
      expect(result.news.content).toContain('Breakthrough Research');
      expect(result.news.content).toContain('<ul>');
    });

    it('should handle news articles without content field', async () => {
      const expectedResponse = {
        news: {
          news_id: '124',
          title: 'Brief Update',
          slug: 'brief-update',
          summary: 'Quick news update',
          publishing_status: 'published',
          category_id: '456',
          author_name: 'News Team',
          // content field is null/undefined
          image_url: 'https://storage.azure.com/images/brief.jpg',
          featured: false,
          order_number: 2
        },
        correlation_id: 'brief-correlation-ghi'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(expectedResponse)
      );

      const result = await client.getNewsBySlug('brief-update');
      
      expect(result).toEqual(expectedResponse);
      expect(result.news.content).toBeUndefined();
    });

    it('should throw validation error for empty slug', async () => {
      await expect(client.getNewsBySlug('')).rejects.toThrow('News article slug is required');
    });
  });

  describe('getNewsCategories', () => {
    it('should call correct v1 API endpoint for categories', async () => {
      const mockResponse = {
        categories: [
          {
            category_id: '456',
            name: 'Health Updates',
            slug: 'health-updates',
            description: 'Latest health and wellness news',
            order_number: 1,
            is_default_unassigned: false
          }
        ],
        count: 1,
        correlation_id: 'categories-correlation-jkl'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

      await client.getNewsCategories();

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/news/categories',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });
  });

  describe('getFeaturedNews', () => {
    it('should call correct v1 API endpoint for featured news', async () => {
      const mockResponse = {
        news: [
          {
            news_id: '789',
            title: 'Featured Article',
            publishing_status: 'published',
            featured: true
          }
        ],
        count: 1,
        correlation_id: 'featured-correlation-mno'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

      await client.getFeaturedNews();

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/news/featured',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });

    it('should handle limit parameter correctly', async () => {
      const mockResponse = {
        news: [],
        count: 0,
        correlation_id: 'limited-correlation-pqr'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

      await client.getFeaturedNews(3);

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/news/featured?limit=3',
        expect.any(Object)
      );
    });
  });

  describe('searchNews', () => {
    it('should call correct v1 API endpoint for news search', async () => {
      const mockResponse = {
        news: [],
        count: 0,
        correlation_id: 'search-correlation-stu'
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(mockResponse)
      );

      await client.searchNews({ q: 'medical research', page: 1, pageSize: 5 });

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:7220/api/v1/news/search?q=medical+research&page=1&pageSize=5',
        expect.any(Object)
      );
    });
  });

  describe('error handling', () => {
    it('should handle 404 errors with correlation_id', async () => {
      const errorResponse = {
        error: {
          code: 'NOT_FOUND',
          message: 'News article not found',
          correlation_id: 'error-correlation-404'
        }
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(errorResponse, false, 404)
      );

      await expect(client.getNewsBySlug('nonexistent')).rejects.toThrow(RestError);
      await expect(client.getNewsBySlug('nonexistent')).rejects.toThrow('Not found');
    });

    it('should handle 400 validation errors', async () => {
      const errorResponse = {
        error: {
          code: 'VALIDATION_ERROR',
          message: 'Invalid news parameters',
          correlation_id: 'error-correlation-400'
        }
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(errorResponse, false, 400)
      );

      await expect(client.getNews({ page: -1 })).rejects.toThrow(RestError);
      await expect(client.getNews({ page: -1 })).rejects.toThrow('Bad request');
    });

    it('should handle 429 rate limit errors', async () => {
      const errorResponse = {
        error: {
          code: 'RATE_LIMIT_EXCEEDED',
          message: 'Too many requests',
          correlation_id: 'error-correlation-429'
        }
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(errorResponse, false, 429)
      );

      await expect(client.getNews()).rejects.toThrow(RestError);
      await expect(client.getNews()).rejects.toThrow('Rate limit exceeded');
    });

    it('should handle 500 server errors', async () => {
      const errorResponse = {
        error: {
          code: 'INTERNAL_ERROR',
          message: 'Database connection failed',
          correlation_id: 'error-correlation-500'
        }
      };

      (global.fetch as any).mockResolvedValueOnce(
        createMockFetchResponse(errorResponse, false, 500)
      );

      await expect(client.getNews()).rejects.toThrow(RestError);
      await expect(client.getNews()).rejects.toThrow('Server error');
    });
  });

  describe('timeout and retry behavior', () => {
    it('should timeout after configured duration', async () => {
      // Never resolves
      (global.fetch as any).mockImplementation(() => 
        new Promise(() => {})
      );

      const start = Date.now();
      await expect(client.getNews()).rejects.toThrow('Request timeout');
      const elapsed = Date.now() - start;
      
      expect(elapsed).toBeGreaterThan(4900);
      expect(elapsed).toBeLessThan(5200);
    }, 15000);

    it('should retry on 500 errors', async () => {
      const errorResponse = {
        error: {
          code: 'INTERNAL_ERROR',
          message: 'Temporary server error',
          correlation_id: 'retry-correlation-500'
        }
      };

      // First call fails, second succeeds
      (global.fetch as any)
        .mockResolvedValueOnce(createMockFetchResponse(errorResponse, false, 500))
        .mockResolvedValueOnce(createMockFetchResponse({
          news: [],
          count: 0,
          correlation_id: 'success-after-retry'
        }));

      const result = await client.getNews();
      
      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result.correlation_id).toBe('success-after-retry');
    });
  });
});