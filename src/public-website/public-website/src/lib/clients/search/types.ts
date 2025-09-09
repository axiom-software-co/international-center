// Search Domain Types - For unified search functionality

import { BaseEntity, PaginationParams, SingleResponse } from '../shared/types';

export interface SearchResult extends BaseEntity {
  title: string;
  excerpt: string;
  content: string;
  content_type: 'service' | 'news' | 'research';
  slug: string;
  category: string;
  rank: number;
}

export interface SearchFacetValue {
  value: string;
  count: number;
}

export interface SearchFacet {
  name: string;
  values: SearchFacetValue[];
}

export interface SearchResponse {
  results: SearchResult[];
  total: number;
  query: string;
  query_time_ms: number;
  facets?: SearchFacet[];
}

export interface SearchParams extends PaginationParams {
  q: string;
  content_type?: string;
  category?: string;
  sortBy?: 'relevance' | 'date-desc' | 'date-asc' | 'title-asc' | 'title-desc';
}

// Unified search response (wraps SearchResponse)
export type UnifiedSearchResponse = SingleResponse<SearchResponse>;

// Index refresh response
export interface SearchIndexRefreshResponse {
  message: string;
  timestamp: string;
}
