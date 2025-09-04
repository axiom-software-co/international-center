// News Domain Types - Updated for REST API responses aligned with database schema

import {
  BaseEntity,
  PaginationParams,
  FilterParams,
  StandardRestResponse,
  SingleRestResponse,
} from '../rest/types';

export interface NewsArticle extends BaseEntity {
  news_id: string;
  title: string;
  summary: string;
  slug: string;
  publishing_status: 'published' | 'draft' | 'archived';
  category_id: string;
  author_name?: string;
  author_email?: string;
  content?: string; // News detailed content stored in PostgreSQL TEXT field
  image_url?: string; // News images stored in Azure Blob Storage
  featured: boolean;
  order_number: number;
  published_at?: string;
  created_on: string;
  created_by?: string;
  modified_on?: string;
  modified_by?: string;
  is_deleted: boolean;
  deleted_on?: string;
  deleted_by?: string;
}

export interface NewsCategory {
  category_id: string;
  name: string;
  slug: string;
  description?: string;
  order_number: number;
  is_default_unassigned: boolean;
  created_on: string;
  created_by?: string;
  modified_on?: string;
  modified_by?: string;
  is_deleted: boolean;
  deleted_on?: string;
  deleted_by?: string;
}

// Standardized response types
export type NewsResponse = StandardRestResponse<NewsArticle>;
export type NewsArticleResponse = SingleRestResponse<NewsArticle>;

export interface GetNewsParams extends PaginationParams, FilterParams {
  category?: string;
  featured?: boolean;
  sortBy?: 'date-asc' | 'date-desc' | 'title';
}

export interface SearchNewsParams extends PaginationParams {
  q: string;
  category?: string;
  sortBy?: string;
}

// Legacy support - will be deprecated
export interface LegacyNewsResponse {
  articles: NewsArticle[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}
