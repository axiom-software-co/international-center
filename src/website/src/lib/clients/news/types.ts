// News Domain Types - Database schema-compliant types matching TABLES-NEWS.md exactly

import {
  BaseEntity,
  PaginationParams,
  FilterParams,
  StandardRestResponse,
  SingleRestResponse,
} from '../rest/types';

export interface NewsArticle extends BaseEntity {
  // Primary key and identifiers
  news_id: string; // UUID PRIMARY KEY
  title: string; // VARCHAR(255) NOT NULL
  summary: string; // TEXT NOT NULL
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

export interface NewsCategory {
  // Primary key and identifiers
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

// Standardized response types following API patterns
export interface NewsResponse {
  news: NewsArticle[]; // Array of news articles
  count: number; // Total count for pagination
  correlation_id: string; // Request correlation ID for tracing
}

export interface NewsArticleResponse {
  news: NewsArticle; // Single news article
  correlation_id: string; // Request correlation ID for tracing
}

export interface NewsCategoriesResponse {
  categories: NewsCategory[]; // Array of news categories
  count: number; // Total count for pagination
  correlation_id: string; // Request correlation ID for tracing
}

// Query parameter interfaces
export interface GetNewsParams extends PaginationParams {
  category?: string; // Filter by category slug
  featured?: boolean; // Filter featured news only
  sortBy?: 'date-asc' | 'date-desc' | 'title' | 'priority'; // Sort options
}

export interface SearchNewsParams extends PaginationParams {
  q: string; // Search query string
  category?: string; // Filter by category slug during search
  sortBy?: 'date-asc' | 'date-desc' | 'relevance'; // Search sort options
}
