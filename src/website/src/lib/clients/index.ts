// Domain Clients - Main exports (Migrated to REST architecture)
// Clean exports for all domain clients and types

// Environment configuration
export { config, isLocal, isStaging, isProduction } from '../../environments';
export type { EnvironmentConfig, Environment } from '../../environments';

// Shared REST types
export type {
  // REST response types
  RestPaginationInfo,
  StandardRestResponse,
  SingleRestResponse,
  RestError,
  // Common parameter types
  PaginationParams,
  FilterParams,
  SortParams,
  SearchParams,
  BaseEntity,
} from './rest/types';

// Services domain (REST-enabled through Public Gateway)
export { servicesClient } from './rest';
export type {
  Service,
  ServicesResponse,
  ServiceResponse,
  GetServicesParams,
  SearchServicesParams,
  ServiceCategory,
} from './rest';

// News domain (REST-enabled through Public Gateway)
export { newsClient } from './rest';
export type {
  NewsArticle,
  NewsCategory,
  NewsResponse,
  NewsArticleResponse,
  GetNewsParams,
  SearchNewsParams,
} from './rest';

// ====================================================================================
// APIS ON HOLD - These will be migrated to REST architecture when development resumes
// ====================================================================================

// Research domain (ON HOLD - gRPC implementation preserved but unused)
// export { researchGrpcClient as researchClient } from './grpc/clients';

// Events domain (ON HOLD - gRPC implementation preserved but unused)  
// export { eventsGrpcClient as eventsClient } from './grpc/clients';

// Contacts domain (ON HOLD - gRPC implementation preserved but unused)
// export { contactsGrpcClient as contactsClient } from './grpc/clients';

// Search domain (ON HOLD - gRPC implementation preserved but unused)
// export { searchGrpcClient as searchClient } from './grpc/clients';

// Newsletter domain (ON HOLD - gRPC implementation preserved but unused)
// export { newsletterGrpcClient as newsletterClient } from './grpc/clients';
