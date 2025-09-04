// Research Domain Types - Database Schema Compliant
// Aligned with TABLES-RESEARCH.md PostgreSQL schema

import {
  PaginationParams,
  FilterParams,
  StandardRestResponse,
  SingleRestResponse,
} from '../rest/types';

// Enums matching database constraints
export type PublishingStatus = 'draft' | 'published' | 'archived';
export type ResearchType = 'clinical_study' | 'case_report' | 'systematic_review' | 'meta_analysis' | 'editorial' | 'commentary';

// Database schema compliant interfaces
export interface ResearchCategory {
  category_id: string; // UUID
  name: string;
  slug: string;
  description?: string; // TEXT, can be NULL
  is_default_unassigned: boolean;
  
  // Audit fields
  created_on: string; // TIMESTAMPTZ as ISO string
  created_by?: string;
  modified_on?: string; // TIMESTAMPTZ as ISO string
  modified_by?: string;
  
  // Soft delete fields
  is_deleted: boolean;
  deleted_on?: string; // TIMESTAMPTZ as ISO string
  deleted_by?: string;
}

export interface ResearchArticle {
  research_id: string; // UUID primary key
  title: string; // VARCHAR(255) NOT NULL
  abstract: string; // TEXT NOT NULL
  content?: string; // TEXT, can be NULL
  slug: string; // VARCHAR(255) UNIQUE NOT NULL
  category_id: string; // UUID NOT NULL
  image_url?: string; // VARCHAR(500) - Azure Blob Storage URL
  author_names: string; // VARCHAR(500) NOT NULL
  publication_date?: string; // DATE as YYYY-MM-DD
  doi?: string; // VARCHAR(100)
  external_url?: string; // VARCHAR(500)
  report_url?: string; // VARCHAR(500) - PDF in Azure Blob Storage
  publishing_status: PublishingStatus;
  
  // Content metadata
  keywords?: string[]; // TEXT[]
  research_type: ResearchType;
  
  // Audit fields
  created_on: string; // TIMESTAMPTZ NOT NULL as ISO string
  created_by?: string;
  modified_on?: string; // TIMESTAMPTZ as ISO string
  modified_by?: string;
  
  // Soft delete fields
  is_deleted: boolean;
  deleted_on?: string; // TIMESTAMPTZ as ISO string
  deleted_by?: string;
  
  // Joined category data (when included)
  category_data?: ResearchCategory;
}

// Featured research relationship
export interface FeaturedResearch {
  featured_research_id: string; // UUID
  research_id: string; // UUID
  created_on: string; // TIMESTAMPTZ
  created_by?: string;
  modified_on?: string; // TIMESTAMPTZ
  modified_by?: string;
}

// API Response types
export type ResearchResponse = StandardRestResponse<ResearchArticle>;
export type ResearchArticleResponse = SingleRestResponse<ResearchArticle>;
export type ResearchCategoriesResponse = SingleRestResponse<ResearchCategory[]>;
export type FeaturedResearchResponse = SingleRestResponse<FeaturedResearch>;

// Query parameter interfaces
export interface GetResearchParams extends PaginationParams, FilterParams {
  category_id?: string; // UUID filter
  research_type?: ResearchType;
  publishing_status?: PublishingStatus;
  publication_date_from?: string; // YYYY-MM-DD
  publication_date_to?: string; // YYYY-MM-DD
  author_names?: string; // Author name filter
  doi?: string; // DOI filter
  featured?: boolean; // Query for featured research
  sortBy?: 'publication_date_asc' | 'publication_date_desc' | 'created_on_asc' | 'created_on_desc' | 'title_asc' | 'title_desc';
}

export interface SearchResearchParams extends PaginationParams {
  q: string; // Search query
  category_id?: string;
  research_type?: ResearchType;
  publishing_status?: PublishingStatus;
  publication_date_from?: string;
  publication_date_to?: string;
  sortBy?: string;
}

export interface CreateResearchParams {
  title: string;
  abstract: string;
  content?: string;
  slug: string;
  category_id: string;
  image_url?: string;
  author_names: string;
  publication_date?: string;
  doi?: string;
  external_url?: string;
  report_url?: string;
  research_type: ResearchType;
  keywords?: string[];
}

export interface UpdateResearchParams extends Partial<CreateResearchParams> {
  research_id: string;
}

