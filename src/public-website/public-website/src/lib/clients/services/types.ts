// Services Domain Types - Updated for REST API responses aligned with database schema

import {
  BaseEntity,
  PaginationParams,
  FilterParams,
  StandardRestResponse,
  SingleRestResponse,
} from '../rest/types';

export interface Service extends BaseEntity {
  service_id: string;
  title: string;
  description: string;
  slug: string;
  publishing_status: 'published' | 'draft' | 'archived';
  category_id: string;
  delivery_mode: 'mobile_service' | 'outpatient_service' | 'inpatient_service';
  content?: string; // Service detailed content stored in PostgreSQL TEXT field
  image_url?: string; // Service images stored in Azure Blob Storage
  order_number: number;
  created_on: string;
  created_by?: string;
  modified_on?: string;
  modified_by?: string;
  is_deleted: boolean;
  deleted_on?: string;
  deleted_by?: string;
}

export interface ServiceCategory {
  category_id: string;
  name: string;
  slug: string;
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
export type ServicesResponse = StandardRestResponse<Service>;
export type ServiceResponse = SingleRestResponse<Service>;

export interface GetServicesParams extends PaginationParams, FilterParams {
  category?: string;
  featured?: boolean;
}

export interface SearchServicesParams extends PaginationParams {
  q: string;
  category?: string;
  sortBy?: string;
}

// Legacy support - will be deprecated
export interface LegacyServicesResponse {
  services: Service[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}
