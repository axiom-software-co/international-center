// Backend News API Response Types  
// These match the expected response format from the Go backend handlers
// Following the same pattern as services domain

export interface BackendNewsArticle {
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

export interface BackendNewsCategory {
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

export interface BackendFeaturedNews {
  featured_news_id: string;
  news_id: string;
  feature_position: 1 | 2;
  created_on: string;
  created_by?: string;
  modified_on?: string;
  modified_by?: string;
}

// Backend Response Formats
export interface BackendNewsResponse {
  news: BackendNewsArticle[];
  count: number;
  correlation_id: string;
}

export interface BackendSingleNewsResponse {
  news: BackendNewsArticle;
  correlation_id: string;
}

export interface BackendNewsCategoriesResponse {
  categories: BackendNewsCategory[];
  count: number;
  correlation_id: string;
}

export interface BackendFeaturedNewsResponse {
  featured_news: BackendFeaturedNews[];
  count: number;
  correlation_id: string;
}

// Backend Error Response Format (shared across domains)
export interface BackendNewsErrorResponse {
  error: {
    code: string;
    message: string;
    correlation_id: string;
  };
}

// Request Parameters (matching backend expectations)
export interface BackendNewsParams {
  page?: number;
  pageSize?: number;
  search?: string;
  category?: string;
  featured?: boolean;
}

export interface BackendSearchNewsParams {
  q: string;
  page?: number;
  pageSize?: number;
  category?: string;
}