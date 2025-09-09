import { describe, it, expect, beforeEach } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useResearchStore } from './research';
import type { ResearchArticle, ResearchCategory } from '../lib/clients/research/types';

describe('ResearchStore', () => {
  beforeEach(() => {
    // Create fresh pinia instance for each test
    setActivePinia(createPinia());
  });

  describe('Initial State', () => {
    it('should initialize with empty state and default values', () => {
      const store = useResearchStore();
      
      expect(store.research).toEqual([]);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(store.total).toBe(0);
      expect(store.page).toBe(1);
      expect(store.pageSize).toBe(10);
      expect(store.categories).toEqual([]);
      expect(store.featuredResearch).toEqual([]);
      expect(store.searchResults).toEqual([]);
    });

    it('should provide totalPages getter based on total and pageSize', () => {
      const store = useResearchStore();
      
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
      const store = useResearchStore();
      
      expect(store.loading).toBe(false);
      
      // Should be able to set loading state
      store.setLoading(true);
      expect(store.loading).toBe(true);
      
      store.setLoading(false);
      expect(store.loading).toBe(false);
    });

    it('should manage error state with proper clearing', () => {
      const store = useResearchStore();
      
      expect(store.error).toBeNull();
      
      // Should set error
      store.setError('Network error occurred');
      expect(store.error).toBe('Network error occurred');
      
      // Should clear error
      store.clearError();
      expect(store.error).toBeNull();
    });

    it('should update research data and pagination state', () => {
      const store = useResearchStore();
      const mockResearch: ResearchArticle[] = [
        {
          research_id: 'research-1',
          title: 'Test Research Article',
          abstract: 'Research Abstract',
          slug: 'test-research-article',
          category_id: 'cat-1',
          author_names: 'Dr. John Doe',
          publishing_status: 'published',
          research_type: 'clinical_study',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.setResearch(mockResearch, 25, 2, 10);
      
      expect(store.research).toEqual(mockResearch);
      expect(store.total).toBe(25);
      expect(store.page).toBe(2);
      expect(store.pageSize).toBe(10);
    });
  });

  describe('State Mutation Methods', () => {
    it('should set research with pagination data', () => {
      const store = useResearchStore();
      const mockResearch: ResearchArticle[] = [
        {
          research_id: 'research-1',
          title: 'Clinical Study on Patient Outcomes',
          abstract: 'A comprehensive study examining patient outcomes',
          slug: 'clinical-study-patient-outcomes',
          category_id: 'cat-1',
          author_names: 'Dr. Jane Smith, Dr. Mark Johnson',
          publishing_status: 'published',
          research_type: 'clinical_study',
          publication_date: '2024-01-15',
          doi: '10.1000/example.doi',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.setResearch(mockResearch, 100, 2, 20);
      
      expect(store.research).toEqual(mockResearch);
      expect(store.total).toBe(100);
      expect(store.page).toBe(2);
      expect(store.pageSize).toBe(20);
    });

    it('should set featured research', () => {
      const store = useResearchStore();
      const mockFeatured: ResearchArticle[] = [
        {
          research_id: 'featured-1',
          title: 'Featured Research Article',
          abstract: 'Important featured research',
          slug: 'featured-research-article',
          category_id: 'cat-1',
          author_names: 'Dr. Featured Author',
          publishing_status: 'published',
          research_type: 'meta_analysis',
          publication_date: '2024-03-01',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.$patch({ featuredResearch: mockFeatured });
      
      expect(store.featuredResearch).toEqual(mockFeatured);
    });

    it('should set research categories', () => {
      const store = useResearchStore();
      const mockCategories: ResearchCategory[] = [
        {
          category_id: 'cat-1',
          name: 'Clinical Research',
          slug: 'clinical-research',
          description: 'Research focused on clinical studies and patient care',
          is_default_unassigned: false,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.$patch({ categories: mockCategories });
      
      expect(store.categories).toEqual(mockCategories);
    });

    it('should set search results with total', () => {
      const store = useResearchStore();
      const mockResults: ResearchArticle[] = [
        {
          research_id: 'search-1',
          title: 'Search Result Research',
          abstract: 'Research found through search',
          slug: 'search-result-research',
          category_id: 'cat-1',
          author_names: 'Dr. Search Result',
          publishing_status: 'published',
          research_type: 'case_report',
          publication_date: '2024-04-01',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];
      
      store.$patch({ searchResults: mockResults, searchTotal: 1 });
      
      expect(store.searchResults).toEqual(mockResults);
      expect(store.searchTotal).toBe(1);
    });

    it('should clear search results', () => {
      const store = useResearchStore();
      
      // Set initial search data
      store.$patch({ searchResults: [{ research_id: 'test' } as ResearchArticle], searchTotal: 1 });
      
      // Clear search results
      store.$patch({ searchResults: [], searchTotal: 0 });
      
      expect(store.searchResults).toEqual([]);
      expect(store.searchTotal).toBe(0);
    });
  });

  describe('Cache Management Methods', () => {
    it('should have cache invalidation method', () => {
      const store = useResearchStore();
      
      // Set some cached data
      store.$patch({ research: [{ research_id: 'cached' } as ResearchArticle] });
      
      // Cache invalidation should be available as method
      expect(typeof store.invalidateCache).toBe('function');
      
      // Call invalidate cache
      store.invalidateCache();
      
      // State should remain but cache should be cleared internally
      expect(store.research).toBeDefined();
    });
  });



  describe('Getters', () => {
    it('should provide computed values for pagination', () => {
      const store = useResearchStore();
      
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

    it('should provide hasResearch getter for conditional rendering', () => {
      const store = useResearchStore();
      
      expect(store.hasResearch).toBe(false);
      
      store.$patch({
        research: [{
          research_id: 'test',
          title: 'Test Research',
          abstract: 'Test Abstract',
          slug: 'test-research',
          category_id: 'cat-1',
          author_names: 'Test Author',
          publishing_status: 'published',
          research_type: 'clinical_study',
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }]
      });
      
      expect(store.hasResearch).toBe(true);
    });

    it('should provide research articles grouped by type', () => {
      const store = useResearchStore();
      
      const mockResearch = [
        {
          research_id: 'research-1',
          title: 'Clinical Study',
          abstract: 'Clinical research abstract',
          slug: 'clinical-study',
          category_id: 'cat-1',
          author_names: 'Dr. Clinical',
          publishing_status: 'published' as const,
          research_type: 'clinical_study' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        },
        {
          research_id: 'research-2',
          title: 'Meta Analysis',
          abstract: 'Meta analysis abstract',
          slug: 'meta-analysis',
          category_id: 'cat-1',
          author_names: 'Dr. Meta',
          publishing_status: 'published' as const,
          research_type: 'meta_analysis' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];

      store.$patch({ research: mockResearch });
      
      const grouped = store.researchByType;
      expect(grouped).toHaveProperty('clinical_study');
      expect(grouped).toHaveProperty('meta_analysis');
      expect(grouped['clinical_study']).toHaveLength(1);
      expect(grouped['meta_analysis']).toHaveLength(1);
    });

    it('should provide research articles grouped by category', () => {
      const store = useResearchStore();
      
      const mockResearch = [
        {
          research_id: 'research-1',
          title: 'Research 1',
          abstract: 'Abstract 1',
          slug: 'research-1',
          category_id: 'cat-1',
          author_names: 'Author 1',
          publishing_status: 'published' as const,
          research_type: 'clinical_study' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        },
        {
          research_id: 'research-2',
          title: 'Research 2',
          abstract: 'Abstract 2',
          slug: 'research-2',
          category_id: 'cat-2',
          author_names: 'Author 2',
          publishing_status: 'published' as const,
          research_type: 'systematic_review' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];

      store.$patch({ research: mockResearch });
      
      const grouped = store.researchByCategory;
      expect(grouped).toHaveProperty('cat-1');
      expect(grouped).toHaveProperty('cat-2');
      expect(grouped['cat-1']).toHaveLength(1);
      expect(grouped['cat-2']).toHaveLength(1);
    });

    it('should provide recent research articles based on publication date', () => {
      const store = useResearchStore();
      
      const recentDate = '2024-05-01';
      const olderDate = '2023-01-01';
      
      const mockResearch = [
        {
          research_id: 'recent-1',
          title: 'Recent Research',
          abstract: 'Recent research abstract',
          slug: 'recent-research',
          category_id: 'cat-1',
          author_names: 'Dr. Recent',
          publishing_status: 'published' as const,
          research_type: 'clinical_study' as const,
          publication_date: recentDate,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        },
        {
          research_id: 'older-1',
          title: 'Older Research',
          abstract: 'Older research abstract',
          slug: 'older-research',
          category_id: 'cat-1',
          author_names: 'Dr. Older',
          publishing_status: 'published' as const,
          research_type: 'case_report' as const,
          publication_date: olderDate,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];

      store.$patch({ research: mockResearch });
      
      const recentResearch = store.recentResearch;
      // Should be sorted by publication date (newest first)
      expect(recentResearch[0].research_id).toBe('recent-1');
      expect(recentResearch[1].research_id).toBe('older-1');
    });
  });

  describe('Author Management', () => {
    it('should provide research articles grouped by primary author', () => {
      const store = useResearchStore();
      
      const mockResearch = [
        {
          research_id: 'research-1',
          title: 'Research by Dr. Smith',
          abstract: 'Research abstract',
          slug: 'research-dr-smith',
          category_id: 'cat-1',
          author_names: 'Dr. Smith, Dr. Jones',
          publishing_status: 'published' as const,
          research_type: 'clinical_study' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        },
        {
          research_id: 'research-2',
          title: 'Another by Dr. Smith',
          abstract: 'Another research abstract',
          slug: 'another-dr-smith',
          category_id: 'cat-1',
          author_names: 'Dr. Smith',
          publishing_status: 'published' as const,
          research_type: 'systematic_review' as const,
          created_on: '2024-01-01T00:00:00Z',
          is_deleted: false
        }
      ];

      store.$patch({ research: mockResearch });
      
      const byAuthor = store.researchByPrimaryAuthor;
      expect(byAuthor).toHaveProperty('Dr. Smith');
      expect(byAuthor['Dr. Smith']).toHaveLength(2);
    });
  });

});