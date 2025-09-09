// Shared types for REST API clients
// Standardized REST response formats for all domain APIs

// Standard REST Response formats
export interface RestPaginationInfo {
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface StandardRestResponse<T> {
  data: T[];
  pagination: RestPaginationInfo;
  success: boolean;
  message?: string;
  errors?: string[];
}

export interface SingleRestResponse<T> {
  data: T;
  success: boolean;
  message?: string;
  errors?: string[];
}

export interface RestError {
  message: string;
  status: number;
  details?: Record<string, any>;
  correlationId?: string;
}

// Common query parameters
export interface PaginationParams {
  page?: number;
  pageSize?: number;
}

export interface FilterParams {
  category?: string;
  featured?: boolean;
  status?: string;
}

export interface SortParams {
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface SearchParams extends PaginationParams {
  q: string;
}

// Base entity interface
export interface BaseEntity {
  id: string;
  createdAt: string;
  updatedAt: string;
}

// API Response Status
export interface ApiStatus {
  healthy: boolean;
  timestamp: string;
  version?: string;
  service: string;
}