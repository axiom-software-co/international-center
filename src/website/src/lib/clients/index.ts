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
import { NewsRestClient } from './news/NewsRestClient';
export { NewsRestClient };
export type {
  NewsArticle,
  NewsCategory,
  NewsResponse,
  NewsArticleResponse,
  NewsCategoriesResponse,
  GetNewsParams,
  SearchNewsParams,
} from './news/types';

// Create news client instance
const newsClientInstance = new NewsRestClient();
export { newsClientInstance as newsClient };

// Events domain (REST-enabled through Public Gateway)
import { EventsRestClient } from './events/EventsRestClient';
export { EventsRestClient };
export type {
  Event,
  EventCategory,
  EventsResponse,
  EventResponse,
  GetEventsParams,
  SearchEventsParams,
} from './events/types';

// Create events client instance
const eventsClientInstance = new EventsRestClient();
export { eventsClientInstance as eventsClient };

// Research domain (REST-enabled through Public Gateway)
import { ResearchRestClient } from './research/ResearchRestClient';
export { ResearchRestClient };
export type {
  ResearchArticle,
  ResearchResponse,
  ResearchArticleResponse,
  GetResearchParams,
  SearchResearchParams,
} from './research/types';

// Create research client instance
const researchClientInstance = new ResearchRestClient();
export { researchClientInstance as researchClient };

// ====================================================================================
// APIS ON HOLD - These will be migrated to REST architecture when development resumes
// ====================================================================================

// Events domain (ON HOLD - gRPC implementation preserved but unused)  
// export { eventsGrpcClient as eventsClient } from './grpc/clients';

// Contacts domain (ON HOLD - gRPC implementation preserved but unused)
// export { contactsGrpcClient as contactsClient } from './grpc/clients';

// Search domain (ON HOLD - gRPC implementation preserved but unused)
// export { searchGrpcClient as searchClient } from './grpc/clients';

// Newsletter domain (ON HOLD - gRPC implementation preserved but unused)
// export { newsletterGrpcClient as newsletterClient } from './grpc/clients';
