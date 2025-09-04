import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, nextTick } from 'vue';
import { useNews, useNewsArticle, useFeaturedNews, useNewsCategories } from './useNews';
import { NewsRestClient } from '../rest/NewsRestClient';
import { RestError } from '../rest/BaseRestClient';

// Mock the NewsRestClient
vi.mock('../rest/NewsRestClient', () => ({
  NewsRestClient: vi.fn().mockImplementation(() => ({
    getNews: vi.fn(),
    getNewsBySlug: vi.fn(),
    getFeaturedNews: vi.fn(),
    getNewsCategories: vi.fn(),
  }))
}));

describe('useNews composables', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('useNews', () => {
    it('should initialize with correct default values', () => {
      const { articles, loading, error, total, page, pageSize, totalPages } = useNews({ enabled: false });

      expect(articles.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      expect(totalPages.value).toBe(0);
    });

    it('should fetch news with backend response format including content', async () => {
      const mockBackendResponse = {
        news: [
          {
            news_id: '123',
            title: 'Healthcare Innovation',
            summary: 'Latest healthcare technology advances',
            slug: 'healthcare-innovation',
            publishing_status: 'published',
            category_id: '456',
            author_name: 'Dr. Johnson',
            content: '<h2>Revolutionary Healthcare Technology</h2><p>Our healthcare team has developed innovative solutions including:</p><ul><li>AI-powered diagnostics</li><li>Telemedicine platforms</li><li>Electronic health records</li><li>Patient monitoring systems</li></ul>',
            image_url: 'https://storage.azure.com/images/healthcare-innovation.jpg',
            featured: false,
            order_number: 1
          }
        ],
        count: 1,
        correlation_id: 'news-correlation-123'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getNews as any).mockResolvedValueOnce(mockBackendResponse);

      const { articles, loading, error, total, refetch } = useNews({ 
        page: 1, 
        pageSize: 10,
        immediate: false 
      });

      expect(loading.value).toBe(false);
      
      await refetch();
      await nextTick();

      expect(articles.value).toEqual(mockBackendResponse.news);
      expect(total.value).toBe(1);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockClient.getNews).toHaveBeenCalledWith({
        page: 1,
        pageSize: 10
      });
    });

    it('should handle search parameter correctly', async () => {
      const mockSearchResponse = {
        news: [
          {
            news_id: '789',
            title: 'Medical Research Update',
            summary: 'Latest medical research findings',
            slug: 'medical-research-update',
            publishing_status: 'published',
            category_id: '456',
            author_name: 'Research Team'
          }
        ],
        count: 1,
        correlation_id: 'search-correlation-789'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getNews as any).mockResolvedValueOnce(mockSearchResponse);

      const { articles, refetch } = useNews({ 
        search: 'medical',
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(mockClient.getNews).toHaveBeenCalledWith({
        search: 'medical'
      });
      expect(articles.value).toEqual(mockSearchResponse.news);
    });

    it('should handle category filtering', async () => {
      const mockCategoryResponse = {
        news: [
          {
            news_id: '456',
            title: 'Health Alert',
            summary: 'Important health information',
            category_id: 'health-alerts-id'
          }
        ],
        count: 1,
        correlation_id: 'category-correlation-456'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getNews as any).mockResolvedValueOnce(mockCategoryResponse);

      const { articles, refetch } = useNews({ 
        category: 'health-alerts',
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(mockClient.getNews).toHaveBeenCalledWith({
        category: 'health-alerts'
      });
    });

    it('should handle featured filtering', async () => {
      const mockFeaturedResponse = {
        news: [
          {
            news_id: '101',
            title: 'Featured Health News',
            featured: true
          }
        ],
        count: 1,
        correlation_id: 'featured-correlation-101'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getNews as any).mockResolvedValueOnce(mockFeaturedResponse);

      const { articles, refetch } = useNews({ 
        featured: true,
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(mockClient.getNews).toHaveBeenCalledWith({
        featured: true
      });
    });

    it('should handle API errors with correlation_id', async () => {
      const mockError = new RestError(
        'News articles not found',
        404,
        { error: { code: 'NOT_FOUND', message: 'News articles not found' } },
        'error-correlation-404'
      );

      const mockClient = new NewsRestClient();
      (mockClient.getNews as any).mockRejectedValueOnce(mockError);

      const { articles, loading, error, refetch } = useNews({ immediate: false });

      await refetch();
      await nextTick();

      expect(articles.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe('News articles not found');
    });

    it('should handle rate limiting errors', async () => {
      const mockRateLimitError = new RestError(
        'Rate limit exceeded: Too many requests',
        429,
        { error: { code: 'RATE_LIMIT_EXCEEDED', message: 'Too many requests' } },
        'rate-limit-correlation-429'
      );

      const mockClient = new NewsRestClient();
      (mockClient.getNews as any).mockRejectedValueOnce(mockRateLimitError);

      const { error, refetch } = useNews({ immediate: false });

      await refetch();
      await nextTick();

      expect(error.value).toBe('Rate limit exceeded: Too many requests');
    });
  });

  describe('useNewsArticle', () => {
    it('should fetch news article by slug with backend format', async () => {
      const mockArticleResponse = {
        news: {
          news_id: '123',
          title: 'Healthcare Policy Update',
          summary: 'Important policy changes',
          slug: 'healthcare-policy-update',
          publishing_status: 'published',
          category_id: '456',
          author_name: 'Policy Team',
          content: '<h2>Healthcare Policy Changes</h2><p>New healthcare policies include comprehensive coverage for:</p><ul><li>Preventive care services</li><li>Mental health support</li><li>Chronic disease management</li></ul><p>Contact your healthcare provider for more information.</p>'
        },
        correlation_id: 'article-correlation-123'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getNewsBySlug as any).mockResolvedValueOnce(mockArticleResponse);

      const { article, loading, error, refetch } = useNewsArticle(ref('healthcare-policy-update'));

      // Wait for initial fetch
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(article.value).toEqual(mockArticleResponse.news);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockClient.getNewsBySlug).toHaveBeenCalledWith('healthcare-policy-update');
    });

    it('should handle null slug gracefully', async () => {
      const { article, loading } = useNewsArticle(ref(null));

      await nextTick();

      expect(article.value).toBe(null);
      expect(loading.value).toBe(false);
      
      const mockClient = new NewsRestClient();
      expect(mockClient.getNewsBySlug).not.toHaveBeenCalled();
    });

    it('should react to slug changes', async () => {
      const mockArticle1 = {
        news: { news_id: '1', title: 'Article 1', slug: 'article-1' },
        correlation_id: 'correlation-1'
      };
      const mockArticle2 = {
        news: { news_id: '2', title: 'Article 2', slug: 'article-2' },
        correlation_id: 'correlation-2'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getNewsBySlug as any)
        .mockResolvedValueOnce(mockArticle1)
        .mockResolvedValueOnce(mockArticle2);

      const slugRef = ref('article-1');
      const { article } = useNewsArticle(slugRef);

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(article.value).toEqual(mockArticle1.news);

      // Change slug
      slugRef.value = 'article-2';
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(article.value).toEqual(mockArticle2.news);
      expect(mockClient.getNewsBySlug).toHaveBeenCalledTimes(2);
    });
  });

  describe('useFeaturedNews', () => {
    it('should fetch featured news for display', async () => {
      const mockFeaturedResponse = {
        news: [
          {
            news_id: '789',
            title: 'Featured Health Alert',
            publishing_status: 'published',
            featured: true
          },
          {
            news_id: '101',
            title: 'Featured Research',
            publishing_status: 'published',
            featured: true
          }
        ],
        count: 2,
        correlation_id: 'featured-correlation-789'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getFeaturedNews as any).mockResolvedValueOnce(mockFeaturedResponse);

      const { articles, loading, error } = useFeaturedNews();

      // Wait for mount and fetch
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(articles.value).toEqual(mockFeaturedResponse.news);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockClient.getFeaturedNews).toHaveBeenCalledWith(undefined);
    });

    it('should handle limit parameter', async () => {
      const mockLimitedResponse = {
        news: [
          { news_id: '1', title: 'News 1', publishing_status: 'published' }
        ],
        count: 1,
        correlation_id: 'limited-correlation-123'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getFeaturedNews as any).mockResolvedValueOnce(mockLimitedResponse);

      const { articles } = useFeaturedNews(5);

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(mockClient.getFeaturedNews).toHaveBeenCalledWith(5);
    });
  });

  describe('useNewsCategories', () => {
    it('should fetch categories with backend format', async () => {
      const mockCategoriesResponse = {
        categories: [
          {
            category_id: '456',
            name: 'Health Updates',
            slug: 'health-updates',
            description: 'Latest health information',
            order_number: 1,
            is_default_unassigned: false
          },
          {
            category_id: '789',
            name: 'Medical Research',
            slug: 'medical-research',
            description: 'Research findings and studies',
            order_number: 2,
            is_default_unassigned: false
          }
        ],
        count: 2,
        correlation_id: 'categories-correlation-456'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getNewsCategories as any).mockResolvedValueOnce(mockCategoriesResponse);

      const { categories, loading, error } = useNewsCategories();

      // Wait for mount and fetch
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(categories.value).toEqual(mockCategoriesResponse.categories);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockClient.getNewsCategories).toHaveBeenCalled();
    });

    it('should handle empty categories response', async () => {
      const mockEmptyResponse = {
        categories: [],
        count: 0,
        correlation_id: 'empty-categories-correlation'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getNewsCategories as any).mockResolvedValueOnce(mockEmptyResponse);

      const { categories } = useNewsCategories();

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(categories.value).toEqual([]);
    });
  });

  describe('error handling across composables', () => {
    it('should handle network errors consistently', async () => {
      const networkError = new Error('Network connection failed');
      
      const mockClient = new NewsRestClient();
      (mockClient.getNews as any).mockRejectedValueOnce(networkError);

      const { error, refetch } = useNews({ immediate: false });

      await refetch();
      await nextTick();

      expect(error.value).toBe('Network connection failed');
    });

    it('should handle timeout errors', async () => {
      const timeoutError = new RestError('Request timeout after 5000ms', 408);
      
      const mockClient = new NewsRestClient();
      (mockClient.getNewsBySlug as any).mockRejectedValueOnce(timeoutError);

      const { error } = useNewsArticle(ref('test-article'));

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(error.value).toBe('Request timeout after 5000ms');
    });

    it('should reset error state on successful refetch', async () => {
      const mockError = new RestError('Temporary error', 500);
      const mockSuccessResponse = {
        news: [{ news_id: '1', title: 'Test Article' }],
        count: 1,
        correlation_id: 'success-correlation'
      };

      const mockClient = new NewsRestClient();
      (mockClient.getNews as any)
        .mockRejectedValueOnce(mockError)
        .mockResolvedValueOnce(mockSuccessResponse);

      const { error, refetch } = useNews({ immediate: false });

      // First call fails
      await refetch();
      await nextTick();
      expect(error.value).toBe('Temporary error');

      // Second call succeeds
      await refetch();
      await nextTick();
      expect(error.value).toBe(null);
    });
  });
});