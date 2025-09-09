import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ref, computed, nextTick } from 'vue';
import { useResearchArticles, useResearchArticle, useFeaturedResearch, useResearchCategories } from './useResearch';

// Mock the store module - RED phase: define store-centric contracts
vi.mock('../stores/research', () => ({
  useResearchStore: vi.fn()
}));

import { useResearchStore } from '../stores/research';


describe('useResearch composables', () => {
  // Define mock store structure - RED phase: store-centric contract
  const mockStore = {
    // State refs that composables should expose via storeToRefs
    research: ref([]),
    article: ref(null), // Individual research article state
    loading: ref(false),
    error: ref(null),
    total: ref(0),
    categories: ref([]),
    featuredResearch: ref([]),
    searchResults: ref([]),
    searchTotal: ref(0),
    
    // Computed values
    totalPages: computed(() => Math.ceil(mockStore.total.value / 10) || 0),
    
    // Explicit action methods that composables should delegate to
    fetchResearch: vi.fn(),
    fetchResearchArticle: vi.fn(),
    fetchFeaturedResearch: vi.fn(),
    fetchResearchCategories: vi.fn(),
    searchResearch: vi.fn(),
  };

  beforeEach(() => {
    // Ensure all store properties are properly initialized as refs
    if (!mockStore.article || !mockStore.article.value !== undefined) {
      mockStore.article = ref(null);
    }
    
    // Reset mock store state
    mockStore.research.value = [];
    mockStore.article.value = null;
    mockStore.loading.value = false;
    mockStore.error.value = null;
    mockStore.total.value = 0;
    mockStore.categories.value = [];
    mockStore.featuredResearch.value = [];
    mockStore.searchResults.value = [];
    mockStore.searchTotal.value = 0;
    
    // Clear all store action mocks
    mockStore.fetchResearch.mockClear();
    mockStore.fetchResearchArticle.mockClear();
    mockStore.fetchFeaturedResearch.mockClear();
    mockStore.fetchResearchCategories.mockClear();
    mockStore.searchResearch.mockClear();
    
    // Setup store mock return
    vi.mocked(useResearchStore).mockReturnValue(mockStore as any);
  });

  describe('useResearchArticles', () => {
    it('should expose store state via storeToRefs and initialize with correct default values', () => {
      const { research, loading, error, total, page, pageSize, totalPages, refetch } = useResearchArticles({ enabled: false });

      // RED phase: expect composable to expose store state directly
      expect(research.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      expect(total.value).toBe(0);
      expect(page.value).toBe(1);
      expect(pageSize.value).toBe(10);
      
      // Contract: composable should expose reactive properties and functions
      expect(totalPages).toBeTruthy();
      expect(typeof refetch).toBe('function');
      
      // Contract: composable should use store
      expect(useResearchStore).toHaveBeenCalled();
    });

    it('should delegate to store.fetchResearch and expose store state', async () => {
      const mockResearch = [
        {
          research_id: '123',
          title: 'Medical Research Study',
          abstract: 'Comprehensive medical research findings',
          slug: 'medical-research-study',
          publishing_status: 'published',
          category_id: '456',
          author_names: 'Dr. Smith, Dr. Johnson',
          publication_date: '2024-01-15',
          doi: '10.1234/research.2024.001',
          research_type: 'clinical_study',
          keywords: ['medical', 'research', 'clinical']
        }
      ];

      // RED phase: simulate store state after successful fetch
      mockStore.research.value = mockResearch;
      mockStore.total.value = 1;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { research, loading, error, total, refetch } = useResearchArticles({ 
        page: 1, 
        pageSize: 10,
        immediate: false 
      });

      // Contract: composable should expose store state
      expect(loading.value).toBe(false);
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation, not direct client calls
      expect(mockStore.fetchResearch).toHaveBeenCalledWith({
        page: 1,
        pageSize: 10
      });
      
      // Contract: composable should expose store state directly
      expect(research.value).toEqual(mockResearch);
      expect(total.value).toBe(1);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should handle API errors with correlation_id', async () => {
      const errorMessage = 'Research not found';

      // RED phase: simulate store error state
      mockStore.research.value = [];
      mockStore.loading.value = false;
      mockStore.error.value = errorMessage;
      mockStore.total.value = 0;

      const { research, loading, error, refetch } = useResearchArticles({ immediate: false });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchResearch).toHaveBeenCalled();
      
      // Contract: composable should expose store error state
      expect(research.value).toEqual([]);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(errorMessage);
    });

    it('should delegate search parameters to store.fetchResearch', async () => {
      const mockResearch = [
        {
          research_id: '789',
          title: 'Clinical Research Update',
          abstract: 'Advanced clinical research findings',
          slug: 'clinical-research-update',
          publishing_status: 'published'
        }
      ];

      // RED phase: simulate store state after search
      mockStore.research.value = mockResearch;
      mockStore.total.value = 1;

      const { research, refetch } = useResearchArticles({ 
        search: 'clinical',
        immediate: false 
      });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation with search params
      expect(mockStore.fetchResearch).toHaveBeenCalledWith({
        search: 'clinical'
      });
      
      // Contract: expose store state
      expect(research.value).toEqual(mockResearch);
    });

    it('should handle category filtering', async () => {
      const mockResearch = [
        {
          research_id: '456',
          title: 'Healthcare Research',
          abstract: 'Important healthcare research',
          category_id: 'healthcare-research-id'
        }
      ];

      // RED phase: simulate store state after category fetch
      mockStore.research.value = mockResearch;
      mockStore.total.value = 1;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { research, refetch } = useResearchArticles({ 
        category: 'healthcare-research',
        immediate: false 
      });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation with category params
      expect(mockStore.fetchResearch).toHaveBeenCalledWith({
        category: 'healthcare-research'
      });
      
      // Contract: expose store state
      expect(research.value).toEqual(mockResearch);
    });
  });

  describe('useResearchArticle', () => {
    it('should delegate to store.fetchResearchArticle and expose store state', async () => {
      const mockArticle = {
        research_id: '123',
        title: 'Healthcare Research Study',
        abstract: 'Comprehensive healthcare research findings',
        slug: 'healthcare-research-study',
        publishing_status: 'published',
        category_id: '456',
        author_names: 'Dr. Research Team',
        publication_date: '2024-01-15',
        doi: '10.1234/research.2024.001',
        research_type: 'clinical_study'
      };

      // RED phase: simulate store state after individual article fetch
      mockStore.article.value = mockArticle;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { article, loading, error, refetch } = useResearchArticle(ref('healthcare-research-study'));
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchResearchArticle).toHaveBeenCalledWith('healthcare-research-study');
      
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

      const { article, loading, error, refetch } = useResearchArticle(ref(null));
      
      await refetch();
      await nextTick();

      // Contract: composable should expose null state
      expect(article.value).toBe(null);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
      
      // RED phase: should not call store action for null slug
      expect(mockStore.fetchResearchArticle).not.toHaveBeenCalled();
    });
  });

  describe('useFeaturedResearch', () => {
    it('should delegate to store.fetchFeaturedResearch and expose store state', async () => {
      const mockFeaturedResearch = [
        {
          research_id: '789',
          title: 'Featured Medical Research',
          publishing_status: 'published',
          featured: true
        },
        {
          research_id: '101',
          title: 'Featured Clinical Study',
          publishing_status: 'published',
          featured: true
        }
      ];

      // RED phase: simulate store state after featured research fetch
      mockStore.featuredResearch.value = mockFeaturedResearch;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { research, loading, error, refetch } = useFeaturedResearch();
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchFeaturedResearch).toHaveBeenCalledWith(undefined);
      
      // Contract: composable should expose store state
      expect(research.value).toEqual(mockFeaturedResearch);
      expect(loading.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should delegate limit parameter to store action', async () => {
      const mockLimitedResearch = [
        { research_id: '1', title: 'Limited Research 1', publishing_status: 'published' }
      ];

      // RED phase: simulate store state after limited featured research fetch
      mockStore.featuredResearch.value = mockLimitedResearch;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { research, refetch } = useFeaturedResearch(5);

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation with limit
      expect(mockStore.fetchFeaturedResearch).toHaveBeenCalledWith(5);
      
      // Contract: composable should expose store state
      expect(research.value).toEqual(mockLimitedResearch);
    });
  });

  describe('useResearchCategories', () => {
    it('should delegate to store.fetchResearchCategories and expose store state', async () => {
      const mockCategories = [
        {
          category_id: '456',
          name: 'Clinical Research',
          slug: 'clinical-research',
          description: 'Clinical research studies',
          order_number: 1,
          is_default_unassigned: false
        },
        {
          category_id: '789',
          name: 'Medical Research',
          slug: 'medical-research',
          description: 'Medical research findings',
          order_number: 2,
          is_default_unassigned: false
        }
      ];

      // RED phase: simulate store state after categories fetch
      mockStore.categories.value = mockCategories;
      mockStore.loading.value = false;
      mockStore.error.value = null;

      const { categories, loading, error, refetch } = useResearchCategories();
      
      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchResearchCategories).toHaveBeenCalled();
      
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

      const { categories, loading, error, refetch } = useResearchCategories();

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchResearchCategories).toHaveBeenCalled();
      
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
      mockStore.research.value = [];

      const { error, refetch } = useResearchArticles({ immediate: false });

      await refetch();
      await nextTick();

      // RED phase: expect store action delegation
      expect(mockStore.fetchResearch).toHaveBeenCalled();
      
      // Contract: composable should expose store error state
      expect(error.value).toBe(errorMessage);
    });

    it('should reset error state on successful refetch via store', async () => {
      const mockResearch = [{ research_id: '1', title: 'Test Research' }];

      // RED phase: simulate store error state initially
      mockStore.error.value = 'Temporary error';
      mockStore.loading.value = false;
      mockStore.research.value = [];
      mockStore.total.value = 0;

      const { research, error, refetch } = useResearchArticles({ immediate: false });

      // First call shows error state
      await refetch();
      await nextTick();
      expect(error.value).toBe('Temporary error');

      // RED phase: simulate store success state after recovery
      mockStore.error.value = null;
      mockStore.research.value = mockResearch;
      mockStore.total.value = 1;

      // Second call shows success state
      await refetch();
      await nextTick();
      
      // RED phase: expect store action delegation for both calls
      expect(mockStore.fetchResearch).toHaveBeenCalledTimes(2);
      
      // Contract: composable should expose updated store state
      expect(error.value).toBe(null);
      expect(research.value).toEqual(mockResearch);
    });
  });
});
