// REST API Clients - All domain clients using standard HTTP/REST
// Clean exports for all REST-enabled domain clients

// Environment configuration
export { config, isLocal, isStaging, isProduction } from '../../environments';
export type { EnvironmentConfig, Environment } from '../../environments';

// Shared REST types
export type {
  RestPaginationInfo,
  StandardRestResponse,
  SingleRestResponse,
  RestError,
  PaginationParams,
  FilterParams,
  SortParams,
  SearchParams,
  BaseEntity,
  ApiStatus,
} from './types';

// Base REST client
export { BaseRestClient, RestError as ClientRestError } from './BaseRestClient';
export type { RestClientConfig, RestResponse, PaginatedRestResponse } from './BaseRestClient';

// Services domain (REST-enabled)
export { ServicesRestClient } from './ServicesRestClient';

// Re-export services types
export type {
  Service,
  ServiceCategory,
  ServicesResponse,
  ServiceResponse,
  GetServicesParams,
  SearchServicesParams,
  LegacyServicesResponse,
} from '../services/types';

// Create singleton instance
export const servicesClient = new ServicesRestClient();