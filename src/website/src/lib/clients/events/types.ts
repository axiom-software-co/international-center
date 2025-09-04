// Events Domain Types - Following standardized API response patterns

import {
  BaseEntity,
  PaginationParams,
  FilterParams,
  StandardResponse,
  SingleResponse,
} from '../shared/types';

export interface EventCategory extends BaseEntity {
  name: string;
  description: string;
  slug: string;
  color: string;
  display_order: number;
  active: boolean;
}

export interface Event extends BaseEntity {
  title: string;
  slug: string;
  excerpt: string;
  content: string;
  featured_image: string;
  event_date: string; // Date in YYYY-MM-DD format
  event_time: string; // Time in HH:MM format
  location: string;
  capacity: number;
  registration_url: string;
  author: string;
  tags: string[];
  status: 'published' | 'draft' | 'archived';
  featured: boolean;
  category: string;
  category_id?: number;
  category_data?: EventCategory;
  meta_title: string;
  meta_description: string;
  published_at: string;
}

// Standardized response types
export type EventsResponse = StandardResponse<Event>;
export type EventResponse = SingleResponse<Event>;
export type EventCategoriesResponse = SingleResponse<EventCategory[]>;

export interface GetEventsParams extends PaginationParams, FilterParams {
  category?: string;
  featured?: boolean;
  sortBy?: 'date-asc' | 'date-desc' | 'published-asc' | 'published-desc' | 'title';
}

export interface SearchEventsParams extends PaginationParams {
  q: string;
  category?: string;
  sortBy?: string;
}

// Legacy support - will be deprecated
export interface LegacyEventsResponse {
  events: Event[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}