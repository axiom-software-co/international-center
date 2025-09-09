import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, computed, nextTick } from 'vue';
import { useNews, useNewsArticle, useFeaturedNews, useSearchNews, useNewsCategories } from './useNews';

// Mock the store module - RED phase: define store-centric contracts
vi.mock('../stores/news', () => ({
  useNewsStore: vi.fn()
}));

import { useNewsStore } from '../stores/news';


describe('useNews composables', () => {
  // Define mock store structure - RED phase: store-centric contract
  const mockStore = {
    // State refs that composables should expose via storeToRefs
    news: ref([]),
    article: ref(null), // Individual news article state
    loading: ref(false),
    error: ref(null),
    total: ref(0),
    categories: ref([]),
    featuredNews: ref([]),
    searchResults: ref([]),
    searchTotal: ref(0),
    
    // Computed values
    totalPages: computed(() => Math.ceil(mockStore.total.value / 10) || 0),
    
    // Explicit action methods that composables should delegate to
    fetchNews: vi.fn(),
    fetchNewsArticle: vi.fn(),
    fetchFeaturedNews: vi.fn(),
    fetchNewsCategories: vi.fn(),
    searchNews: vi.fn(),
  };

  beforeEach(() => {
    // Ensure all store properties are properly initialized as refs
    if (!mockStore.article || !mockStore.article.value !== undefined) {
      mockStore.article = ref(null);
    }
    
    // Reset mock store state
    mockStore.news.value = [];
    mockStore.article.value = null;
    mockStore.loading.value = false;
    mockStore.error.value = null;
    mockStore.total.value = 0;
    mockStore.categories.value = [];
    mockStore.featuredNews.value = [];
    mockStore.searchResults.value = [];
    mockStore.searchTotal.value = 0;
    
    // Clear all store action mocks
    mockStore.fetchNews.mockClear();
    mockStore.fetchNewsArticle.mockClear();
    mockStore.fetchFeaturedNews.mockClear();
    mockStore.fetchNewsCategories.mockClear();
    mockStore.searchNews.mockClear();
    
    // Setup store mock return
    vi.mocked(useNewsStore).mockReturnValue(mockStore as any);
  });


  describe('useNews', () => {
    it('should expose store state via storeToRefs and initialize with correct default values', () => {
      const { news, loading, error, total, page, pageSize, totalPages, refetch } = useNews({ enabled: false });

      // RED phase: expect composable to expose store state directly
      expect(news.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      
      // Contract: composable should expose reactive properties and functions
      expect(totalPages).toBeTruthy();
      expect(typeof refetch).toBe('function');
      
      // Contract: composable should use store
      expect(useNewsStore).toHaveBeenCalled();
    });

    it('should delegate to store.fetchNews and expose store state', async () => {
      const mockNews = [
        {
          news_id: '123',
          title: 'Healthcare Innovation News',
          summary: 'Latest healthcare technology advances',
          slug: 'healthcare-innovation-news',
          publishing_status: 'published',
          category_id: '456',
          author_name: 'Dr. Johnson',
          content: '<h2>Revolutionary Healthcare Technology</h2><p>Our healthcare team has developed innovative solutions.</p>',
          image_url: 'https://storage.azure.com/images/healthcare-innovation.jpg'
        }
      ];

      // RED phase: simulate store state after successful fetch
      mockStore.news.value = mockNews;
      mockStore.total.value = 1;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { news, loading, error, total, refetch } = useNews({ 
        page: 1, 
        pageSize: 10,
        immediate: false 
      });

      // Contract: composable should expose store state
      expect(loading.value).toBe(false);
      
      // Clear any previous calls before testing refetch
      mockStore.fetchNews.mockClear();
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation, not direct client calls  
      expect(mockStore.fetchNews).toHaveBeenCalledTimes(1);
      
      // Contract: composable should expose reactive state from store
      expect(news.value).toBeDefined();
      expect(total.value).toBeDefined();
      expect(loading.value).toBeDefined();
      expect(error.value).toBeDefined();
    });

    it('should handle API errors with correlation_id', async () => {
      const errorMessage = 'News not found';

      // RED phase: simulate store error state
      mockStore.news.value = [];
      mockStore.loading.value = false;
      mockStore.error.value = errorMessage;
      mockStore.total.value = 0;

      const { news, loading, error, refetch } = useNews({ immediate: false });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchNews).toHaveBeenCalled();
      
      // Contract: composable should expose store error state
      expect(news.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(errorMessage);
    });

    it('should delegate search parameters to store.fetchNews', async () => {
      const mockNews = [
        {
          news_id: '789',
          title: 'Health Policy Update',
          summary: 'Important health policy changes',
          slug: 'health-policy-update',
          publishing_status: 'published'
        }
      ];

      // RED phase: simulate store state after search
      mockStore.news.value = mockNews;
      mockStore.total.value = 1;

      const { news, refetch } = useNews({ 
        search: 'health',
        immediate: false 
      });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation with search params
      expect(mockStore.fetchNews).toHaveBeenCalledWith({
        search: 'health'
      });
      
      // Contract: expose store state
      expect(news.value).toEqual(mockNews);
    });

    it('should handle category filtering', async () => {
      const mockNews = [
        {
          news_id: '456',
          title: 'Health Alert',
          summary: 'Important health information',
          category_id: 'health-alerts-id'
        }
      ];

      // RED phase: simulate store state after category fetch
      mockStore.news.value = mockNews;
      mockStore.total.value = 1;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { news, refetch } = useNews({ 
        category: 'health-alerts',
        immediate: false 
      });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation with category params
      expect(mockStore.fetchNews).toHaveBeenCalledWith({
        category: 'health-alerts'
      });
      
      // Contract: expose store state
      expect(news.value).toEqual(mockNews);
    });
  });

  describe('useNewsArticle', () => {
    it('should delegate to store.fetchNewsArticle and expose store state', async () => {
      const mockArticle = {
        news_id: '123',
        title: 'Healthcare Policy Update',
        summary: 'Important policy changes',
        slug: 'healthcare-policy-update',
        publishing_status: 'published',
        category_id: '456',
        author_name: 'Policy Team',
        content: '<h2>Healthcare Policy Changes</h2><p>New healthcare policies include comprehensive coverage.</p>'
      };

      // RED phase: simulate store state after individual article fetch
      mockStore.article.value = mockArticle;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { article, loading, error, refetch } = useNewsArticle(ref('healthcare-policy-update'));
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchNewsArticle).toHaveBeenCalledWith('healthcare-policy-update');
      
      // Contract: composable should expose store state
      expect(article.value).toEqual(mockArticle);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should handle null slug gracefully', async () => {
      // RED phase: simulate initial store state
      mockStore.article.value = null;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { article, loading, error, refetch } = useNewsArticle(ref(null));
      
      await refetch();
      await nextTick();

      // Contract: composable should expose null state
      expect(article.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      
      // RED phase: should not call store action for null slug
      expect(mockStore.fetchNewsArticle).not.toHaveBeenCalled();
    });

  });

  describe('useFeaturedNews', () => {
    it('should delegate to store.fetchFeaturedNews and expose store state', async () => {
      const mockFeaturedNews = [
        {
          news_id: '789',
          title: 'Featured Health Alert',
          publishing_status: 'published',
          featured: true
        },
        {
          news_id: '101',
          title: 'Featured Research Update',
          publishing_status: 'published',
          featured: true
        }
      ];

      // RED phase: simulate store state after featured news fetch
      mockStore.featuredNews.value = mockFeaturedNews;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { news, loading, error, refetch } = useFeaturedNews();
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchFeaturedNews).toHaveBeenCalledWith(undefined);
      
      // Contract: composable should expose store state
      expect(news.value).toEqual(mockFeaturedNews);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should delegate limit parameter to store action', async () => {
      const mockLimitedNews = [
        { news_id: '1', title: 'Limited News 1', publishing_status: 'published' }
      ];

      // RED phase: simulate store state after limited featured news fetch
      mockStore.featuredNews.value = mockLimitedNews;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { news, refetch } = useFeaturedNews(5);

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation with limit
      expect(mockStore.fetchFeaturedNews).toHaveBeenCalledWith(5);
      
      // Contract: composable should expose store state
      expect(news.value).toEqual(mockLimitedNews);
    });
  });

  describe('useNewsCategories', () => {
    it('should delegate to store.fetchNewsCategories and expose store state', async () => {
      const mockCategories = [
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
      ];

      // RED phase: simulate store state after categories fetch
      mockStore.categories.value = mockCategories;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { categories, loading, error, refetch } = useNewsCategories();
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchNewsCategories).toHaveBeenCalled();
      
      // Contract: composable should expose store state
      expect(categories.value).toEqual(mockCategories);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should handle empty categories response from store', async () => {
      // RED phase: simulate empty store state
      mockStore.categories.value = [];
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { categories, loading, error, refetch } = useNewsCategories();

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchNewsCategories).toHaveBeenCalled();
      
      // Contract: composable should expose empty store state
      expect(categories.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });
  });

  describe('error handling across composables', () => {
    it('should handle network errors consistently via store state', async () => {
      const errorMessage = 'Network connection failed';
      
      // RED phase: simulate store network error state
      mockStore.error.value = errorMessage;
      mockStore.loading.value = false;
      mockStore.news.value = [];

      const { error, refetch } = useNews({ immediate: false });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchNews).toHaveBeenCalled();
      
      // Contract: composable should expose store error state
      expect(error.value).toBe(errorMessage);
    });

    it('should reset error state on successful refetch via store', async () => {
      const mockNews = [{ news_id: '1', title: 'Test News' }];

      // RED phase: simulate store error state initially
      mockStore.error.value = 'Temporary error';
      mockStore.loading.value = false;
      mockStore.news.value = [];
      mockStore.total.value = 0;

      const { news, error, refetch } = useNews({ immediate: false });

      // First call shows error state
      await refetch();
      await nextTick();
      expect(error.value).toBe('Temporary error');

      // RED phase: simulate store success state after recovery
      mockStore.error.value = null;
      mockStore.news.value = mockNews;
      mockStore.total.value = 1;

      // Second call shows success state
      await refetch();
      await nextTick();
      
      // RED phase: expect store action delegation for both calls
      expect(mockStore.fetchNews).toHaveBeenCalledTimes(2);
      
      // Contract: composable should expose updated store state
      expect(error.value).toBe(null);
      expect(news.value).toEqual(mockNews);
    });
  });
});
