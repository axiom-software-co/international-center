/**
 * Common types and interfaces for API clients
 */

export interface ApiError {
  code: string;
  message: string;
  details?: any;
  correlation_id: string;
  timestamp: string;
}

export interface ApiResponse<T> {
  data: T | null;
  error: ApiError | null;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface PaginationInfo {
  current_page: number;
  total_pages: number;
  total_items: number;
  items_per_page: number;
  has_next: boolean;
  has_previous: boolean;
}

export interface RequestOptions extends RequestInit {
  timeout?: number;
  retries?: number;
}

export interface SearchParams extends PaginationParams {
  search?: string;
  sort?: string;
}

export interface FilterParams {
  category_id?: string;
  status?: string;
  type?: string;
  date_from?: string;
  date_to?: string;
  tags?: string[];
  featured?: boolean;
}

// Form submission types
export interface FormSubmissionResponse {
  inquiry_id: string;
  submitted_at: string;
}

// Admin-specific types
export interface AdminCredentials {
  email: string;
  password: string;
  remember_me?: boolean;
}

export interface TokenInfo {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

// Error handling types
export interface ValidationError {
  field: string;
  message: string;
  code: string;
}

export interface ApiValidationError extends ApiError {
  validation_errors: ValidationError[];
}

// Health check types
export interface HealthStatus {
  status: 'healthy' | 'unhealthy';
  timestamp: string;
  version: string;
  checks: {
    database: ServiceCheck;
    dapr: ServiceCheck;
    vault: ServiceCheck;
  };
}

export interface ServiceCheck {
  status: 'up' | 'down';
  response_time_ms: number;
  error?: string;
}