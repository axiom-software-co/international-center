import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, nextTick, defineComponent } from 'vue';
import { mount } from '@vue/test-utils';
import { useNews, useNewsArticle, useFeaturedNews, useNewsCategories } from './useNews';
import { NewsRestClient } from '../rest/NewsRestClient';
import { RestError } from '../rest/BaseRestClient';

// Mock the NewsRestClient with hoisted functions
const {
  mockGetNews,
  mockGetNewsBySlug,
  mockGetFeaturedNews,
  mockGetNewsCategories,
  MockedNewsRestClient
} = vi.hoisted(() => {
  const mockGetNewsFunc = vi.fn();
  const mockGetNewsBySlugFunc = vi.fn();
  const mockGetFeaturedNewsFunc = vi.fn();
  const mockGetNewsCategoriesFunc = vi.fn();
  
  return {
    mockGetNews: mockGetNewsFunc,
    mockGetNewsBySlug: mockGetNewsBySlugFunc,
    mockGetFeaturedNews: mockGetFeaturedNewsFunc,
    mockGetNewsCategories: mockGetNewsCategoriesFunc,
    MockedNewsRestClient: vi.fn().mockImplementation(() => ({
      getNews: mockGetNewsFunc,
      getNewsBySlug: mockGetNewsBySlugFunc,
      getFeaturedNews: mockGetFeaturedNewsFunc,
      getNewsCategories: mockGetNewsCategoriesFunc,
    }))
  };
});

vi.mock('../rest/NewsRestClient', () => ({
  NewsRestClient: MockedNewsRestClient
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

        mockGetNews.mockResolvedValueOnce(mockBackendResponse);

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
      expect(mockGetNews).toHaveBeenCalledWith({
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

        mockGetNews.mockResolvedValueOnce(mockSearchResponse);

      const { articles, refetch } = useNews({ 
        search: 'medical',
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(mockGetNews).toHaveBeenCalledWith({
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

        mockGetNews.mockResolvedValueOnce(mockCategoryResponse);

      const { articles, refetch } = useNews({ 
        category: 'health-alerts',
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(mockGetNews).toHaveBeenCalledWith({
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

        mockGetNews.mockResolvedValueOnce(mockFeaturedResponse);

      const { articles, refetch } = useNews({ 
        featured: true,
        immediate: false 
      });

      await refetch();
      await nextTick();

      expect(mockGetNews).toHaveBeenCalledWith({
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

        mockGetNews.mockRejectedValueOnce(mockError);

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

        mockGetNews.mockRejectedValueOnce(mockRateLimitError);

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

        mockGetNewsBySlug.mockResolvedValueOnce(mockArticleResponse);

      const { article, loading, error, refetch } = useNewsArticle(ref('healthcare-policy-update'));

      // Wait for initial fetch
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(article.value).toEqual(mockArticleResponse.news);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(mockGetNewsBySlug).toHaveBeenCalledWith('healthcare-policy-update');
    });

    it('should handle null slug gracefully', async () => {
      const { article, loading } = useNewsArticle(ref(null));

      await nextTick();

      expect(article.value).toBe(null);
      expect(loading.value).toBe(false);
      
        expect(mockGetNewsBySlug).not.toHaveBeenCalled();
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

        mockGetNewsBySlug
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
      expect(mockGetNewsBySlug).toHaveBeenCalledTimes(2);
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

      mockGetFeaturedNews.mockResolvedValueOnce(mockFeaturedResponse);

      const TestComponent = defineComponent({
        setup() {
          return useFeaturedNews();
        },
        template: '<div></div>'
      });

      const wrapper = mount(TestComponent);
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      const { articles, loading, error } = (wrapper.vm as any);

      expect(articles).toEqual(mockFeaturedResponse.news);
      expect(loading).toBe(false);
      expect(error).toBe(null);
      expect(mockGetFeaturedNews).toHaveBeenCalledWith(undefined);
    });

    it('should handle limit parameter', async () => {
      const mockLimitedResponse = {
        news: [
          { news_id: '1', title: 'News 1', publishing_status: 'published' }
        ],
        count: 1,
        correlation_id: 'limited-correlation-123'
      };

        mockGetFeaturedNews.mockResolvedValueOnce(mockLimitedResponse);

      const { articles } = useFeaturedNews(5);

      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 0));

      expect(mockGetFeaturedNews).toHaveBeenCalledWith(5);
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

        mockGetNewsCategories.mockResolvedValueOnce(mockCategoriesResponse);

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

      expect(categories).toEqual(mockCategoriesResponse.categories);
      expect(loading).toBe(false);
      expect(error).toBe(null);
      expect(mockGetNewsCategories).toHaveBeenCalled();
    });

    it('should handle empty categories response', async () => {
      const mockEmptyResponse = {
        categories: [],
        count: 0,
        correlation_id: 'empty-categories-correlation'
      };

        mockGetNewsCategories.mockResolvedValueOnce(mockEmptyResponse);

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
    });
  });

  describe('error handling across composables', () => {
    it('should handle network errors consistently', async () => {
      const networkError = new Error('Network connection failed');
      
        mockGetNews.mockRejectedValueOnce(networkError);

      const { error, refetch } = useNews({ immediate: false });

      await refetch();
      await nextTick();

      expect(error.value).toBe('Network connection failed');
    });

    it('should handle timeout errors', async () => {
      const timeoutError = new RestError('Request timeout after 5000ms', 408);
      
        mockGetNewsBySlug.mockRejectedValueOnce(timeoutError);

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

        mockGetNews
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