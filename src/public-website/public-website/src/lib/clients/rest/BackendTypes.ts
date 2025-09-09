// Backend API Response Types
// These match the exact response format from the Go backend handlers

export interface BackendService {
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

export interface BackendServiceCategory {
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

export interface BackendFeaturedCategory {
  featured_category_id: string;
  category_id: string;
  feature_position: 1 | 2;
  created_on: string;
  created_by?: string;
  modified_on?: string;
  modified_by?: string;
}

// Backend Response Formats
export interface BackendServicesResponse {
  services: BackendService[];
  count: number;
  correlation_id: string;
}

export interface BackendServiceResponse {
  service: BackendService;
  correlation_id: string;
}

export interface BackendServiceCategoriesResponse {
  categories: BackendServiceCategory[];
  count: number;
  correlation_id: string;
}

export interface BackendFeaturedCategoriesResponse {
  featured_categories: BackendFeaturedCategory[];
  count: number;
  correlation_id: string;
}

// Backend Error Response Format
export interface BackendErrorResponse {
  error: {
    code: string;
    message: string;
    correlation_id: string;
  };
}

// Request Parameters (matching backend expectations)
export interface BackendServicesParams {
  page?: number;
  pageSize?: number;
  search?: string;
  category?: string;
}

export interface BackendSearchServicesParams {
  q: string;
  page?: number;
  pageSize?: number;
  category?: string;
}