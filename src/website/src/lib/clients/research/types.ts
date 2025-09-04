// Research Domain Types - Updated for standardized API responses

import {
  BaseEntity,
  PaginationParams,
  FilterParams,
  StandardResponse,
  SingleResponse,
} from '../shared/types';

export interface ResearchArticle extends BaseEntity {
  title: string;
  slug: string;
  excerpt: string;
  content: string;
  featured_image: string;
  author: string;
  tags: string[];
  status: 'published' | 'draft' | 'archived';
  featured: boolean;
  category: string;
  client_name?: string;
  industry?: string;
  challenge?: string;
  solution?: string;
  results?: string;
  technologies: string[];
  gallery_images: string[];
  meta_title: string;
  meta_description: string;
  published_at: string;
}

// Standardized response types
export type ResearchResponse = StandardResponse<ResearchArticle>;
export type ResearchArticleResponse = SingleResponse<ResearchArticle>;

export interface GetResearchParams extends PaginationParams, FilterParams {
  category?: string;
  featured?: boolean;
  industry?: string;
}

export interface SearchResearchParams extends PaginationParams {
  q: string;
  category?: string;
  sortBy?: string;
}

// Legacy support - will be deprecated
export interface LegacyResearchResponse {
  articles: ResearchArticle[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

