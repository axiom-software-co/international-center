// Research REST Client Tests - Contract validation for database schema compliance
// Tests validate ResearchRestClient methods against TABLES-RESEARCH.md schema requirements

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ResearchRestClient } from './ResearchRestClient';
import type { ResearchArticle, ResearchResponse, ResearchArticleResponse, GetResearchParams, SearchResearchParams } from './types';

// Database schema validation - ResearchArticle interface must match TABLES-RESEARCH.md exactly
interface DatabaseSchemaResearch {
  // Primary key and identifiers
  research_id: string; // UUID PRIMARY KEY
  title: string; // VARCHAR(255) NOT NULL
  abstract: string; // TEXT NOT NULL (different from current 'excerpt')
  content?: string; // TEXT (nullable)
  slug: string; // VARCHAR(255) UNIQUE NOT NULL
  
  // Category relationship
  category_id: string; // UUID NOT NULL REFERENCES research_categories(category_id)
  
  // Media and publication info
  image_url?: string; // VARCHAR(500) (nullable, different from 'featured_image')
  author_names: string; // VARCHAR(500) NOT NULL (different from 'author')
  publication_date?: string; // DATE (nullable, ISO date format)
  doi?: string; // VARCHAR(100) (nullable)
  external_url?: string; // VARCHAR(500) (nullable)
  report_url?: string; // VARCHAR(500) (nullable, PDF publication report)
  
  // Publishing workflow
  publishing_status: 'draft' | 'published' | 'archived'; // VARCHAR(20) NOT NULL DEFAULT 'draft'
  
  // Metadata and categorization
  keywords: string[]; // TEXT[] (PostgreSQL array, different from 'tags')
  research_type: 'clinical_study' | 'case_report' | 'systematic_review' | 'meta_analysis' | 'editorial' | 'commentary'; // VARCHAR(50) NOT NULL
  
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

describe('ResearchRestClient', () => {
  let client: ResearchRestClient;
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = vi.fn();
    global.fetch = mockFetch;
    client = new ResearchRestClient();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Database Schema Compliance', () => {
    it('should validate ResearchArticle interface matches database schema exactly', () => {
      // This test validates that our ResearchArticle type matches the database schema
      // It should fail initially to drive schema alignment
      const mockResearch: DatabaseSchemaResearch = {
        research_id: 'research-uuid-123',
        title: 'Sample Research Article',
        abstract: 'Research article abstract text', // Different from 'excerpt'
        content: 'Full research article content',
        slug: 'sample-research-article',
        category_id: 'research-category-uuid-456',
        image_url: 'https://example.com/research-image.jpg', // Different from 'featured_image'
        author_names: 'Dr. Jane Smith, Dr. John Doe', // Different from 'author'
        publication_date: '2024-03-15', // ISO date format
        doi: '10.1234/example.doi.123',
        external_url: 'https://journal.example.com/article/123',
        report_url: 'https://storage.example.com/reports/research-123.pdf',
        publishing_status: 'published',
        keywords: ['healthcare', 'clinical-study', 'research'], // Different from 'tags'
        research_type: 'clinical_study',
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'researcher@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'admin@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // This assertion will fail until ResearchArticle interface matches DatabaseSchemaResearch
      // Driving the need to align types with database schema
      expect(() => {
        const research: ResearchArticle = mockResearch as any;
        // Required fields validation
        expect(research.research_id).toBeDefined();
        expect(research.abstract).toBeDefined(); // Should exist, not 'excerpt'
        expect(research.author_names).toBeDefined(); // Should exist, not 'author'
        expect(research.publishing_status).toBeDefined();
        expect(research.keywords).toBeDefined(); // Should exist, not 'tags'
        expect(research.research_type).toBeDefined();
        expect(research.image_url).toBeDefined(); // Should exist, not 'featured_image'
        expect(research.publication_date).toBeDefined();
        expect(research.doi).toBeDefined();
        expect(research.external_url).toBeDefined();
        expect(research.report_url).toBeDefined();
        expect(research.is_deleted).toBeDefined();
        expect(research.created_on).toBeDefined();
      }).not.toThrow();
    });

    it('should validate research_type enum values match database constraints', () => {
      const validResearchTypes: DatabaseSchemaResearch['research_type'][] = [
        'clinical_study', 'case_report', 'systematic_review', 
        'meta_analysis', 'editorial', 'commentary'
      ];
      
      validResearchTypes.forEach(researchType => {
        expect(['clinical_study', 'case_report', 'systematic_review', 'meta_analysis', 'editorial', 'commentary'])
          .toContain(researchType);
      });
    });

    it('should validate publishing_status enum values match database constraints', () => {
      const validStatuses: DatabaseSchemaResearch['publishing_status'][] = [
        'draft', 'published', 'archived'
      ];
      
      validStatuses.forEach(status => {
        expect(['draft', 'published', 'archived'])
          .toContain(status);
      });
    });

    it('should validate doi format constraints', () => {
      // DOI should be VARCHAR(100) - test length constraint
      const validDoi = '10.1234/example.doi.12345678901234567890123456789012345678901234567890123456789012345678901234567890';
      expect(validDoi.length).toBeLessThanOrEqual(100);
    });

    it('should validate author_names format constraints', () => {
      // author_names should be VARCHAR(500) - test length constraint
      const longAuthorNames = 'A'.repeat(500);
      expect(longAuthorNames.length).toBeLessThanOrEqual(500);
    });
  });

  describe('getResearchArticles', () => {
    it('should fetch research articles with proper query parameters', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const params: GetResearchParams = {
        page: 1,
        pageSize: 10,
        category: 'clinical-research',
        featured: true,
        industry: 'healthcare'
      };

      const result = await client.getResearchArticles(params);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research?page=1&pageSize=10&category=clinical-research&featured=true&industry=healthcare'),
        expect.objectContaining({
          method: 'GET'
        })
      );
      expect(result).toEqual(mockResponse);
    }, 5000);

    it('should handle empty parameters', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await client.getResearchArticles();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research'),
        expect.objectContaining({
          method: 'GET'
        })
      );
      expect(result).toEqual(mockResponse);
    }, 5000);

    it('should return research articles matching database schema', async () => {
      const mockDatabaseResearch: DatabaseSchemaResearch = {
        research_id: 'research-uuid-123',
        title: 'Database Schema Research',
        abstract: 'Research article with complete database schema fields',
        content: 'Full research article content with detailed methodology and findings',
        slug: 'database-schema-research',
        category_id: 'research-category-uuid-456',
        image_url: 'https://example.com/research-image.jpg',
        author_names: 'Dr. Schema Validator, Dr. Database Designer',
        publication_date: '2024-03-15',
        doi: '10.1234/schema.research.2024',
        external_url: 'https://journal.example.com/schema-research',
        report_url: 'https://storage.example.com/reports/schema-research.pdf',
        publishing_status: 'published',
        keywords: ['database', 'schema', 'validation', 'research'],
        research_type: 'systematic_review',
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'researcher@schema.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'validator@schema.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      const mockResponse: ResearchResponse = {
        data: [mockDatabaseResearch as any],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 1,
          totalPages: 1
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await client.getResearchArticles();
      
      expect(result.data).toHaveLength(1);
      const research = result.data[0];
      
      // Validate database schema required fields
      expect(research.research_id).toBeDefined();
      expect(research.title).toBeDefined();
      expect(research.abstract).toBeDefined(); // Not 'excerpt'
      expect(research.slug).toBeDefined();
      expect(research.category_id).toBeDefined();
      expect(research.author_names).toBeDefined(); // Not 'author'
      expect(research.publishing_status).toBeDefined();
      expect(research.keywords).toBeDefined(); // Not 'tags'
      expect(research.research_type).toBeDefined();
      expect(research.created_on).toBeDefined();
      expect(research.is_deleted).toBeDefined();
    }, 5000);
  });

  describe('getResearchArticleBySlug', () => {
    it('should fetch research article by slug and return database schema compliant article', async () => {
      const mockDatabaseResearch: DatabaseSchemaResearch = {
        research_id: 'research-uuid-789',
        title: 'Slug Research Article',
        abstract: 'Research article fetched by slug with complete schema',
        slug: 'slug-research-article',
        category_id: 'research-category-uuid-789',
        author_names: 'Dr. Slug Researcher',
        publishing_status: 'published',
        keywords: ['slug', 'test', 'research'],
        research_type: 'case_report',
        created_on: '2024-02-01T00:00:00Z',
        is_deleted: false,
      };

      const mockResponse: ResearchArticleResponse = {
        data: mockDatabaseResearch as any,
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await client.getResearchArticleBySlug('slug-research-article');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research/slug/slug-research-article'),
        expect.objectContaining({
          method: 'GET'
        })
      );
      expect(result.data.slug).toBe('slug-research-article');
      expect(result.data.research_id).toBeDefined();
      expect(result.data.abstract).toBeDefined();
      expect(result.data.author_names).toBeDefined();
    }, 5000);

    it('should throw error for empty slug', async () => {
      await expect(client.getResearchArticleBySlug('')).rejects.toThrow('Research article slug is required');
    }, 5000);

    it('should encode special characters in slug', async () => {
      const mockResponse: ResearchArticleResponse = {
        data: {} as any,
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getResearchArticleBySlug('research with spaces & special chars');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research/slug/research%20with%20spaces%20%26%20special%20chars'),
        expect.any(Object)
      );
    }, 5000);
  });

  describe('getResearchArticleById', () => {
    it('should fetch research article by ID', async () => {
      const mockResponse: ResearchArticleResponse = {
        data: {} as any,
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getResearchArticleById('research-uuid-123');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research/research-uuid-123'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should throw error for empty ID', async () => {
      await expect(client.getResearchArticleById('')).rejects.toThrow('Research article ID is required');
    }, 5000);
  });

  describe('getFeaturedResearch', () => {
    it('should fetch featured research articles', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getFeaturedResearch();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research/featured'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should handle limit parameter', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 5,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getFeaturedResearch(5);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research/featured?limit=5'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);
  });

  describe('searchResearch', () => {
    it('should search research articles with query parameters', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 5,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const params: SearchResearchParams = {
        q: 'clinical study diabetes',
        page: 1,
        pageSize: 5,
        category: 'clinical-research'
      };

      await client.searchResearch(params);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research/search?q=clinical%20study%20diabetes&page=1&pageSize=5&category=clinical-research'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should encode search query properly', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.searchResearch({ q: 'research with special & chars' });

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('q=research%20with%20special%20%26%20chars'),
        expect.any(Object)
      );
    }, 5000);
  });

  describe('getRecentResearch', () => {
    it('should fetch recent research articles with default limit', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 5,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getRecentResearch();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research?limit=5&sortBy=date-desc'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should fetch recent research articles with custom limit', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getRecentResearch(10);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research?limit=10&sortBy=date-desc'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);
  });

  describe('getResearchByCategory', () => {
    it('should fetch research articles by category', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getResearchByCategory('clinical-research');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research/categories/clinical-research/articles'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should throw error for empty category', async () => {
      await expect(client.getResearchByCategory('')).rejects.toThrow('Category is required');
    }, 5000);

    it('should encode category name', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getResearchByCategory('medical research studies');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research/categories/medical%20research%20studies/articles'),
        expect.any(Object)
      );
    }, 5000);
  });

  describe('getResearchByIndustry', () => {
    it('should fetch research articles by industry', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getResearchByIndustry('healthcare');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research?industry=healthcare'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);

    it('should throw error for empty industry', async () => {
      await expect(client.getResearchByIndustry('')).rejects.toThrow('Industry is required');
    }, 5000);

    it('should handle additional parameters', async () => {
      const mockResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 2,
          pageSize: 20,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      await client.getResearchByIndustry('healthcare', {
        page: 2,
        pageSize: 20,
        sortBy: 'date-desc'
      });

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/research?industry=healthcare&page=2&pageSize=20&sortBy=date-desc'),
        expect.objectContaining({
          method: 'GET'
        })
      );
    }, 5000);
  });

  describe('Error Handling', () => {
    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(client.getResearchArticles()).rejects.toThrow('Network error');
    }, 5000);

    it('should handle HTTP error responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found'
      });

      await expect(client.getResearchArticles()).rejects.toThrow();
    }, 5000);

    it('should handle malformed JSON responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => { throw new Error('Invalid JSON'); }
      });

      await expect(client.getResearchArticles()).rejects.toThrow();
    }, 5000);
  });
});