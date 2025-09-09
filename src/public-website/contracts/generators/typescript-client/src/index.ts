/**
 * Main exports for the International Center TypeScript API clients
 */

// Client wrappers
export { PublicApiClient } from './wrappers/PublicApiClient';
export { AdminApiClient } from './wrappers/AdminApiClient';
export type { PublicApiClientConfig } from './wrappers/PublicApiClient';
export type { AdminApiClientConfig, LoginCredentials, LoginResponse } from './wrappers/AdminApiClient';

// Common types
export type {
  ApiError,
  ApiResponse,
  PaginationParams,
  PaginationInfo,
  RequestOptions,
  SearchParams,
  FilterParams,
  FormSubmissionResponse,
  AdminCredentials,
  TokenInfo,
  ValidationError,
  ApiValidationError,
  HealthStatus,
  ServiceCheck
} from './types/common';

// Utility functions
export {
  handleApiError,
  buildRequestOptions,
  extractValidationErrors,
  isValidationError,
  isAuthError,
  isPermissionError,
  isNotFoundError,
  isServerError,
  formatErrorMessage,
  createRetryFunction,
  buildQueryParams,
  getCorrelationId,
  createRequestLogger,
  createResponseLogger
} from './utils/api-utils';

// Re-export generated types for convenience
export type {
  Service,
  ServiceCategory,
  NewsArticle,
  NewsCategory,
  ResearchPublication,
  ResearchCategory,
  Event,
  EventCategory,
  EventRegistration,
  EventRegistrationRequest,
  MediaInquiryRequest,
  BusinessInquiryRequest,
  DonationInquiryRequest,
  VolunteerInquiryRequest
} from './generated/public';

export type {
  AdminUser,
  CreateAdminUserRequest,
  UpdateAdminUserRequest,
  CreateNewsArticleRequest,
  UpdateNewsArticleRequest,
  CreateNewsCategoryRequest,
  CreateResearchPublicationRequest,
  CreateServiceRequest,
  CreateEventRequest,
  Inquiry,
  DashboardAnalytics,
  SystemSettings,
  UpdateSystemSettingsRequest
} from './generated/admin';

// Create default client instances
export const createPublicApiClient = (config?: Partial<import('./wrappers/PublicApiClient').PublicApiClientConfig>) => {
  return new PublicApiClient(config);
};

export const createAdminApiClient = (config?: Partial<import('./wrappers/AdminApiClient').AdminApiClientConfig>) => {
  return new AdminApiClient(config);
};