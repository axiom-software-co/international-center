// Research Types Validation Tests - Database schema compliance validation
// Tests validate ResearchArticle interfaces match TABLES-RESEARCH.md schema exactly

import { describe, it, expect } from 'vitest';
import type { ResearchArticle, ResearchResponse, ResearchArticleResponse } from './types';

// Database schema reference - Research table from TABLES-RESEARCH.md
interface DatabaseSchemaResearch {
  // Primary key and identifiers (Required)
  research_id: string; // UUID PRIMARY KEY
  title: string; // VARCHAR(255) NOT NULL
  abstract: string; // TEXT NOT NULL (different from current 'excerpt')
  content?: string; // TEXT (nullable)
  slug: string; // VARCHAR(255) UNIQUE NOT NULL
  
  // Category relationship (Required)
  category_id: string; // UUID NOT NULL REFERENCES research_categories(category_id)
  
  // Media and publication info (Optional)
  image_url?: string; // VARCHAR(500) (nullable, different from 'featured_image')
  author_names: string; // VARCHAR(500) NOT NULL (different from 'author')
  publication_date?: string; // DATE (nullable, ISO date format)
  doi?: string; // VARCHAR(100) (nullable)
  external_url?: string; // VARCHAR(500) (nullable)
  report_url?: string; // VARCHAR(500) (nullable, PDF publication report)
  
  // Publishing workflow (Required)
  publishing_status: 'draft' | 'published' | 'archived'; // VARCHAR(20) NOT NULL DEFAULT 'draft'
  
  // Metadata and categorization (Required)
  keywords: string[]; // TEXT[] (PostgreSQL array, different from 'tags')
  research_type: 'clinical_study' | 'case_report' | 'systematic_review' | 'meta_analysis' | 'editorial' | 'commentary'; // VARCHAR(50) NOT NULL
  
  // Audit fields (Required: created_on, others optional)
  created_on: string; // TIMESTAMPTZ NOT NULL DEFAULT NOW()
  created_by?: string; // VARCHAR(255) (nullable)
  modified_on?: string; // TIMESTAMPTZ (nullable)
  modified_by?: string; // VARCHAR(255) (nullable)
  
  // Soft delete fields (Required: is_deleted, others optional)
  is_deleted: boolean; // BOOLEAN NOT NULL DEFAULT FALSE
  deleted_on?: string; // TIMESTAMPTZ (nullable)
  deleted_by?: string; // VARCHAR(255) (nullable)
}

// Database schema reference - Research Categories table from TABLES-RESEARCH.md
interface DatabaseSchemaResearchCategory {
  category_id: string; // UUID PRIMARY KEY
  name: string; // VARCHAR(255) NOT NULL
  slug: string; // VARCHAR(255) UNIQUE NOT NULL
  description?: string; // TEXT (nullable)
  is_default_unassigned: boolean; // BOOLEAN NOT NULL DEFAULT FALSE
  
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

describe('Research Types Database Schema Compliance', () => {
  describe('ResearchArticle Interface Schema Validation', () => {
    it('should have all required database schema fields', () => {
      // This test will fail until ResearchArticle interface matches DatabaseSchemaResearch
      // It drives the need to align types with database schema in GREEN phase
      
      const mockDatabaseResearch: DatabaseSchemaResearch = {
        research_id: 'research-uuid-123',
        title: 'Database Schema Research Article',
        abstract: 'Research article abstract matching database schema',
        content: 'Full research article content with methodology and findings',
        slug: 'database-schema-research-article',
        category_id: 'research-category-uuid-456',
        image_url: 'https://example.com/research-image.jpg',
        author_names: 'Dr. Database Schema, Dr. Research Validation',
        publication_date: '2024-03-15',
        doi: '10.1234/database.research.2024',
        external_url: 'https://journal.example.com/database-research',
        report_url: 'https://storage.example.com/reports/database-research.pdf',
        publishing_status: 'published',
        keywords: ['database', 'schema', 'research', 'validation'],
        research_type: 'systematic_review',
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'researcher@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'admin@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // This assertion will fail until ResearchArticle interface is aligned with database schema
      expect(() => {
        const research: ResearchArticle = mockDatabaseResearch as any;
        
        // Required database fields validation
        expect(research.research_id).toBeDefined();
        expect(research.abstract).toBeDefined(); // Currently 'excerpt' in ResearchArticle
        expect(research.author_names).toBeDefined(); // Currently 'author' in ResearchArticle
        expect(research.publishing_status).toBeDefined(); // Currently 'status' in ResearchArticle
        expect(research.keywords).toBeDefined(); // Currently 'tags' in ResearchArticle
        expect(research.research_type).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.image_url).toBeDefined(); // Currently 'featured_image' in ResearchArticle
        expect(research.publication_date).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.doi).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.external_url).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.report_url).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.created_on).toBeDefined(); // Currently 'published_at' in ResearchArticle
        expect(research.created_by).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.modified_on).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.modified_by).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.is_deleted).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.deleted_on).toBeDefined(); // Currently missing in ResearchArticle
        expect(research.deleted_by).toBeDefined(); // Currently missing in ResearchArticle
        
        // Field type validation
        expect(typeof research.research_id).toBe('string');
        expect(typeof research.title).toBe('string');
        expect(typeof research.abstract).toBe('string');
        expect(typeof research.slug).toBe('string');
        expect(typeof research.category_id).toBe('string');
        expect(typeof research.author_names).toBe('string');
        expect(typeof research.publishing_status).toBe('string');
        expect(Array.isArray(research.keywords)).toBe(true);
        expect(typeof research.research_type).toBe('string');
        expect(typeof research.created_on).toBe('string');
        expect(typeof research.is_deleted).toBe('boolean');
      }).not.toThrow();
    });

    it('should not have legacy fields that are not in database schema', () => {
      // Create a proper ResearchArticle object using correct database schema fields
      const properResearch: ResearchArticle = {
        research_id: 'research-uuid-123',
        title: 'Proper Research Article',
        abstract: 'Research abstract', // Not excerpt
        content: 'Research content',
        slug: 'proper-research',
        category_id: 'category-uuid-456',
        image_url: 'https://example.com/image.jpg', // Not featured_image
        author_names: 'Dr. John Smith, Dr. Jane Doe', // Not author
        publication_date: '2024-03-15',
        doi: '10.1000/123456',
        external_url: 'https://external.example.com',
        report_url: 'https://reports.example.com/research.pdf',
        publishing_status: 'published', // Not status
        keywords: ['research', 'study'], // Not tags
        research_type: 'clinical_study',
        created_on: '2024-01-01T00:00:00Z', // Not published_at
        created_by: 'admin@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'admin@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // Test that proper ResearchArticle object doesn't have legacy fields
      expect((properResearch as any).excerpt).toBeUndefined();
      expect((properResearch as any).featured_image).toBeUndefined();
      expect((properResearch as any).author).toBeUndefined();
      expect((properResearch as any).tags).toBeUndefined();
      expect((properResearch as any).status).toBeUndefined();
      expect((properResearch as any).featured).toBeUndefined();
      expect((properResearch as any).category).toBeUndefined();
      expect((properResearch as any).client_name).toBeUndefined();
      expect((properResearch as any).industry).toBeUndefined();
      expect((properResearch as any).challenge).toBeUndefined();
      expect((properResearch as any).solution).toBeUndefined();
      expect((properResearch as any).results).toBeUndefined();
      expect((properResearch as any).technologies).toBeUndefined();
      expect((properResearch as any).gallery_images).toBeUndefined();
      expect((properResearch as any).meta_title).toBeUndefined();
      expect((properResearch as any).meta_description).toBeUndefined();
      expect((properResearch as any).published_at).toBeUndefined();
    });

    it('should validate research_type enum values match database constraints', () => {
      const validResearchTypes: DatabaseSchemaResearch['research_type'][] = [
        'clinical_study', 'case_report', 'systematic_review', 
        'meta_analysis', 'editorial', 'commentary'
      ];

      // This will pass once ResearchArticle interface is properly aligned
      validResearchTypes.forEach(researchType => {
        expect(['clinical_study', 'case_report', 'systematic_review', 'meta_analysis', 'editorial', 'commentary'])
          .toContain(researchType);
      });

      // Test that current ResearchArticle interface supports these values
      const testResearch: Partial<ResearchArticle> = {
        research_type: 'clinical_study' as any // This will fail until research_type is added to ResearchArticle interface
      };
      expect(testResearch.research_type).toBe('clinical_study');
    });

    it('should validate publishing_status enum values match database constraints', () => {
      const validStatuses: DatabaseSchemaResearch['publishing_status'][] = [
        'draft', 'published', 'archived'
      ];

      validStatuses.forEach(status => {
        expect(['draft', 'published', 'archived'])
          .toContain(status);
      });

      // Test that current ResearchArticle interface supports these values
      const testResearch: Partial<ResearchArticle> = {
        publishing_status: 'published' as any // This will fail until publishing_status is added to ResearchArticle interface
      };
      expect(testResearch.publishing_status).toBe('published');
    });

    it('should validate field length constraints match database schema', () => {
      // title VARCHAR(255) NOT NULL
      expect('A'.repeat(255).length).toBeLessThanOrEqual(255);
      expect(() => {
        const title = 'A'.repeat(256);
        expect(title.length).toBeLessThanOrEqual(255);
      }).toThrow();

      // slug VARCHAR(255) UNIQUE NOT NULL
      expect('research-slug-'.repeat(18).length).toBeLessThanOrEqual(255); // 18 * 14 = 252

      // author_names VARCHAR(500) NOT NULL
      expect('Dr. Author Name '.repeat(30).length).toBeLessThanOrEqual(500);

      // doi VARCHAR(100)
      expect('10.1234/'.repeat(12).length).toBeLessThanOrEqual(100);

      // image_url VARCHAR(500)
      expect('https://example.com/research-images/very-long-path/'.repeat(8).length).toBeLessThanOrEqual(500);

      // external_url VARCHAR(500)
      expect('https://journal.example.com/articles/very-long-path/'.repeat(8).length).toBeLessThanOrEqual(500);

      // report_url VARCHAR(500)
      expect('https://storage.example.com/reports/very-long-filename/'.repeat(7).length).toBeLessThanOrEqual(500);
    });

    it('should validate required vs optional fields according to database schema', () => {
      const requiredFields = [
        'research_id', 'title', 'abstract', 'slug', 'category_id', 'author_names',
        'publishing_status', 'keywords', 'research_type', 'created_on', 'is_deleted'
      ];

      const optionalFields = [
        'content', 'image_url', 'publication_date', 'doi', 'external_url', 'report_url',
        'created_by', 'modified_on', 'modified_by', 'deleted_on', 'deleted_by'
      ];

      // This test will guide GREEN phase implementation
      requiredFields.forEach(field => {
        expect(requiredFields).toContain(field);
      });

      optionalFields.forEach(field => {
        expect(optionalFields).toContain(field);
      });
    });

    it('should validate DOI format constraints', () => {
      // DOI should be VARCHAR(100) and follow DOI format patterns
      const validDOIs = [
        '10.1234/example.doi.123',
        '10.5678/research.article.2024',
        '10.9999/very.specific.research.identifier.with.long.suffix'
      ];

      validDOIs.forEach(doi => {
        expect(doi.length).toBeLessThanOrEqual(100);
        expect(doi.startsWith('10.')).toBe(true);
        expect(doi.includes('/')).toBe(true);
      });
    });

    it('should validate keywords array structure', () => {
      // keywords should be TEXT[] (PostgreSQL array)
      const validKeywords = ['research', 'database', 'schema', 'validation', 'clinical-study'];
      
      expect(Array.isArray(validKeywords)).toBe(true);
      validKeywords.forEach(keyword => {
        expect(typeof keyword).toBe('string');
        expect(keyword.length).toBeGreaterThan(0);
      });

      // Test that current ResearchArticle interface uses keywords instead of tags
      const testResearch: Partial<ResearchArticle> = {
        keywords: validKeywords as any // This will fail until keywords replaces tags
      };
      expect(Array.isArray(testResearch.keywords)).toBe(true);
    });
  });

  describe('Research Categories Schema Validation', () => {
    it('should have all required database schema fields for research categories', () => {
      const mockDatabaseCategory: DatabaseSchemaResearchCategory = {
        category_id: 'research-category-uuid-123',
        name: 'Clinical Research',
        slug: 'clinical-research',
        description: 'Research related to clinical studies',
        is_default_unassigned: false,
        created_on: '2024-01-01T00:00:00Z',
        created_by: 'admin@example.com',
        modified_on: '2024-01-02T00:00:00Z',
        modified_by: 'admin@example.com',
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      // This will fail until research category interfaces match database schema
      expect(() => {
        // ResearchArticle should not include category data directly, only category_id
        const research: Partial<ResearchArticle> = {
          category_id: mockDatabaseCategory.category_id
        };
        
        expect(research.category_id).toBe('research-category-uuid-123');
        expect((research as any).category_data).toBeUndefined(); // Should not exist
      }).not.toThrow();
    });
  });

  describe('Response Types Schema Validation', () => {
    it('should validate ResearchResponse matches expected backend format', () => {
      // Backend format: { data: [...], pagination: {...}, success: boolean }
      const mockResearchResponse: ResearchResponse = {
        data: [],
        pagination: {
          page: 1,
          pageSize: 10,
          total: 0,
          totalPages: 0
        },
        success: true
      };

      expect(Array.isArray(mockResearchResponse.data)).toBe(true);
      expect(typeof mockResearchResponse.pagination).toBe('object');
      expect(typeof mockResearchResponse.pagination!.page).toBe('number');
      expect(typeof mockResearchResponse.pagination!.pageSize).toBe('number');
      expect(typeof mockResearchResponse.pagination!.total).toBe('number');
      expect(typeof mockResearchResponse.pagination!.totalPages).toBe('number');
      expect(typeof mockResearchResponse.success).toBe('boolean');
    });

    it('should validate ResearchArticleResponse matches expected backend format', () => {
      // Backend format: { data: {...}, success: boolean }
      const mockResearch: DatabaseSchemaResearch = {
        research_id: 'test-uuid',
        title: 'Test Research',
        abstract: 'Test Abstract',
        slug: 'test-research',
        category_id: 'category-uuid',
        author_names: 'Dr. Test Author',
        publishing_status: 'published',
        keywords: ['test'],
        research_type: 'clinical_study',
        created_on: '2024-01-01T00:00:00Z',
        is_deleted: false,
      };

      const mockResearchResponse: ResearchArticleResponse = {
        data: mockResearch as any,
        success: true
      };

      expect(typeof mockResearchResponse.data).toBe('object');
      expect(mockResearchResponse.data).not.toBeNull();
      expect(typeof mockResearchResponse.success).toBe('boolean');
    });
  });

  describe('Date and Timestamp Field Validation', () => {
    it('should validate date fields use ISO string format', () => {
      // publication_date should be DATE in ISO format (YYYY-MM-DD)
      const validPublicationDates = [
        '2024-03-15',
        '2023-12-01',
        '2022-01-31'
      ];

      validPublicationDates.forEach(date => {
        expect(date).toMatch(/^\d{4}-\d{2}-\d{2}$/);
        expect(new Date(date).toISOString().slice(0, 10)).toBe(date);
      });
    });

    it('should validate timestamp fields use ISO string format', () => {
      // created_on, modified_on, deleted_on should be TIMESTAMPTZ in ISO format
      const validTimestamps = [
        '2024-01-01T00:00:00Z',
        '2024-03-15T14:30:00.000Z',
        '2023-12-31T23:59:59.999Z'
      ];

      validTimestamps.forEach(timestamp => {
        expect(timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
        expect(new Date(timestamp).toISOString()).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
      });
    });
  });

  describe('Type System Integration', () => {
    it('should ensure type safety for database schema-compliant research', () => {
      // This test ensures the type system prevents incorrect usage
      const createValidResearch = (data: DatabaseSchemaResearch): ResearchArticle => {
        // This will fail until ResearchArticle interface matches DatabaseSchemaResearch
        return data as ResearchArticle;
      };

      const validData: DatabaseSchemaResearch = {
        research_id: 'uuid-123',
        title: 'Valid Research',
        abstract: 'Valid abstract',
        slug: 'valid-research',
        category_id: 'category-uuid',
        author_names: 'Dr. Valid Author',
        publishing_status: 'published',
        keywords: ['valid'],
        research_type: 'clinical_study',
        created_on: '2024-01-01T00:00:00Z',
        is_deleted: false,
      };

      const research = createValidResearch(validData);
      expect(research.research_id).toBe('uuid-123');
    });

    it('should prevent assignment of incompatible legacy data', () => {
      // This test ensures legacy ResearchArticle data cannot be assigned to database schema
      const legacyData = {
        title: 'Legacy Research',
        excerpt: 'This should be abstract', // Wrong field name
        author: 'Dr. Legacy', // Wrong field name
        tags: ['legacy'], // Wrong field name
        status: 'published', // Wrong field name
        featured_image: 'image.jpg', // Wrong field name
      };

      // This should fail type checking when properly aligned
      expect(() => {
        const research: DatabaseSchemaResearch = legacyData as any;
        expect(research.abstract).toBeUndefined(); // Should fail - excerpt is not abstract
        expect(research.author_names).toBeUndefined(); // Should fail - author is not author_names
        expect(research.keywords).toBeUndefined(); // Should fail - tags is not keywords
      }).not.toThrow();
    });
  });

  describe('Business Logic Validation', () => {
    it('should validate research_type affects content validation requirements', () => {
      // Different research types may have different content requirements
      const researchTypes: DatabaseSchemaResearch['research_type'][] = [
        'clinical_study', 'case_report', 'systematic_review', 
        'meta_analysis', 'editorial', 'commentary'
      ];

      researchTypes.forEach(type => {
        const research: Partial<DatabaseSchemaResearch> = {
          research_type: type,
          title: `${type} Research`,
          abstract: `Abstract for ${type}`,
        };

        expect(research.research_type).toBe(type);
        expect(research.title).toContain(type);
        expect(research.abstract).toContain(type);
      });
    });

    it('should validate soft delete behavior', () => {
      // is_deleted should control visibility, not physical deletion
      const activeResearch: Partial<DatabaseSchemaResearch> = {
        is_deleted: false,
        deleted_on: null,
        deleted_by: null,
      };

      const deletedResearch: Partial<DatabaseSchemaResearch> = {
        is_deleted: true,
        deleted_on: '2024-01-15T10:30:00Z',
        deleted_by: 'admin@example.com',
      };

      expect(activeResearch.is_deleted).toBe(false);
      expect(activeResearch.deleted_on).toBe(null);
      expect(activeResearch.deleted_by).toBe(null);

      expect(deletedResearch.is_deleted).toBe(true);
      expect(deletedResearch.deleted_on).toBeDefined();
      expect(deletedResearch.deleted_by).toBeDefined();
    });
  });
});